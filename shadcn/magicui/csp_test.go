package magicui

import (
	"html"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
)

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
			for _, class := range strings.Fields(html.UnescapeString(m[1])) {
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
