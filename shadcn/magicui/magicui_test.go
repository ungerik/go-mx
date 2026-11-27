package magicui

import (
	"context"
	stdhtml "html"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// render renders a component with a CheckedWriter and fails the test on any
// render error, which includes a duplicate attribute on one element.
func render(t *testing.T, c mx.Component) string {
	t.Helper()
	var b strings.Builder
	if err := c.Render(t.Context(), mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render error: %v\npartial output:\n%s", err, b.String())
	}
	return b.String()
}

// renderErr renders a component and returns its render error.
func renderErr(t *testing.T, c mx.Component) error {
	t.Helper()
	var b strings.Builder
	return c.Render(t.Context(), mx.NewCheckedWriter(&b))
}

// mustContain fails the test for every want missing from out.
func mustContain(t *testing.T, out string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

// tailwindAnimations are the animate-* utilities Tailwind v4 defines itself.
var tailwindAnimations = []string{"animate-none", "animate-spin", "animate-ping", "animate-pulse", "animate-bounce"}

var (
	themeAnimationRE = regexp.MustCompile(`--(animate-[a-z0-9-]+)\s*:`)
	utilityRE        = regexp.MustCompile(`@utility\s+(animate-[a-z0-9-]+)`)
	keyframesRE      = regexp.MustCompile(`@keyframes\s+([a-z0-9-]+)`)
	sourceAnimateRE  = regexp.MustCompile(`(?:^|[\s"':\[(])(animate-[a-z0-9-]+)`)
)

// definedAnimations returns the animate-* utilities CSS defines, through
// @theme --animate-* entries or @utility rules.
func definedAnimations() []string {
	var names []string
	for _, m := range themeAnimationRE.FindAllStringSubmatch(CSS, -1) {
		names = append(names, m[1])
	}
	for _, m := range utilityRE.FindAllStringSubmatch(CSS, -1) {
		names = append(names, m[1])
	}
	return names
}

// reducedMotionCSS returns the bodies of all prefers-reduced-motion blocks.
func reducedMotionCSS(t *testing.T) string {
	t.Helper()
	var b strings.Builder
	rest := CSS
	for {
		i := strings.Index(rest, "@media (prefers-reduced-motion: reduce)")
		if i < 0 {
			return b.String()
		}
		rest = rest[i:]
		open := strings.Index(rest, "{")
		depth, end := 0, -1
		for j := open; j < len(rest); j++ {
			switch rest[j] {
			case '{':
				depth++
			case '}':
				depth--
			}
			if depth == 0 {
				end = j
				break
			}
		}
		if end < 0 {
			t.Fatal("unbalanced braces in a prefers-reduced-motion block")
		}
		b.WriteString(rest[open : end+1])
		rest = rest[end+1:]
	}
}

func TestCSSHeader(t *testing.T) {
	mustContain(t, CSS, UpstreamCommit, "MIT License", "@theme inline")
}

// TestCSSReducedMotion checks that every animation the package defines is
// switched off under prefers-reduced-motion.
func TestCSSReducedMotion(t *testing.T) {
	reduced := reducedMotionCSS(t)
	for _, name := range definedAnimations() {
		if !regexp.MustCompile(`\.` + regexp.QuoteMeta(name) + `\b`).MatchString(reduced) {
			t.Errorf("animation %s is not switched off in a prefers-reduced-motion block", name)
		}
	}
}

// TestCSSKeyframesUsed checks that every @keyframes rule is referenced by an
// animation, so no keyframes are dead.
func TestCSSKeyframesUsed(t *testing.T) {
	for _, m := range keyframesRE.FindAllStringSubmatch(CSS, -1) {
		name := m[1]
		uses := regexp.MustCompile(`(?:animation(?:-name)?\s*:|--animate-[a-z0-9-]+\s*:)[^;]*\b` + regexp.QuoteMeta(name) + `\b`)
		if !uses.MatchString(CSS) {
			t.Errorf("@keyframes %s is not used by any animation", name)
		}
	}
}

// TestSourceAnimationsDefined checks that every animate-* class in the
// package's Go source is defined in CSS or is one of Tailwind's own.
func TestSourceAnimationsDefined(t *testing.T) {
	defined := append(definedAnimations(), tailwindAnimations...)
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range sourceAnimateRE.FindAllStringSubmatch(string(src), -1) {
			if !slices.Contains(defined, m[1]) {
				t.Errorf("%s uses %s, which neither CSS nor Tailwind defines", file, m[1])
			}
		}
	}
}

// TestCSSNoTailwindOverride checks that the package does not redefine one
// of Tailwind's own animations, which shadcn components rely on.
func TestCSSNoTailwindOverride(t *testing.T) {
	for _, name := range definedAnimations() {
		if slices.Contains(tailwindAnimations, name) {
			t.Errorf("CSS redefines Tailwind's %s", name)
		}
	}
	for _, m := range keyframesRE.FindAllStringSubmatch(CSS, -1) {
		if slices.Contains([]string{"spin", "ping", "pulse", "bounce"}, m[1]) {
			t.Errorf("CSS redefines Tailwind's @keyframes %s", m[1])
		}
	}
}

func TestCheckCSSValue(t *testing.T) {
	for _, valid := range []string{"#ffaa40", "rgba(0, 0, 0, 1)", "0.05em", "var(--primary)", "calc(100% - 2px)", "oklch(0.7 0.1 200)", "linear-gradient(red, rgb(0 0 255))", "[x]"} {
		if err := checkCSSValue("x", valid); err != nil {
			t.Errorf("checkCSSValue(%q): %v", valid, err)
		}
	}
	for _, invalid := range []string{"red;background:blue", "red}", "{", "url(https://example.com/x.png)", "URL(x)", `"x"`, "a\\b", "<x>", "#fff/*", "*/red",
		// An unclosed or stray parenthesis or bracket would pull the
		// declarations after it into the value, or end a function early.
		"rgb(1 2 3", "red), linear-gradient(red,red", "a[", "a]", "(]", "([)]"} {
		if err := checkCSSValue("x", invalid); err == nil {
			t.Errorf("checkCSSValue(%q) accepted an unsafe value", invalid)
		}
	}
}

func TestFinishMergesStyle(t *testing.T) {
	e := finish(html.Div(html.Style("color: red;"), html.Class("p-4")), "x", "p-2 flex", "--a:1")
	out := render(t, e)
	mustContain(t, out, `style="--a:1;color: red"`, `data-slot="x"`, `class="flex p-4"`)
}

// styleFromContext is a style attribute whose value comes from the render
// context, to check that finish values the caller's style at render time.
type styleFromContext struct{}

type styleContextKey struct{}

func (styleFromContext) AttribName() string { return "style" }

func (styleFromContext) AttribValue(ctx context.Context) (string, error) {
	color, _ := ctx.Value(styleContextKey{}).(string)
	return "color:" + color, nil
}

func TestFinishStyleUsesRenderContext(t *testing.T) {
	e := finish(html.Div(styleFromContext{}), "x", "", "--a:1")
	ctx := context.WithValue(t.Context(), styleContextKey{}, "blue")
	var b strings.Builder
	if err := e.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
		t.Fatal(err)
	}
	mustContain(t, b.String(), `style="--a:1;color:blue"`)
}

