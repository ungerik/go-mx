package magicui

import (
	"strings"
	"testing"
)

// devicePercent formats part/whole as a CSS percentage, computed in float64
// steps like upstream's JavaScript so the digits match.
func devicePercent(part, whole float64) string {
	return num(part/whole*100) + "%"
}

// TestDeviceScreenClasses: the screen classes of Iphone and Safari are
// literals, so Tailwind finds them, and they have to stay the screen
// constants as percentages of the frame.
func TestDeviceScreenClasses(t *testing.T) {
	for _, tc := range []struct {
		classes string
		want    []string
	}{
		{iphoneScreenClasses, []string{
			"left-[" + devicePercent(iphoneScreenX, iphoneWidth) + "]",
			"top-[" + devicePercent(iphoneScreenY, iphoneHeight) + "]",
			"w-[" + devicePercent(iphoneScreenWidth, iphoneWidth) + "]",
			"h-[" + devicePercent(iphoneScreenHeight, iphoneHeight) + "]",
			"[border-radius:" + devicePercent(iphoneScreenRadius, iphoneScreenWidth) + "_/_" + devicePercent(iphoneScreenRadius, iphoneScreenHeight) + "]",
		}},
		{safariScreenClasses, []string{
			"left-[" + devicePercent(safariScreenX, safariWidth) + "]",
			"top-[" + devicePercent(safariScreenY, safariHeight) + "]",
			"w-[" + devicePercent(safariScreenWidth, safariWidth) + "]",
			"h-[" + devicePercent(safariScreenHeight, safariHeight) + "]",
		}},
	} {
		fields := strings.Fields(tc.classes)
		if len(fields) != len(tc.want) {
			t.Errorf("%q has %d classes, want %d", tc.classes, len(fields), len(tc.want))
		}
		for _, want := range tc.want {
			if !strings.Contains(tc.classes, want) {
				t.Errorf("%q lacks %s", tc.classes, want)
			}
		}
	}
}
