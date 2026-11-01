package mx

import "testing"

func TestCheckTextLength(t *testing.T) {
	tag := ParseFormTagString("minlength=2,maxlength=4")
	if tag.MinLength != 2 || tag.MaxLength != 4 {
		t.Fatalf("parsed %+v", tag)
	}
	for _, tt := range []struct {
		raw  string
		want string
	}{
		{"", ""},
		{"ab", ""},
		{"abcd", ""},
		{"a", "must be at least 2 characters"},
		{"abcde", "must be at most 4 characters"},
		// The browser counts UTF-16 code units: an emoji is two.
		{"a😀b", ""},
		{"😀😀😀", "must be at most 4 characters"},
	} {
		err := CheckTextLength(tag, tt.raw)
		got := ""
		if err != nil {
			got = err.Error()
			if _, ok := UserMessage(err); !ok {
				t.Errorf("%q: the error is meant for the user", tt.raw)
			}
		}
		if got != tt.want {
			t.Errorf("%q: got %q, want %q", tt.raw, got, tt.want)
		}
	}
	if CheckTextLength(FormTag{}, "anything") != nil {
		t.Error("no bound, no error")
	}
	other := ParseFormTagString("autocomplete=one-time-code,inputmode=numeric,maxlength=x")
	if other.Autocomplete != "one-time-code" || other.InputMode != "numeric" || other.MaxLength != 0 {
		t.Errorf("parsed %+v", other)
	}
}