func TestSeconds(t *testing.T) {
	if got := seconds(0, 1500*1e6); got != "1.5s" {
		t.Errorf("seconds default = %q", got)
	}
	if got := seconds(2e9, 1e9); got != "2s" {
		t.Errorf("seconds = %q", got)
	}
}

// defaultComponents renders every component of the package with its option
// defaults, plus only what it cannot render without: text, words, an ID for
// the components that reference one.
var defaultComponents = map[string]mx.Component{
	"Android":                     Android(AndroidOptions{ID: "a"}),
	"AnimatedCircularProgressBar": AnimatedCircularProgressBar(AnimatedCircularProgressBarOptions{Value: 50}),
	"AnimatedGradientText":        AnimatedGradientText(AnimatedGradientTextOptions{}, "x"),
	"AnimatedList":                AnimatedList(AnimatedListOptions{}, "a", "b"),
	"AnimatedShinyText":           AnimatedShinyText(AnimatedShinyTextOptions{}, "x"),
	"AuroraText":                  AuroraText(AuroraTextOptions{}, "x"),
	"AvatarCircles":               AvatarCircles(AvatarCirclesOptions{}),
	"Backlight":                   Backlight(BacklightOptions{ID: "b"}, "x"),
	"BlurFade":                    BlurFade(BlurFadeOptions{}, "x"),
	"BorderBeam":                  BorderBeam(BorderBeamOptions{}),
	"ComicText":                   ComicText(ComicTextOptions{}, "x"),
	"DotPattern":                  DotPattern(DotPatternOptions{ID: "d"}),
	"GlareHover":                  GlareHover(GlareHoverOptions{}, "x"),
	"GridPattern":                 GridPattern(GridPatternOptions{ID: "g"}),
	"HeroVideoDialog":             HeroVideoDialog(HeroVideoDialogOptions{ID: "h", VideoSrc: "https://example.com/v", ThumbnailSrc: "/t.png"}),
	"HexagonPattern":              HexagonPattern(HexagonPatternOptions{ID: "h"}),
	"InteractiveHoverButton":      InteractiveHoverButton("x"),
	"Iphone":                      Iphone(IphoneOptions{ID: "i", Src: "/s.png"}),
	"KineticText":                 KineticText(KineticTextOptions{Text: "ab"}),
	"LightRays":                   LightRays(LightRaysOptions{}),
	"LineShadowText":              LineShadowText(LineShadowTextOptions{}, "x"),
	"Marquee":                     Marquee(MarqueeOptions{}, "x"),
	"Meteors":                     Meteors(MeteorsOptions{}),
	"NoiseTexture":                NoiseTexture(NoiseTextureOptions{ID: "n"}),
	"OrbitingCircles":             OrbitingCircles(OrbitingCirclesOptions{}, "a", "b"),
	"ProgressiveBlur":             ProgressiveBlur(ProgressiveBlurOptions{}),
	"RainbowButton":               RainbowButton(RainbowButtonOptions{}, "x"),
	"Ripple":                      Ripple(RippleOptions{}),
	"Safari":                      Safari(SafariOptions{ID: "s", ImageSrc: "/s.png"}),
	"ScrollProgress":              ScrollProgress(),
	"ShimmerButton":               ShimmerButton(ShimmerButtonOptions{}, "x"),
	"ShineBorder":                 ShineBorder(ShineBorderOptions{}),
	"ShinyButton":                 ShinyButton("x"),
	"SparklesText":                SparklesText(SparklesTextOptions{}, "x"),
	"SpinningText":                SpinningText(SpinningTextOptions{}, "ab"),
	"StripedPattern":              StripedPattern(StripedPatternOptions{ID: "s"}),
	"Text3DFlip":                  Text3DFlip(Text3DFlipOptions{}, "ab"),
	"TextAnimate":                 TextAnimate(TextAnimateOptions{}, "a b"),
	"TextReveal":                  TextReveal("a b"),
	"WarpBackground":              WarpBackground(WarpBackgroundOptions{}, "x"),
	"WordRotate":                  WordRotate(WordRotateOptions{Words: []string{"a", "b"}}),
}

