package pdf

// The variables in this file are the package-level configuration of the pdf
// package, mirroring the configuration vars of the mx package. Each has a
// working default and is consulted while a document tree is built or rendered,
// so assigning a different value changes that behavior for the whole program.
// They are plain package variables with no locking: set them once during
// initialization (before any concurrent rendering), not while rendering.
var (
	// AsComponent converts a value passed as a child into a Component. It
	// defaults to DefaultAsComponent — see its docs for the recognized types and
	// the github.com/domonda/go-pretty fallback used to draw an unexpected value
	// as text. Document, page, state and component constructors call AsComponent
	// indirectly (via AsComponents), so assigning a different func changes child
	// conversion everywhere.
	//
	// For example, to turn the silent "unexpected value becomes text" fallback
	// into a hard failure during development (widen the accepted set to taste):
	//
	//	base := pdf.AsComponent
	//	pdf.AsComponent = func(c any) pdf.Component {
	//		switch c.(type) {
	//		case nil, pdf.Component, string:
	//			return base(c)
	//		default:
	//			panic(fmt.Sprintf("pdf: unexpected child of type %T", c))
	//		}
	//	}
	AsComponent = DefaultAsComponent
)
