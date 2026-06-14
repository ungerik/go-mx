//go:build ignore

// Command gen regenerates entity.go from entities.json (a vendored copy of
// the WHATWG named character reference table, https://html.spec.whatwg.org/entities.json).
// Run it with: go generate ./html/entity/
package main

import (
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"strings"
	"unicode"
)

type ent struct{ go_, name, desc string } // go_ = Go ident, name = entity without & and ;

type group struct {
	header string
	ents   []ent
}

var groups = []group{
	{"Markup characters that must be escaped to appear literally.", []ent{
		{"LessThan", "lt", "less-than sign"},
		{"GreaterThan", "gt", "greater-than sign"},
		{"Ampersand", "amp", "ampersand"},
		{"DoubleQuote", "quot", "quotation mark"},
		{"SingleQuote", "apos", "apostrophe"},
	}},
	{"Spaces and invisible formatting.", []ent{
		{"NonBreakingSpace", "nbsp", "non-breaking space (no line break)"},
		{"EnSpace", "ensp", "en space (width of an en)"},
		{"EmSpace", "emsp", "em space (width of an em)"},
		{"ThinSpace", "thinsp", "thin space"},
		{"HairSpace", "hairsp", "hair space (thinnest space)"},
		{"FigureSpace", "numsp", "figure space (width of a digit)"},
		{"PunctuationSpace", "puncsp", "punctuation space"},
		{"ZeroWidthSpace", "ZeroWidthSpace", "zero-width space (allows a line break)"},
		{"ZeroWidthNonJoiner", "zwnj", "zero-width non-joiner"},
		{"ZeroWidthJoiner", "zwj", "zero-width joiner"},
		{"LeftToRightMark", "lrm", "left-to-right mark (bidi control)"},
		{"RightToLeftMark", "rlm", "right-to-left mark (bidi control)"},
		{"SoftHyphen", "shy", "soft hyphen (visible only at a line break)"},
	}},
	{"Currency signs.", []ent{
		{"Cent", "cent", "cent sign"},
		{"Pound", "pound", "pound sign"},
		{"Yen", "yen", "yen sign"},
		{"Euro", "euro", "euro sign"},
		{"Currency", "curren", "generic currency sign"},
		{"Florin", "fnof", "florin sign"},
	}},
	{"Legal and editorial marks.", []ent{
		{"Copyright", "copy", "copyright sign"},
		{"Registered", "reg", "registered-trademark sign"},
		{"Trademark", "trade", "trademark sign"},
		{"SoundRecordingCopyright", "copysr", "sound-recording copyright sign"},
	}},
	{"Punctuation and typography.", []ent{
		{"EmDash", "mdash", "em dash"},
		{"EnDash", "ndash", "en dash"},
		{"HorizontalBar", "horbar", "horizontal bar (quotation dash)"},
		{"Ellipsis", "hellip", "horizontal ellipsis"},
		{"Bullet", "bull", "bullet"},
		{"MiddleDot", "middot", "middle dot"},
		{"SectionSign", "sect", "section sign"},
		{"Paragraph", "para", "pilcrow (paragraph) sign"},
		{"Numero", "numero", "numero sign"},
		{"Dagger", "dagger", "dagger"},
		{"DoubleDagger", "Dagger", "double dagger"},
		{"PerMille", "permil", "per-mille sign"},
		{"Prime", "prime", "prime (feet, minutes)"},
		{"DoublePrime", "Prime", "double prime (inches, seconds)"},
		{"Overline", "oline", "overline (spacing overscore)"},
	}},
	{"Quotation marks and guillemets.", []ent{
		{"LeftDoubleQuote", "ldquo", "left double quotation mark"},
		{"RightDoubleQuote", "rdquo", "right double quotation mark"},
		{"LeftSingleQuote", "lsquo", "left single quotation mark"},
		{"RightSingleQuote", "rsquo", "right single quotation mark (also apostrophe)"},
		{"DoubleLowQuote", "bdquo", "double low-9 quotation mark"},
		{"SingleLowQuote", "sbquo", "single low-9 quotation mark"},
		{"LeftGuillemet", "laquo", "left-pointing double angle quotation mark"},
		{"RightGuillemet", "raquo", "right-pointing double angle quotation mark"},
		{"LeftSingleGuillemet", "lsaquo", "single left-pointing angle quotation mark"},
		{"RightSingleGuillemet", "rsaquo", "single right-pointing angle quotation mark"},
	}},
	{"Superscripts and vulgar fractions.", []ent{
		{"Superscript1", "sup1", "superscript one"},
		{"Superscript2", "sup2", "superscript two"},
		{"Superscript3", "sup3", "superscript three"},
		{"OneHalf", "frac12", "one half"},
		{"OneThird", "frac13", "one third"},
		{"TwoThirds", "frac23", "two thirds"},
		{"OneQuarter", "frac14", "one quarter"},
		{"ThreeQuarters", "frac34", "three quarters"},
		{"OneFifth", "frac15", "one fifth"},
		{"TwoFifths", "frac25", "two fifths"},
		{"ThreeFifths", "frac35", "three fifths"},
		{"FourFifths", "frac45", "four fifths"},
		{"OneSixth", "frac16", "one sixth"},
		{"FiveSixths", "frac56", "five sixths"},
		{"OneEighth", "frac18", "one eighth"},
		{"ThreeEighths", "frac38", "three eighths"},
		{"FiveEighths", "frac58", "five eighths"},
		{"SevenEighths", "frac78", "seven eighths"},
	}},
	{"Mathematical operators.", []ent{
		{"PlusMinus", "plusmn", "plus-minus sign"},
		{"MinusPlus", "mnplus", "minus-or-plus sign"},
		{"Times", "times", "multiplication sign"},
		{"Divide", "divide", "division sign"},
		{"Minus", "minus", "minus sign (typographic, not a hyphen)"},
		{"Asterisk", "lowast", "asterisk operator"},
		{"DotOperator", "sdot", "dot operator"},
		{"RingOperator", "compfn", "ring operator (function composition)"},
		{"SquareRoot", "radic", "square root (radical)"},
		{"Summation", "sum", "n-ary summation"},
		{"Product", "prod", "n-ary product"},
		{"Coproduct", "coprod", "n-ary coproduct"},
		{"Integral", "int", "integral"},
		{"DoubleIntegral", "Int", "double integral"},
		{"TripleIntegral", "tint", "triple integral"},
		{"ContourIntegral", "oint", "contour integral"},
		{"PartialDifferential", "part", "partial differential"},
		{"Nabla", "nabla", "nabla (del / gradient)"},
		{"Infinity", "infin", "infinity"},
		{"Proportional", "prop", "proportional to"},
		{"Angle", "ang", "angle"},
		{"MeasuredAngle", "angmsd", "measured angle"},
		{"Perpendicular", "perp", "perpendicular (up tack)"},
		{"Parallel", "parallel", "parallel to"},
		{"NotParallel", "nparallel", "not parallel to"},
		{"Therefore", "there4", "therefore"},
		{"Because", "because", "because"},
		{"Degree", "deg", "degree sign"},
		{"Micro", "micro", "micro sign"},
		{"FractionSlash", "frasl", "fraction slash"},
	}},
	{"Set theory.", []ent{
		{"EmptySet", "empty", "empty set"},
		{"ElementOf", "isin", "element of"},
		{"NotElementOf", "notin", "not an element of"},
		{"Contains", "ni", "contains as member"},
		{"NotContains", "notni", "does not contain as member"},
		{"Intersection", "cap", "intersection"},
		{"Union", "cup", "union"},
		{"Subset", "sub", "subset of"},
		{"Superset", "sup", "superset of"},
		{"NotSubset", "nsub", "not a subset of"},
		{"SubsetEqual", "sube", "subset of or equal to"},
		{"SupersetEqual", "supe", "superset of or equal to"},
		{"NotSubsetEqual", "nsube", "neither a subset of nor equal to"},
		{"NotSupersetEqual", "nsupe", "neither a superset of nor equal to"},
		{"SetMinus", "setminus", "set minus (difference)"},
		{"SubsetNotEqual", "subne", "subset of, not equal to"},
		{"SupersetNotEqual", "supne", "superset of, not equal to"},
		{"SquareSubset", "sqsub", "square image of (square subset)"},
		{"SquareSuperset", "sqsup", "square original of (square superset)"},
		{"SquareIntersection", "sqcap", "square cap (intersection)"},
		{"SquareUnion", "sqcup", "square cup (union)"},
		{"BigUnion", "xcup", "n-ary union"},
		{"BigIntersection", "xcap", "n-ary intersection"},
	}},
	{"Logic.", []ent{
		{"ForAll", "forall", "for all"},
		{"ThereExists", "exist", "there exists"},
		{"NotExists", "nexist", "there does not exist"},
		{"LogicalAnd", "and", "logical and"},
		{"LogicalOr", "or", "logical or"},
		{"LogicalNot", "not", "not sign (negation)"},
		{"DirectSum", "oplus", "circled plus (direct sum)"},
		{"CircledTimes", "otimes", "circled times (tensor product)"},
		{"CircledDot", "odot", "circled dot operator"},
		{"CircledMinus", "ominus", "circled minus"},
		{"CircledSlash", "osol", "circled division slash"},
	}},
	{"Relations.", []ent{
		{"NotEqual", "ne", "not-equal sign"},
		{"LessOrEqual", "le", "less-than-or-equal sign"},
		{"GreaterOrEqual", "ge", "greater-than-or-equal sign"},
		{"NotLess", "nless", "not less-than"},
		{"NotGreater", "ngtr", "not greater-than"},
		{"NotLessOrEqual", "nle", "neither less-than nor equal to"},
		{"NotGreaterOrEqual", "nge", "neither greater-than nor equal to"},
		{"MuchLess", "ll", "much less-than"},
		{"MuchGreater", "gg", "much greater-than"},
		{"Approximately", "asymp", "approximately-equal sign"},
		{"TildeOperator", "sim", "tilde operator (similar to)"},
		{"AsymptoticallyEqual", "simeq", "asymptotically equal to"},
		{"Congruent", "cong", "approximately equal to (congruent)"},
		{"Identical", "equiv", "identical to (equivalent)"},
		{"NotIdentical", "nequiv", "not identical to"},
		{"NotTilde", "nsim", "not similar to"},
		{"VeryMuchLess", "Ll", "much much less-than"},
		{"VeryMuchGreater", "Gg", "much much greater-than"},
		{"LessTilde", "lsim", "less-than or similar to"},
		{"GreaterTilde", "gsim", "greater-than or similar to"},
	}},
	{"Letterlike and blackboard symbols.", []ent{
		{"Reals", "reals", "set of real numbers"},
		{"Complexes", "complexes", "set of complex numbers"},
		{"Integers", "integers", "set of integers"},
		{"Naturals", "naturals", "set of natural numbers"},
		{"Rationals", "rationals", "set of rational numbers"},
		{"Quaternions", "quaternions", "set of quaternions"},
		{"RealPart", "real", "real part (black-letter R)"},
		{"ImaginaryPart", "image", "imaginary part (black-letter I)"},
		{"Weierstrass", "weierp", "Weierstrass power set / p"},
		{"Aleph", "aleph", "aleph (transfinite cardinal)"},
		{"ReducedPlanck", "planck", "reduced Planck constant (h-bar)"},
		{"ScriptL", "ell", "script small l (e.g. liters)"},
	}},
	{"Greek small letters.", []ent{
		{"Alpha", "alpha", "Greek small letter alpha"},
		{"Beta", "beta", "Greek small letter beta"},
		{"Gamma", "gamma", "Greek small letter gamma"},
		{"Delta", "delta", "Greek small letter delta"},
		{"Epsilon", "epsilon", "Greek small letter epsilon"},
		{"Zeta", "zeta", "Greek small letter zeta"},
		{"Eta", "eta", "Greek small letter eta"},
		{"Theta", "theta", "Greek small letter theta"},
		{"Iota", "iota", "Greek small letter iota"},
		{"Kappa", "kappa", "Greek small letter kappa"},
		{"Lambda", "lambda", "Greek small letter lambda"},
		{"Mu", "mu", "Greek small letter mu"},
		{"Nu", "nu", "Greek small letter nu"},
		{"Xi", "xi", "Greek small letter xi"},
		{"Omicron", "omicron", "Greek small letter omicron"},
		{"Pi", "pi", "Greek small letter pi"},
		{"Rho", "rho", "Greek small letter rho"},
		{"Sigma", "sigma", "Greek small letter sigma"},
		{"FinalSigma", "sigmaf", "Greek small letter final sigma"},
		{"Tau", "tau", "Greek small letter tau"},
		{"Upsilon", "upsilon", "Greek small letter upsilon"},
		{"Phi", "phi", "Greek small letter phi"},
		{"Chi", "chi", "Greek small letter chi"},
		{"Psi", "psi", "Greek small letter psi"},
		{"Omega", "omega", "Greek small letter omega"},
		{"ThetaSymbol", "thetasym", "Greek theta symbol"},
		{"PhiSymbol", "phiv", "Greek phi symbol"},
		{"PiSymbol", "piv", "Greek pi symbol"},
		{"EpsilonSymbol", "epsiv", "Greek lunate epsilon symbol"},
		{"VarRho", "rhov", "Greek rho symbol"},
		{"VarKappa", "varkappa", "Greek kappa symbol"},
	}},
	{"Greek capital letters (those distinct from Latin).", []ent{
		{"CapGamma", "Gamma", "Greek capital letter gamma"},
		{"CapDelta", "Delta", "Greek capital letter delta"},
		{"CapTheta", "Theta", "Greek capital letter theta"},
		{"CapLambda", "Lambda", "Greek capital letter lambda"},
		{"CapXi", "Xi", "Greek capital letter xi"},
		{"CapPi", "Pi", "Greek capital letter pi"},
		{"CapSigma", "Sigma", "Greek capital letter sigma"},
		{"CapUpsilon", "Upsilon", "Greek capital letter upsilon"},
		{"CapPhi", "Phi", "Greek capital letter phi"},
		{"CapPsi", "Psi", "Greek capital letter psi"},
		{"CapOmega", "Omega", "Greek capital letter omega"},
	}},
	{"Arrows.", []ent{
		{"LeftArrow", "larr", "leftwards arrow"},
		{"RightArrow", "rarr", "rightwards arrow"},
		{"UpArrow", "uarr", "upwards arrow"},
		{"DownArrow", "darr", "downwards arrow"},
		{"LeftRightArrow", "harr", "left-right arrow"},
		{"UpDownArrow", "varr", "up-down arrow"},
		{"UpLeftArrow", "nwarr", "north-west arrow"},
		{"UpRightArrow", "nearr", "north-east arrow"},
		{"DownLeftArrow", "swarr", "south-west arrow"},
		{"DownRightArrow", "searr", "south-east arrow"},
		{"DoubleLeftArrow", "lArr", "leftwards double arrow"},
		{"DoubleRightArrow", "rArr", "rightwards double arrow (implies)"},
		{"DoubleUpArrow", "uArr", "upwards double arrow"},
		{"DoubleDownArrow", "dArr", "downwards double arrow"},
		{"DoubleLeftRightArrow", "hArr", "left-right double arrow (iff)"},
		{"MapsTo", "mapsto", "rightwards arrow from bar (maps to)"},
		{"HookLeftArrow", "larrhk", "leftwards arrow with hook"},
		{"HookRightArrow", "rarrhk", "rightwards arrow with hook"},
		{"CarriageReturn", "crarr", "downwards arrow with corner leftwards"},
		{"NotRightArrow", "nrarr", "rightwards arrow with stroke (not)"},
		{"LongLeftArrow", "xlarr", "long leftwards arrow"},
		{"LongRightArrow", "xrarr", "long rightwards arrow"},
		{"LongLeftRightArrow", "xharr", "long left-right arrow"},
		{"LeftHarpoonUp", "lharu", "leftwards harpoon with barb up"},
		{"RightHarpoonUp", "rharu", "rightwards harpoon with barb up"},
		{"LeftHarpoonDown", "lhard", "leftwards harpoon with barb down"},
		{"RightHarpoonDown", "rhard", "rightwards harpoon with barb down"},
		{"Equilibrium", "rlhar", "rightwards harpoon over leftwards harpoon"},
		{"ReverseEquilibrium", "lrhar", "leftwards harpoon over rightwards harpoon"},
	}},
	{"Geometric shapes.", []ent{
		{"BlackStar", "starf", "black star"},
		{"WhiteStar", "star", "white star"},
		{"Lozenge", "loz", "lozenge"},
		{"WhiteSquare", "square", "white square"},
		{"BlackSquare", "squf", "black small square"},
		{"WhiteCircle", "cir", "white circle"},
		{"LargeCircle", "bigcirc", "large circle"},
		{"UpTriangle", "utri", "white up-pointing triangle"},
		{"BlackUpTriangle", "utrif", "black up-pointing triangle"},
		{"DownTriangle", "dtri", "white down-pointing triangle"},
		{"BlackDownTriangle", "dtrif", "black down-pointing triangle"},
		{"RightTriangle", "rtri", "white right-pointing triangle"},
		{"BlackRightTriangle", "rtrif", "black right-pointing triangle"},
		{"LeftTriangle", "ltri", "white left-pointing triangle"},
		{"BlackLeftTriangle", "ltrif", "black left-pointing triangle"},
		{"BlackLozenge", "lozf", "black lozenge"},
		{"Rectangle", "rect", "white rectangle"},
	}},
	{"Cards, music, and other symbols.", []ent{
		{"Heart", "hearts", "heart suit"},
		{"Diamond", "diams", "diamond suit"},
		{"Club", "clubs", "club suit"},
		{"Spade", "spades", "spade suit"},
		{"CheckMark", "checkmark", "check mark"},
		{"BallotX", "cross", "ballot x (cross mark)"},
		{"MusicNote", "sung", "eighth note"},
		{"Flat", "flat", "music flat sign"},
		{"Natural", "natur", "music natural sign"},
		{"Sharp", "sharp", "music sharp sign"},
		{"Male", "male", "male sign"},
		{"Female", "female", "female sign"},
		{"Telephone", "phone", "telephone sign"},
	}},
	{"Spacing diacritical marks.", []ent{
		{"Acute", "acute", "acute accent"},
		{"Circumflex", "circ", "modifier circumflex accent"},
		{"Tilde", "tilde", "small tilde"},
		{"Macron", "macr", "macron"},
		{"Diaeresis", "uml", "diaeresis (umlaut)"},
		{"Cedilla", "cedil", "cedilla"},
		{"Breve", "breve", "breve"},
		{"Caron", "caron", "caron (háček)"},
		{"RingAbove", "ring", "ring above"},
		{"DoubleAcute", "dblac", "double acute accent"},
		{"Ogonek", "ogon", "ogonek"},
	}},
}

