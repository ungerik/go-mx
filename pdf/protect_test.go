package pdf

import (
	"crypto/rc4"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestProtectEncryptsEveryStringOnItsOwn checks that every string of a
// protected document decrypts with a new RC4 keystream from its object key,
// the way a reader decrypts it, and not only the first string of an object:
// the info dictionary holds several.
func TestProtectEncryptsEveryStringOnItsOwn(t *testing.T) {
	r := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	r.SetProtection(0, "", "owner")
	r.SetTitle("The title", false)
	r.SetAuthor("The author", false)
	r.SetSubject("The subject", false)
	r.AddPage()
	data := string(renderBytes(t, r))

	header, n := objectHeader(t, data, "/Producer")
	for key, want := range map[string]string{
		"/Title":   "The title",
		"/Author":  "The author",
		"/Subject": "The subject",
	} {
		encrypted, ok := literalStringAfter(data, header, key+" (")
		if !ok {
			t.Errorf("no %s string in the info dictionary", key)
			continue
		}
		if got := decryptString(t, r, n, encrypted); got != want {
			t.Errorf("decrypted %s = %q, want %q", key, got, want)
		}
	}
}

// objectHeader returns the start of the object whose dictionary begins with
// firstKey, and its object number.
func objectHeader(t *testing.T, data, firstKey string) (string, uint32) {
	t.Helper()
	m := regexp.MustCompile(`(\d+) 0 obj\n<<\n` + regexp.QuoteMeta(firstKey)).FindStringSubmatch(data)
	if m == nil {
		t.Fatalf("no object starting with %s in the output", firstKey)
	}
	n, err := strconv.ParseUint(m[1], 10, 32)
	if err != nil {
		t.Fatal(err)
	}
	return m[0], uint32(n)
}

// literalStringAfter returns the unescaped bytes of the PDF literal string
// that starts with open, searched after the first occurrence of after.
func literalStringAfter(data, after, open string) ([]byte, bool) {
	from := strings.Index(data, after)
	if from < 0 {
		return nil, false
	}
	start := strings.Index(data[from:], open)
	if start < 0 {
		return nil, false
	}
	var str []byte
	for i := from + start + len(open); i < len(data) && data[i] != ')'; i++ {
		c := data[i]
		if c == '\\' && i+1 < len(data) {
			i++
			c = data[i]
			if c == 'r' {
				c = '\r'
			}
		}
		str = append(str, c)
	}
	return str, true
}

// decryptString decrypts a string of object n like a reader does: with a new
// RC4 keystream from the object key.
func decryptString(t *testing.T, r *Renderer, n uint32, encrypted []byte) string {
	t.Helper()
	cipher, err := rc4.NewCipher(r.protect.objectKey(n))
	if err != nil {
		t.Fatal(err)
	}
	decrypted := make([]byte, len(encrypted))
	cipher.XORKeyStream(decrypted, encrypted)
	return string(decrypted)
}
