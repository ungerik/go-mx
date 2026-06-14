package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func InputOTPDemo() mx.Component {
	return shadcn.InputOTP("otp", "one-time-code", 6)
}

func InputOTPWithSeparator() mx.Component {
	return html.Div(html.Class("flex items-center gap-2"),
		shadcn.InputOTP("otp-a", "otp-first", 3),
		shadcn.InputOTPSeparator(),
		shadcn.InputOTP("otp-b", "otp-second", 3),
	)
}
