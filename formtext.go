package mx

import (
	"strconv"
	"unicode/utf16"
)

// CheckTextLength checks raw, the posted text of a text, textarea or
// catch-all field, against the tag's form:"minlength=…,maxlength=…", as
// the browser does with the input's attributes, so a scripted POST is held
// to them too. The length is counted in UTF-16 code units, the browser's
// unit. An empty text is not checked; Required rejects it. The error is a
// [FormMessage] naming the bound, never the text.
func CheckTextLength(tag FormTag, raw string) error {
	if raw == "" || tag.MinLength <= 0 && tag.MaxLength <= 0 {
		return nil
	}
	n := 0
	for _, r := range raw {
		n += utf16.RuneLen(r)
	}
	switch {
	case tag.MaxLength > 0 && n > tag.MaxLength:
		return FormMessage("must be at most " + strconv.Itoa(tag.MaxLength) + " characters")
	case tag.MinLength > 0 && n < tag.MinLength:
		return FormMessage("must be at least " + strconv.Itoa(tag.MinLength) + " characters")
	}
	return nil
}
