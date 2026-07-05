package html

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenericAttribs(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		// Number-or-string attributes accept both forms.
		{"max number", Input(Max(100)).String(), `<input max='100'/>`},
		{"max date string", Input(Max("2025-01-01")).String(), `<input max='2025-01-01'/>`},
		{"min number", Input(Min(0)).String(), `<input min='0'/>`},
		{"step number", Input(Step(0.5)).String(), `<input step='0.5'/>`},
		{"step any keyword", Input(Step("any")).String(), `<input step='any'/>`},

		// value carries a number for numeric controls and a string otherwise.
		{"value number", Input(Value(42)).String(), `<input value='42'/>`},
		{"value string", Input(Value("hello")).String(), `<input value='hello'/>`},

		// Integer attributes formerly typed as string now also accept numbers.
		{"size number", Input(Size(20)).String(), `<input size='20'/>`},
		{"span number", Col(SpanAttr(2)).String(), `<col span='2'/>`},
		{"start number", OL(Start(3), LI("x")).String(), `<ol start='3'><li>x</li></ol>`},
		{"tabindex number", Div(TabIndex(-1)).String(), `<div tabindex='-1'></div>`},
		{"tabindex string still works", Div(TabIndex("0")).String(), `<div tabindex='0'></div>`},

		// Floats render as plain decimals, never scientific notation.
		{"float plain decimal", Input(Value(0.00005)).String(), `<input value='0.00005'/>`},

		// The escape-hatch Attrib is generic too.
		{"attrib number", Div(Attrib("data-count", 5)).String(), `<div data-count='5'></div>`},
		{"attrib string", Div(Attrib("data-name", "x")).String(), `<div data-name='x'></div>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.got)
		})
	}
}

func TestPopoverCommandAttribs(t *testing.T) {
	// The popover invoker attributes render as plain string id references.
	require.Equal(t,
		`<button popovertarget='menu'>Open</button>`,
		Button(PopoverTarget("menu"), "Open").String(),
	)
	require.Equal(t,
		`<button commandfor='dialog'>Open</button>`,
		Button(CommandFor("dialog"), "Open").String(),
	)

	// Built-in command keyword constants render as command='<keyword>'.
	require.Equal(t,
		`<button command='show-modal' commandfor='dialog'>Open</button>`,
		Button(CommandShowModal, CommandFor("dialog"), "Open").String(),
	)
	require.Equal(t,
		`<button command='toggle-popover'>Open</button>`,
		Button(CommandTogglePopover, "Open").String(),
	)

	// The Command constructor accepts author-defined custom commands (must start with --).
	require.Equal(t,
		`<button command='--rotate' commandfor='img'>Rotate</button>`,
		Button(Command("--rotate"), CommandFor("img"), "Rotate").String(),
	)
}
