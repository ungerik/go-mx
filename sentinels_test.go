package mx

import "testing"

func TestSentinelRoundTrip(t *testing.T) {
	cases := []FieldPath{
		"Name",
		"Branding-PrimaryColor",
		"Address-Street",
		"",
	}
	for _, p := range cases {
		t.Run(string(p), func(t *testing.T) {
			pname := PresentSentinelName(p)
			cname := ClearSentinelName(p)
			if got, ok := ParsePresentSentinel(pname); !ok || got != p {
				t.Errorf("ParsePresentSentinel(%q) = (%q, %v), want (%q, true)", pname, got, ok, p)
			}
			if got, ok := ParseClearSentinel(cname); !ok || got != p {
				t.Errorf("ParseClearSentinel(%q) = (%q, %v), want (%q, true)", cname, got, ok, p)
			}
			if _, ok := ParsePresentSentinel(cname); ok {
				t.Errorf("present prefix matched clear-prefixed name %q", cname)
			}
			if _, ok := ParseClearSentinel(pname); ok {
				t.Errorf("clear prefix matched present-prefixed name %q", pname)
			}
		})
	}
}

func TestSentinelPrefixRejectsPlainName(t *testing.T) {
	for _, n := range []string{"Name", "__pres__Name", "something"} {
		if _, ok := ParsePresentSentinel(n); ok {
			t.Errorf("ParsePresentSentinel(%q) = ok, want not ok", n)
		}
		if _, ok := ParseClearSentinel(n); ok {
			t.Errorf("ParseClearSentinel(%q) = ok, want not ok", n)
		}
	}
}

func TestFieldPathAppend(t *testing.T) {
	cases := []struct {
		base FieldPath
		name string
		want FieldPath
	}{
		{"", "Name", "Name"},
		{"Branding", "PrimaryColor", "Branding-PrimaryColor"},
		{"Branding-PrimaryColor", "Hue", "Branding-PrimaryColor-Hue"},
		{"Name", "", "Name"},
	}
	for _, c := range cases {
		got := c.base.Append(c.name)
		if got != c.want {
			t.Errorf("FieldPath(%q).Append(%q) = %q, want %q", c.base, c.name, got, c.want)
		}
	}
}
