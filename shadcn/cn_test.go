package shadcn

import "testing"

// TestCn keeps the original package test cases. Case 3's expectation was
// corrected to match real tailwind-merge: an important and a non-important
// class do not conflict, so both are kept.
func TestCn(t *testing.T) {
	examples := []struct {
		input    []any
		expected string
	}{
		{[]any{"px-2 py-1", "bg-red-500"}, "px-2 py-1 bg-red-500"},
		{[]any{"px-2", "p-4"}, "p-4"},
		{[]any{"px-2", "!px-4"}, "px-2 !px-4"},
		{[]any{"px-2", "", false, "py-1"}, "px-2 py-1"},
	}
	for _, ex := range examples {
		if got := Cn(ex.input...); got != ex.expected {
			t.Fatalf("Cn(%v) = %q, want %q", ex.input, got, ex.expected)
		}
	}
}

// TestCnFlatten covers the clsx flatten layer: slices, conditional maps and
// falsy values.
func TestCnFlatten(t *testing.T) {
	cases := []struct {
		in   []any
		want string
	}{
		{[]any{[]string{"px-2", "py-1"}}, "px-2 py-1"},
		{[]any{[]string{"p-2", "", "p-4"}}, "p-4"},
		{[]any{"flex", map[string]bool{"hidden": false, "block": false}}, "flex"},
		{[]any{map[string]bool{"font-bold": true}}, "font-bold"},
		{[]any{nil, false, "", "block"}, "block"},
		{[]any{[]any{"foo", []any{"bar", []any{"", []any{[]any{"baz"}}}}}}, "foo bar baz"},
		{[]any{[]any{}, []any{}}, ""},
	}
	for _, c := range cases {
		if got := Cn(c.in...); got != c.want {
			t.Errorf("Cn(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestTwMerge is derived verbatim from tailwind-merge v3.6.0's own test
// suite (tests/*.test.ts). Each entry is one upstream `expect(twMerge(...))
// .toBe(...)` assertion.
func TestTwMerge(t *testing.T) {
	cases := []struct {
		in   []any
		want string
	}{
		// --- conflicts-across-class-groups ---
		{[]any{"inset-1 inset-x-1"}, "inset-1 inset-x-1"},
		{[]any{"inset-x-1 inset-1"}, "inset-1"},
		{[]any{"inset-x-1 left-1 inset-1"}, "inset-1"},
		{[]any{"inset-x-1 inset-1 left-1"}, "inset-1 left-1"},
		{[]any{"inset-x-1 right-1 inset-1"}, "inset-1"},
		{[]any{"inset-x-1 right-1 inset-x-1"}, "inset-x-1"},
		{[]any{"inset-x-1 right-1 inset-y-1"}, "inset-x-1 right-1 inset-y-1"},
		{[]any{"right-1 inset-x-1 inset-y-1"}, "inset-x-1 inset-y-1"},
		{[]any{"inset-x-1 hover:left-1 inset-1"}, "hover:left-1 inset-1"},
		{[]any{"ring shadow"}, "ring shadow"},
		{[]any{"ring-2 shadow-md"}, "ring-2 shadow-md"},
		{[]any{"shadow ring"}, "shadow ring"},
		{[]any{"shadow-md ring-2"}, "shadow-md ring-2"},
		{[]any{"touch-pan-x touch-pan-right"}, "touch-pan-right"},
		{[]any{"touch-none touch-pan-x"}, "touch-pan-x"},
		{[]any{"touch-pan-x touch-none"}, "touch-none"},
		{[]any{"touch-pan-x touch-pan-y touch-pinch-zoom"}, "touch-pan-x touch-pan-y touch-pinch-zoom"},
		{[]any{"touch-manipulation touch-pan-x touch-pan-y touch-pinch-zoom"}, "touch-pan-x touch-pan-y touch-pinch-zoom"},
		{[]any{"touch-pan-x touch-pan-y touch-pinch-zoom touch-auto"}, "touch-auto"},
		{[]any{"overflow-auto inline line-clamp-1"}, "line-clamp-1"},
		{[]any{"line-clamp-1 overflow-auto inline"}, "line-clamp-1 overflow-auto inline"},

		// --- class-group-conflicts ---
		{[]any{"overflow-x-auto overflow-x-hidden"}, "overflow-x-hidden"},
		{[]any{"basis-full basis-auto"}, "basis-auto"},
		{[]any{"w-full w-fit"}, "w-fit"},
		{[]any{"overflow-x-auto overflow-x-hidden overflow-x-scroll"}, "overflow-x-scroll"},
		{[]any{"overflow-x-auto hover:overflow-x-hidden overflow-x-scroll"}, "hover:overflow-x-hidden overflow-x-scroll"},
		{[]any{"overflow-x-auto hover:overflow-x-hidden hover:overflow-x-auto overflow-x-scroll"}, "hover:overflow-x-auto overflow-x-scroll"},
		{[]any{"col-span-1 col-span-full"}, "col-span-full"},
		{[]any{"gap-2 gap-px basis-px basis-3"}, "gap-px basis-3"},
		{[]any{"lining-nums tabular-nums diagonal-fractions"}, "lining-nums tabular-nums diagonal-fractions"},
		{[]any{"normal-nums tabular-nums diagonal-fractions"}, "tabular-nums diagonal-fractions"},
		{[]any{"tabular-nums diagonal-fractions normal-nums"}, "normal-nums"},
		{[]any{"tabular-nums proportional-nums"}, "proportional-nums"},

		// --- standalone-classes ---
		{[]any{"inline block"}, "block"},
		{[]any{"hover:block hover:inline"}, "hover:inline"},
		{[]any{"hover:block hover:block"}, "hover:block"},
		{[]any{"inline hover:inline focus:inline hover:block hover:focus:block"}, "inline focus:inline hover:block hover:focus:block"},
		{[]any{"underline line-through"}, "line-through"},
		{[]any{"line-through no-underline"}, "no-underline"},

		// --- non-tailwind-classes ---
		{[]any{"non-tailwind-class inline block"}, "non-tailwind-class block"},
		{[]any{"inline block inline-1"}, "block inline-1"},
		{[]any{"inline block i-inline"}, "block i-inline"},
		{[]any{"focus:inline focus:block focus:inline-1"}, "focus:block focus:inline-1"},

		// --- colors ---
		{[]any{"bg-grey-5 bg-hotpink"}, "bg-hotpink"},
		{[]any{"hover:bg-grey-5 hover:bg-hotpink"}, "hover:bg-hotpink"},
		{[]any{"stroke-[hsl(350_80%_0%)] stroke-[10px]"}, "stroke-[hsl(350_80%_0%)] stroke-[10px]"},

		// --- content-utilities ---
		{[]any{"content-['hello'] content-[attr(data-content)]"}, "content-[attr(data-content)]"},

		// --- negative-values ---
		{[]any{"-m-2 -m-5"}, "-m-5"},
		{[]any{"-top-12 -top-2000"}, "-top-2000"},
		{[]any{"-m-2 m-auto"}, "m-auto"},
		{[]any{"top-12 -top-69"}, "-top-69"},
		{[]any{"-right-1 inset-x-1"}, "inset-x-1"},
		{[]any{"hover:focus:-right-1 focus:hover:inset-x-1"}, "focus:hover:inset-x-1"},

		// --- important-modifier ---
		{[]any{"font-medium! font-bold!"}, "font-bold!"},
		{[]any{"font-medium! font-bold! font-thin"}, "font-bold! font-thin"},
		{[]any{"right-2! -inset-x-px!"}, "-inset-x-px!"},
		{[]any{"focus:inline! focus:block!"}, "focus:block!"},
		{[]any{"[--my-var:20px]! [--my-var:30px]!"}, "[--my-var:30px]!"},
		{[]any{"font-medium! !font-bold"}, "!font-bold"},
		{[]any{"!font-medium !font-bold"}, "!font-bold"},
		{[]any{"!font-medium !font-bold font-thin"}, "!font-bold font-thin"},
		{[]any{"!right-2 !-inset-x-px"}, "!-inset-x-px"},
		{[]any{"focus:!inline focus:!block"}, "focus:!block"},
		{[]any{"![--my-var:20px] ![--my-var:30px]"}, "![--my-var:30px]"},

		// --- arbitrary-values ---
		{[]any{"m-[2px] m-[10px]"}, "m-[10px]"},
		{[]any{"m-[2px] m-[11svmin] m-[12in] m-[13lvi] m-[14vb] m-[15vmax] m-[16mm] m-[17%] m-[18em] m-[19px] m-[10dvh]"}, "m-[10dvh]"},
		{[]any{"h-[10px] h-[11cqw] h-[12cqh] h-[13cqi] h-[14cqb] h-[15cqmin] h-[16cqmax]"}, "h-[16cqmax]"},
		{[]any{"z-20 z-[99]"}, "z-[99]"},
		{[]any{"my-[2px] m-[10rem]"}, "m-[10rem]"},
		{[]any{"cursor-pointer cursor-[grab]"}, "cursor-[grab]"},
		{[]any{"m-[2px] m-[calc(100%-var(--arbitrary))]"}, "m-[calc(100%-var(--arbitrary))]"},
		{[]any{"m-[2px] m-[length:var(--mystery-var)]"}, "m-[length:var(--mystery-var)]"},
		{[]any{"opacity-10 opacity-[0.025]"}, "opacity-[0.025]"},
		{[]any{"scale-75 scale-[1.7]"}, "scale-[1.7]"},
		{[]any{"brightness-90 brightness-[1.75]"}, "brightness-[1.75]"},
		{[]any{"min-h-[0.5px] min-h-[0]"}, "min-h-[0]"},
		{[]any{"text-[0.5px] text-[color:0]"}, "text-[0.5px] text-[color:0]"},
		{[]any{"text-[0.5px] text-(--my-0)"}, "text-[0.5px] text-(--my-0)"},
		{[]any{"hover:m-[2px] hover:m-[length:var(--c)]"}, "hover:m-[length:var(--c)]"},
		{[]any{"hover:focus:m-[2px] focus:hover:m-[length:var(--c)]"}, "focus:hover:m-[length:var(--c)]"},
		{[]any{"grid-rows-[1fr,auto] grid-rows-2"}, "grid-rows-2"},
		{[]any{"grid-rows-[repeat(20,minmax(0,1fr))] grid-rows-3"}, "grid-rows-3"},
		{[]any{"mt-2 mt-[calc(theme(fontSize.4xl)/1.125)]"}, "mt-[calc(theme(fontSize.4xl)/1.125)]"},
		{[]any{"mt-2 mt-[length:theme(someScale.someValue)]"}, "mt-[length:theme(someScale.someValue)]"},
		{[]any{"mt-2 mt-[theme(someScale.someValue)]"}, "mt-[theme(someScale.someValue)]"},
		{[]any{"text-2xl text-[length:theme(someScale.someValue)]"}, "text-[length:theme(someScale.someValue)]"},
		{[]any{"text-2xl text-[calc(theme(fontSize.4xl)/1.125)]"}, "text-[calc(theme(fontSize.4xl)/1.125)]"},
		{[]any{"font-[400] font-[600]"}, "font-[600]"},
		{[]any{"font-[var(--a)] font-[var(--b)]"}, "font-[var(--b)]"},
		{[]any{"font-[weight:var(--a)] font-[var(--b)]"}, "font-[var(--b)]"},
		{[]any{"font-[400] font-[weight:var(--b)]"}, "font-[weight:var(--b)]"},
		{[]any{"font-[weight:var(--a)] font-[weight:var(--b)]"}, "font-[weight:var(--b)]"},
		{[]any{"font-[family-name:var(--a)] font-[var(--b)]"}, "font-[family-name:var(--a)] font-[var(--b)]"},
		{[]any{"bg-red bg-(--other-red) bg-bottom bg-(position:-my-pos)"}, "bg-(--other-red) bg-(position:-my-pos)"},
		{[]any{"font-(--a) font-(--b)"}, "font-(--b)"},
		{[]any{"font-(weight:--a) font-(--b)"}, "font-(--b)"},
		{[]any{"font-(family-name:--a) font-(--b)"}, "font-(family-name:--a) font-(--b)"},

		// --- arbitrary-properties ---
		{[]any{"[paint-order:markers] [paint-order:normal]"}, "[paint-order:normal]"},
		{[]any{"[paint-order:markers] [--my-var:2rem] [paint-order:normal] [--my-var:4px]"}, "[paint-order:normal] [--my-var:4px]"},
		{[]any{"[paint-order:markers] hover:[paint-order:normal]"}, "[paint-order:markers] hover:[paint-order:normal]"},
		{[]any{"hover:[paint-order:markers] hover:[paint-order:normal]"}, "hover:[paint-order:normal]"},
		{[]any{"hover:focus:[paint-order:markers] focus:hover:[paint-order:normal]"}, "focus:hover:[paint-order:normal]"},
		{[]any{"[paint-order:markers] [paint-order:normal] [--my-var:2rem] lg:[--my-var:4px]"}, "[paint-order:normal] [--my-var:2rem] lg:[--my-var:4px]"},
		{[]any{"[-unknown-prop:::123:::] [-unknown-prop:url(https://hi.com)]"}, "[-unknown-prop:url(https://hi.com)]"},
		{[]any{"![some:prop] [some:other]"}, "![some:prop] [some:other]"},
		{[]any{"![some:prop] [some:other] [some:one] ![some:another]"}, "[some:one] ![some:another]"},

		// --- arbitrary-variants ---
		{[]any{"[p]:underline [p]:line-through"}, "[p]:line-through"},
		{[]any{"[&>*]:underline [&>*]:line-through"}, "[&>*]:line-through"},
		{[]any{"[&>*]:underline [&>*]:line-through [&_div]:line-through"}, "[&>*]:line-through [&_div]:line-through"},
		{[]any{"supports-[display:grid]:flex supports-[display:grid]:grid"}, "supports-[display:grid]:grid"},
		{[]any{"dark:lg:hover:[&>*]:underline dark:lg:hover:[&>*]:line-through"}, "dark:lg:hover:[&>*]:line-through"},
		{[]any{"dark:lg:hover:[&>*]:underline dark:hover:lg:[&>*]:line-through"}, "dark:hover:lg:[&>*]:line-through"},
		{[]any{"hover:[&>*]:underline [&>*]:hover:line-through"}, "hover:[&>*]:underline [&>*]:hover:line-through"},
		{[]any{"[&[data-open]]:underline [&[data-open]]:line-through"}, "[&[data-open]]:line-through"},
		{[]any{"[&>*]:[&_div]:underline [&>*]:[&_div]:line-through"}, "[&>*]:[&_div]:line-through"},
		{[]any{"[&>*]:[&_div]:underline [&_div]:[&>*]:line-through"}, "[&>*]:[&_div]:underline [&_div]:[&>*]:line-through"},
		{[]any{"[&>*]:[color:red] [&>*]:[color:blue]"}, "[&>*]:[color:blue]"},

		// --- modifiers (prefix / postfix / sorting) ---
		{[]any{"hover:block hover:focus:inline"}, "hover:block hover:focus:inline"},
		{[]any{"hover:block hover:focus:inline focus:hover:inline"}, "hover:block focus:hover:inline"},
		{[]any{"focus-within:inline focus-within:block"}, "focus-within:block"},
		{[]any{"text-lg/7 text-lg/8"}, "text-lg/8"},
		{[]any{"text-lg/none leading-9"}, "text-lg/none leading-9"},
		{[]any{"leading-9 text-lg/none"}, "text-lg/none"},
		{[]any{"w-full w-1/2"}, "w-1/2"},
		{[]any{"c:d:e:block d:c:e:inline"}, "d:c:e:inline"},
		{[]any{"*:before:block *:before:inline"}, "*:before:inline"},
		{[]any{"*:before:block before:*:inline"}, "*:before:block before:*:inline"},
		{[]any{"x:y:*:z:block y:x:*:z:inline"}, "y:x:*:z:inline"},

		// --- wonky-inputs ---
		{[]any{" block"}, "block"},
		{[]any{"block "}, "block"},
		{[]any{" block "}, "block"},
		{[]any{"  block  px-2     py-4  "}, "block px-2 py-4"},
		{[]any{"  block  px-2", " ", "     py-4  "}, "block px-2 py-4"},
		{[]any{"block\npx-2"}, "block px-2"},
		{[]any{"\nblock\npx-2\n"}, "block px-2"},

		// --- tailwind-css-versions: v3.3 / v3.4 ---
		{[]any{"text-red text-lg/7 text-lg/8"}, "text-red text-lg/8"},
		{[]any{"hyphens-auto hyphens-manual"}, "hyphens-manual"},
		{[]any{"from-0% from-10% from-[12.5%]"}, "from-[12.5%]"},
		{[]any{"from-0% from-red"}, "from-0% from-red"},
		{[]any{"caption-top caption-bottom"}, "caption-bottom"},
		{[]any{"line-clamp-2 line-clamp-none line-clamp-[10]"}, "line-clamp-[10]"},
		{[]any{"delay-150 delay-0 duration-150 duration-0"}, "delay-0 duration-0"},
		{[]any{"justify-normal justify-center justify-stretch"}, "justify-stretch"},
		{[]any{"content-normal content-center content-stretch"}, "content-stretch"},
		{[]any{"whitespace-nowrap whitespace-break-spaces"}, "whitespace-break-spaces"},
		{[]any{"h-svh h-dvh w-svw w-dvw"}, "h-dvh w-dvw"},
		{[]any{"has-[[data-potato]]:p-1 has-[[data-potato]]:p-2 group-has-[:checked]:grid group-has-[:checked]:flex"}, "has-[[data-potato]]:p-2 group-has-[:checked]:flex"},
		{[]any{"text-wrap text-pretty"}, "text-pretty"},
		{[]any{"w-5 h-3 size-10 w-12"}, "size-10 w-12"},
		{[]any{"grid-cols-2 grid-cols-subgrid grid-rows-5 grid-rows-subgrid"}, "grid-cols-subgrid grid-rows-subgrid"},
		{[]any{"min-w-0 min-w-50 min-w-px max-w-0 max-w-50 max-w-px"}, "min-w-px max-w-px"},
		{[]any{"float-start float-end clear-start clear-end"}, "float-end clear-end"},
		{[]any{"*:p-10 *:p-20 hover:*:p-10 hover:*:p-20"}, "*:p-20 hover:*:p-20"},

		// --- tailwind-css-versions: v4.0 / v4.1 ---
		{[]any{"transform-3d transform-flat"}, "transform-flat"},
		{[]any{"rotate-12 rotate-x-2 rotate-none rotate-y-3"}, "rotate-x-2 rotate-none rotate-y-3"},
		{[]any{"perspective-dramatic perspective-none perspective-midrange"}, "perspective-midrange"},
		{[]any{"bg-linear-to-r bg-linear-45"}, "bg-linear-45"},
		{[]any{"ring-4 ring-orange inset-ring inset-ring-3 inset-ring-blue"}, "ring-4 ring-orange inset-ring-3 inset-ring-blue"},
		{[]any{"field-sizing-content field-sizing-fixed"}, "field-sizing-fixed"},
		{[]any{"scheme-normal scheme-dark"}, "scheme-dark"},
		{[]any{"col-span-full col-2 row-span-3 row-4"}, "col-2 row-4"},
		{[]any{"items-baseline items-baseline-last"}, "items-baseline-last"},
		{[]any{"wrap-break-word wrap-normal wrap-anywhere"}, "wrap-anywhere"},
		{[]any{"text-shadow-none text-shadow-2xl"}, "text-shadow-2xl"},
		{[]any{"mask-add mask-subtract"}, "mask-subtract"},
		{[]any{"shadow-md shadow-lg/25 text-shadow-md text-shadow-lg/25"}, "shadow-lg/25 text-shadow-lg/25"},

		// --- tailwind-css-versions: v4.2 / v4.3 ---
		{[]any{"inset-s-1 inset-s-2"}, "inset-s-2"},
		{[]any{"start-1 inset-s-2"}, "inset-s-2"},
		{[]any{"inset-s-1 start-2"}, "start-2"},
		{[]any{"inset-s-1 inset-e-2 inset-bs-3 inset-be-4 inset-0"}, "inset-0"},
		{[]any{"inset-0 inset-s-1 inset-bs-1"}, "inset-0 inset-s-1 inset-bs-1"},
		{[]any{"pbs-1 pbs-2"}, "pbs-2"},
		{[]any{"pt-1 pbs-2"}, "pt-1 pbs-2"},
		{[]any{"pbs-1 pbe-1 p-0"}, "p-0"},
		{[]any{"border-bs-1 border-bs-2"}, "border-bs-2"},
		{[]any{"border-2 border-bs-4 border-be-6"}, "border-2 border-bs-4 border-be-6"},
		{[]any{"border-bs-4 border-be-6 border-2"}, "border-2"},
		{[]any{"inline-1/2 inline-3/4"}, "inline-3/4"},
		{[]any{"size-10 inline-20 block-30"}, "size-10 inline-20 block-30"},
		{[]any{"aspect-8/11 aspect-8.5/11"}, "aspect-8.5/11"},
		{[]any{"w-8/11 w-8.5/11"}, "w-8.5/11"},
		{[]any{"scrollbar-auto scrollbar-thin scrollbar-none"}, "scrollbar-none"},
		{[]any{"@container @container-normal @container-size"}, "@container-size"},
		{[]any{"@container @container-size/sidebar"}, "@container-size/sidebar"},
		{[]any{"@container-size/sidebar @container"}, "@container-size/sidebar @container"},
		{[]any{"zoom-50 zoom-100"}, "zoom-100"},
		{[]any{"zoom-50 scale-125"}, "zoom-50 scale-125"},
		{[]any{"tab-2 tab-8"}, "tab-8"},
		{[]any{"tab-4 tabular-nums"}, "tab-4 tabular-nums"},
	}

	for _, c := range cases {
		if got := Cn(c.in...); got != c.want {
			t.Errorf("Cn(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}
