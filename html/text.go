package html

import "github.com/ungerik/go-mx"

type (
	Text     = mx.Text
	Raw      = mx.Raw
	RawBytes = mx.RawBytes
)

const (
	LessThan         Raw = `&lt;`
	GreaterThan      Raw = `&gt;`
	NonBreakingSpace Raw = `&nbsp;`
	Ampersand        Raw = `&amp;`
	DoubleQuote      Raw = `&quot;`
	SingleQuote      Raw = `&apos;`
	Cent             Raw = `&cent;`
	Pound            Raw = `&pound;`
	Yen              Raw = `&yen;`
	Euro             Raw = `&euro;`
	Copyright        Raw = `&copy;`
	Trademark        Raw = `&reg;`
)
