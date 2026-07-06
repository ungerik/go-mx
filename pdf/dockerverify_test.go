package pdf

// The tests in this file verify PDF output with poppler-utils running in a
// docker container — an independent reader implementation, so a check cannot
// share a bug with the writer under test. They skip when docker is not
// available or in -short mode; the first run pulls the poppler image.

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

// popplerImage is the docker image providing pdfinfo, pdfdetach & co.
// minidocks/poppler publishes no version tags besides latest, so the
// assertions below stay tolerant of poppler version differences.
const popplerImage = "minidocks/poppler:latest"

var dockerAvailable = sync.OnceValue(func() bool {
	return exec.Command("docker", "info").Run() == nil
})

func skipWithoutDocker(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping docker verification test in short mode")
	}
	if !dockerAvailable() {
		t.Skip("docker is not available")
	}
}

// dockerTempDir returns a temp directory with symlinks resolved, because
// docker on macOS shares /private/var but not the /var symlink to it that
// t.TempDir returns.
func dockerTempDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

// popplerTool runs a poppler command in docker with dir mounted at /work and
// returns its combined output.
func popplerTool(t *testing.T, dir string, args ...string) string {
	t.Helper()
	dockerArgs := append([]string{"run", "--rm", "-v", dir + ":/work", popplerImage}, args...)
	out, err := exec.Command("docker", dockerArgs...).CombinedOutput()
	if err != nil {
		t.Fatalf("docker %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

// writeTestPDF renders doc into dir with fixed dates so output is deterministic.
func writeTestPDF(t *testing.T, dir, name string, doc *Document, fixed time.Time) {
	t.Helper()
	r := doc.NewRenderer()
	r.SetCreationDate(fixed)
	r.SetModificationDate(fixed)
	if err := doc.Render(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if err := r.OutputFileAndClose(filepath.Join(dir, name)); err != nil {
		t.Fatal(err)
	}
}

// matchOutput asserts that every pattern matches out, reporting the full
// output once on failure.
func matchOutput(t *testing.T, out string, patterns ...string) {
	t.Helper()
	failed := false
	for _, pattern := range patterns {
		if !regexp.MustCompile("(?m)" + pattern).MatchString(out) {
			t.Errorf("output does not match %q", pattern)
			failed = true
		}
	}
	if failed {
		t.Logf("full output:\n%s", out)
	}
}

// An embedded file must survive the round trip through a foreign reader:
// pdfdetach has to see exactly one attachment under its declared name and
// extract the original bytes unchanged, or the "hybrid" part of a
// ZUGFeRD/Factur-X invoice would be unreadable in practice.
func TestDockerPoppler_embeddedFileRoundTrip(t *testing.T) {
	skipWithoutDocker(t)
	fixed := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	doc, invoiceXML := facturXTestDocument(fixed)
	dir := dockerTempDir(t)
	writeTestPDF(t, dir, "facturx.pdf", doc, fixed)

	list := popplerTool(t, dir, "pdfdetach", "-list", "/work/facturx.pdf")
	matchOutput(t, list,
		`^1 embedded files$`,
		`^1: factur-x\.xml$`,
	)

	popplerTool(t, dir, "pdfdetach", "-save", "1", "-o", "/work/extracted.xml", "/work/facturx.pdf")
	extracted, err := os.ReadFile(filepath.Join(dir, "extracted.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(extracted, invoiceXML) {
		t.Errorf("extracted attachment differs from the embedded bytes:\n%s", extracted)
	}
}

// The XMP metadata stream must be found and exposed by a foreign reader with
// the PDF/A identification and the Factur-X properties intact, and the info
// dictionary values pdfinfo reports must agree with it.
func TestDockerPoppler_xmpMetadata(t *testing.T) {
	skipWithoutDocker(t)
	fixed := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	doc, _ := facturXTestDocument(fixed)
	dir := dockerTempDir(t)
	writeTestPDF(t, dir, "facturx.pdf", doc, fixed)

	meta := popplerTool(t, dir, "pdfinfo", "-meta", "/work/facturx.pdf")
	matchOutput(t, meta,
		`<pdfaid:part>3</pdfaid:part>`,
		`<pdfaid:conformance>B</pdfaid:conformance>`,
		`<fx:DocumentType>INVOICE</fx:DocumentType>`,
		`<fx:DocumentFileName>factur-x\.xml</fx:DocumentFileName>`,
		`<fx:ConformanceLevel>EN 16931</fx:ConformanceLevel>`,
		`<xmp:CreateDate>2024-05-06T07:08:09\+00:00</xmp:CreateDate>`,
	)

	// the info dictionary stays consistent with the XMP values
	info := popplerTool(t, dir, "pdfinfo", "/work/facturx.pdf")
	matchOutput(t, info,
		`^Metadata Stream:\s+yes$`,
		`^PDF version:\s+1\.7$`,
		`^Title:\s+Invoice R2024-001$`,
		`^Author:\s+ACME GmbH$`,
		`^Creator:\s+go-mx/pdf test$`,
		`^CreationDate:\s+Mon May\s+6 07:08:09 2024`,
	)
}

// Basic document properties — page count, page size and orientation, PDF
// version, and the info dictionary metadata — must be reported back
// unchanged by a foreign reader.
func TestDockerPoppler_documentInfo(t *testing.T) {
	skipWithoutDocker(t)
	fixed := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	doc := NewDocument("Layout Test",
		Paragraph("page one"),
		Page(), Paragraph("page two"),
		Page(), Paragraph("page three"),
	)
	doc.Author = "ACME GmbH"
	doc.Subject = "Subject line"
	doc.Keywords = "alpha, beta"
	doc.Creator = "go-mx/pdf"
	dir := dockerTempDir(t)
	writeTestPDF(t, dir, "info.pdf", doc, fixed)

	out := popplerTool(t, dir, "pdfinfo", "/work/info.pdf")
	matchOutput(t, out,
		`^Pages:\s+3$`,
		`^Page size:\s+595\.28 x 841\.89 pts \(A4\)$`,
		`^Page rot:\s+0$`,
		`^PDF version:\s+1\.3$`,
		`^Title:\s+Layout Test$`,
		`^Subject:\s+Subject line$`,
		`^Keywords:\s+alpha, beta$`,
		`^Author:\s+ACME GmbH$`,
		`^Creator:\s+go-mx/pdf$`,
		`^Producer:\s+FPDF 1\.7$`,
		`^CreationDate:\s+Mon May\s+6 07:08:09 2024`,
		`^ModDate:\s+Mon May\s+6 07:08:09 2024`,
		`^Encrypted:\s+no$`,
	)
}

// Landscape orientation and non-default page sizes must arrive in the
// document as actual page dimensions, not just as writer-side state.
func TestDockerPoppler_pageSizeAndOrientation(t *testing.T) {
	skipWithoutDocker(t)
	fixed := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	doc := NewDocument("Landscape", Paragraph("wide"))
	doc.Orientation = OrientationLandscape
	doc.PageSize = PageSizeLetter
	dir := dockerTempDir(t)
	writeTestPDF(t, dir, "landscape.pdf", doc, fixed)

	out := popplerTool(t, dir, "pdfinfo", "/work/landscape.pdf")
	matchOutput(t, out,
		`^Pages:\s+1$`,
		`^Page size:\s+792 x 612 pts \(letter\)$`,
		`^Page rot:\s+0$`,
	)
}