func visible(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if unicode.IsSpace(r) || unicode.IsControl(r) ||
			unicode.Is(unicode.Cf, r) || unicode.Is(unicode.Mn, r) {
			return false
		}
	}
	return true
}

func main() {
	data, err := os.ReadFile("entities.json")
	if err != nil {
		panic(err)
	}
	var table map[string]struct {
		Characters string `json:"characters"`
	}
	if err := json.Unmarshal(data, &table); err != nil {
		panic(err)
	}

	var b strings.Builder
	b.WriteString(`// Code generated by gen.go from entities.json; DO NOT EDIT.

package entity

import "github.com/ungerik/go-mx"

// HTML named character references. Each renders the named entity shown,
// which the browser displays as the corresponding character.
const (
`)

	seenGo := map[string]bool{}
	var missing, dupGo []string
	count := 0
	for gi, g := range groups {
		if gi > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "\t// %s\n\n", g.header)
		for _, e := range g.ents {
			key := "&" + e.name + ";"
			rec, ok := table[key]
			if !ok {
				missing = append(missing, e.go_+" -> "+key)
				continue
			}
			if seenGo[e.go_] {
				dupGo = append(dupGo, e.go_)
			}
			seenGo[e.go_] = true
			comment := e.desc
			if visible(rec.Characters) {
				comment = rec.Characters + " " + e.desc
			}
			fmt.Fprintf(&b, "\t%s mx.Raw = `%s` // %s\n", e.go_, key, comment)
			count++
		}
	}
	b.WriteString(")\n")

	if len(missing) > 0 {
		fmt.Fprintln(os.Stderr, "MISSING ENTITIES:")
		for _, m := range missing {
			fmt.Fprintln(os.Stderr, "  "+m)
		}
	}
	if len(dupGo) > 0 {
		fmt.Fprintln(os.Stderr, "DUPLICATE GO NAMES:", dupGo)
	}
	fmt.Fprintf(os.Stderr, "GENERATED %d constants (missing %d)\n", count, len(missing))

	if len(missing) > 0 || len(dupGo) > 0 {
		os.Exit(1) // don't write a broken file
	}
	src, err := format.Source([]byte(b.String()))
	if err != nil {
		fmt.Fprintln(os.Stderr, "format error:", err)
		os.Exit(1)
	}
	if err := os.WriteFile("entity.go", src, 0o644); err != nil {
		panic(err)
	}
}