// needStyleAttribute are the components that still write a style attribute
// with their defaults, because a value differs from element to element: an
// index, a position drawn from Seed, a percentage, an id in url(). The README
// lists them under "Strict Content-Security-Policy"; every other component
// renders its defaults with classes alone.
var needStyleAttribute = []string{
	// An index per item, segment or letter (--i, --index, a stagger delay).
	"AnimatedList", "Ripple", "SpinningText", "Text3DFlip", "TextAnimate", "TextReveal", "WordRotate",
	// Positions and timings drawn from Seed.
	"LightRays", "Meteors", "SparklesText", "WarpBackground",
	// Values computed from the options: an angle per orbiting item, the
	// gauge's percentage, the mask stops of each blur layer.
	"OrbitingCircles", "AnimatedCircularProgressBar", "ProgressiveBlur",
	// filter:url(#id) with the id of its own SVG filter.
	"Backlight",
}

// TestDefaultsWithoutStyleAttribute: under a Content-Security-Policy without
// 'unsafe-inline' in style-src, the browser ignores every style attribute, so
// a component left at its defaults has to carry them in classes.
func TestDefaultsWithoutStyleAttribute(t *testing.T) {
	for name, c := range defaultComponents {
		out := render(t, c)
		has := strings.Contains(out, " style=")
		switch need := slices.Contains(needStyleAttribute, name); {
		case has && !need:
			t.Errorf("%s writes a style attribute with its defaults:\n%s", name, out)
		case !has && need:
			t.Errorf("%s no longer needs a style attribute; remove it from needStyleAttribute and the README", name)
		}
	}
}

// arbitraryProperty returns the property name of a Tailwind arbitrary
// property class such as hover:[mask-image:none], and false for any other
// class. Variants are split off at the colons outside brackets, so an
// arbitrary variant like [&_svg:not(x)]:size-4 is not taken for a property.
func arbitraryProperty(class string) (string, bool) {
	depth, last := 0, 0
	for i, r := range class {
		switch r {
		case '[', '(':
			depth++
		case ']', ')':
			depth--
		case ':':
			if depth == 0 {
				last = i + 1
			}
		}
	}
	utility := class[last:]
	if !strings.HasPrefix(utility, "[") || !strings.HasSuffix(utility, "]") {
		return "", false
	}
	name, _, found := strings.Cut(utility[1:], ":")
	return name, found
}

// TestArbitraryPropertiesAreCSS: a class like [clipPath:inset(0)], a React
// style key carried over from upstream, compiles to nothing in Tailwind, and
// nothing renders wrong enough to notice. Every arbitrary property the
// components use is a CSS property name: lowercase, or a custom property.
func TestArbitraryPropertiesAreCSS(t *testing.T) {
	classRE := regexp.MustCompile(`class="([^"]*)"`)
	propertyRE := regexp.MustCompile(`^(--[a-z0-9-]+|-?[a-z]+(-[a-z]+)*)$`)
	for name, c := range defaultComponents {
		for _, m := range classRE.FindAllStringSubmatch(render(t, c), -1) {
			for _, class := range strings.Fields(stdhtml.UnescapeString(m[1])) {
				if property, ok := arbitraryProperty(class); ok && !propertyRE.MatchString(property) {
					t.Errorf("%s: %s is not a CSS property in %s", name, property, class)
				}
			}
		}
	}
	if property, ok := arbitraryProperty("[clipPath:inset(0)]"); !ok || propertyRE.MatchString(property) {
		t.Error("the check misses a camelCase property")
	}
}
