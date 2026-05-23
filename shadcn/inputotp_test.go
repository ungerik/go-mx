package shadcn

import (
	"strings"
	"testing"
)

func TestInputOTPDefault(t *testing.T) {
	out := render(t, InputOTP("code", "otp", 6))
	for _, want := range []string{
		`data-slot="input-otp"`,
		`role="group"`,
		`data-input-otp="code"`,
		`data-slot="input-otp-slot"`,
		`maxlength="1"`,
		`inputmode="numeric"`,
		`data-otp-id="code"`,
		`data-otp-index="0"`,
		`data-otp-index="5"`,
		`id="code-slot-0"`,
		`id="code-slot-5"`,
		`autocomplete="one-time-code"`,        // only on the first slot
		`oninput="otpAdvance(this)"`,
		`onkeydown="otpKey(this,event)"`,
		`type="hidden"`,
		`name="otp"`,
		`id="code-value"`,
		"<script>",
		"window.otpAdvance",
		"window.otpKey",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// Count actual <input data-slot=...> openings, not the script's selector
	// string which also contains data-slot="input-otp-slot".
	if n := strings.Count(out, `<input data-slot="input-otp-slot"`); n != 6 {
		t.Errorf("expected 6 slot inputs, got %d: %s", n, out)
	}
	if n := strings.Count(out, `autocomplete="one-time-code"`); n != 1 {
		t.Errorf("autocomplete should be on exactly the first slot, got %d: %s", n, out)
	}
	if n := strings.Count(out, "<script>"); n != 1 {
		t.Errorf("focus script should be emitted once, got %d: %s", n, out)
	}
}

func TestInputOTPSeparator(t *testing.T) {
	out := render(t, InputOTPSeparator())
	for _, want := range []string{
		`data-slot="input-otp-separator"`,
		`role="separator"`,
		"lucide-minus",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestInputOTPSeparatorCustomChild(t *testing.T) {
	out := render(t, InputOTPSeparator("—"))
	if strings.Contains(out, "lucide-minus") {
		t.Errorf("explicit children should suppress default icon: %s", out)
	}
	if !strings.Contains(out, "—") {
		t.Errorf("expected caller child: %s", out)
	}
}

func TestInputOTPPanics(t *testing.T) {
	for _, bad := range []string{"", "bad id"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected panic for id %q", bad)
				}
			}()
			_ = InputOTP(bad, "n", 4)
		}()
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Errorf("expected panic for length 0")
			}
		}()
		_ = InputOTP("c", "n", 0)
	}()
}
