package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// InputOTPDemo renders a six-digit one-time-password input.
func InputOTPDemo() mx.Component {
	return shadcn.InputOTP("otp", "one-time-code", 6)
}

// InputOTPWithSeparator renders two three-digit OTP inputs joined by a separator.
func InputOTPWithSeparator() mx.Component {
	return html.DivClass("flex items-center gap-2",
		shadcn.InputOTP("otp-a", "otp-first", 3),
		shadcn.InputOTPSeparator(),
		shadcn.InputOTP("otp-b", "otp-second", 3),
	)
}
