package twmerge

import "testing"

// TestMerge is derived verbatim from tailwind-merge v3.6.0's own test suite
// (tests/*.test.ts). Each entry is one upstream
// `expect(twMerge(...)).toBe(...)` assertion.
func TestMerge(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		// --- conflicts-across-class-groups ---
		{"inset-1 inset-x-1", "inset-1 inset-x-1"},
		{"inset-x-1 inset-1", "inset-1"},
		{"inset-x-1 left-1 inset-1", "inset-1"},
		{"inset-x-1 inset-1 left-1", "inset-1 left-1"},
		{"inset-x-1 right-1 inset-1", "inset-1"},
		{"inset-x-1 right-1 inset-x-1", "inset-x-1"},
		{"inset-x-1 right-1 inset-y-1", "inset-x-1 right-1 inset-y-1"},
		{"right-1 inset-x-1 inset-y-1", "inset-x-1 inset-y-1"},
		{"inset-x-1 hover:left-1 inset-1", "hover:left-1 inset-1"},
		{"ring shadow", "ring shadow"},
		{"ring-2 shadow-md", "ring-2 shadow-md"},
		{"shadow ring", "shadow ring"},
		{"shadow-md ring-2", "shadow-md ring-2"},
		{"touch-pan-x touch-pan-right", "touch-pan-right"},
		{"touch-none touch-pan-x", "touch-pan-x"},
		{"touch-pan-x touch-none", "touch-none"},
		{"touch-pan-x touch-pan-y touch-pinch-zoom", "touch-pan-x touch-pan-y touch-pinch-zoom"},
		{"touch-manipulation touch-pan-x touch-pan-y touch-pinch-zoom", "touch-pan-x touch-pan-y touch-pinch-zoom"},
		{"touch-pan-x touch-pan-y touch-pinch-zoom touch-auto", "touch-auto"},
		{"overflow-auto inline line-clamp-1", "line-clamp-1"},
		{"line-clamp-1 overflow-auto inline", "line-clamp-1 overflow-auto inline"},

		// --- class-group-conflicts ---
		{"overflow-x-auto overflow-x-hidden", "overflow-x-hidden"},
		{"basis-full basis-auto", "basis-auto"},
		{"w-full w-fit", "w-fit"},
		{"overflow-x-auto overflow-x-hidden overflow-x-scroll", "overflow-x-scroll"},
		{"overflow-x-auto hover:overflow-x-hidden overflow-x-scroll", "hover:overflow-x-hidden overflow-x-scroll"},
		{"overflow-x-auto hover:overflow-x-hidden hover:overflow-x-auto overflow-x-scroll", "hover:overflow-x-auto overflow-x-scroll"},
		{"col-span-1 col-span-full", "col-span-full"},
		{"gap-2 gap-px basis-px basis-3", "gap-px basis-3"},
		{"lining-nums tabular-nums diagonal-fractions", "lining-nums tabular-nums diagonal-fractions"},
		{"normal-nums tabular-nums diagonal-fractions", "tabular-nums diagonal-fractions"},
		{"tabular-nums diagonal-fractions normal-nums", "normal-nums"},
		{"tabular-nums proportional-nums", "proportional-nums"},

		// --- standalone-classes ---
		{"inline block", "block"},
		{"hover:block hover:inline", "hover:inline"},
		{"hover:block hover:block", "hover:block"},
		{"inline hover:inline focus:inline hover:block hover:focus:block", "inline focus:inline hover:block hover:focus:block"},
		{"underline line-through", "line-through"},
		{"line-through no-underline", "no-underline"},

		// --- non-tailwind-classes ---
		{"non-tailwind-class inline block", "non-tailwind-class block"},
		{"inline block inline-1", "block inline-1"},
		{"inline block i-inline", "block i-inline"},
		{"focus:inline focus:block focus:inline-1", "focus:block focus:inline-1"},

		// --- colors ---
		{"bg-grey-5 bg-hotpink", "bg-hotpink"},
		{"hover:bg-grey-5 hover:bg-hotpink", "hover:bg-hotpink"},
		{"stroke-[hsl(350_80%_0%)] stroke-[10px]", "stroke-[hsl(350_80%_0%)] stroke-[10px]"},

		// --- content-utilities ---
		{"content-['hello'] content-[attr(data-content)]", "content-[attr(data-content)]"},

		// --- negative-values ---
		{"-m-2 -m-5", "-m-5"},
		{"-top-12 -top-2000", "-top-2000"},
		{"-m-2 m-auto", "m-auto"},
		{"top-12 -top-69", "-top-69"},
		{"-right-1 inset-x-1", "inset-x-1"},
		{"hover:focus:-right-1 focus:hover:inset-x-1", "focus:hover:inset-x-1"},

		// --- important-modifier ---
		{"font-medium! font-bold!", "font-bold!"},
		{"font-medium! font-bold! font-thin", "font-bold! font-thin"},
		{"right-2! -inset-x-px!", "-inset-x-px!"},
		{"focus:inline! focus:block!", "focus:block!"},
		{"[--my-var:20px]! [--my-var:30px]!", "[--my-var:30px]!"},
		{"font-medium! !font-bold", "!font-bold"},
		{"!font-medium !font-bold", "!font-bold"},
		{"!font-medium !font-bold font-thin", "!font-bold font-thin"},
		{"!right-2 !-inset-x-px", "!-inset-x-px"},
		{"focus:!inline focus:!block", "focus:!block"},
		{"![--my-var:20px] ![--my-var:30px]", "![--my-var:30px]"},

		// --- arbitrary-values ---
		{"m-[2px] m-[10px]", "m-[10px]"},
		{"m-[2px] m-[11svmin] m-[12in] m-[13lvi] m-[14vb] m-[15vmax] m-[16mm] m-[17%] m-[18em] m-[19px] m-[10dvh]", "m-[10dvh]"},
		{"h-[10px] h-[11cqw] h-[12cqh] h-[13cqi] h-[14cqb] h-[15cqmin] h-[16cqmax]", "h-[16cqmax]"},
		{"z-20 z-[99]", "z-[99]"},
		{"my-[2px] m-[10rem]", "m-[10rem]"},
		{"cursor-pointer cursor-[grab]", "cursor-[grab]"},
		{"m-[2px] m-[calc(100%-var(--arbitrary))]", "m-[calc(100%-var(--arbitrary))]"},
		{"m-[2px] m-[length:var(--mystery-var)]", "m-[length:var(--mystery-var)]"},
		{"opacity-10 opacity-[0.025]", "opacity-[0.025]"},
		{"scale-75 scale-[1.7]", "scale-[1.7]"},
		{"brightness-90 brightness-[1.75]", "brightness-[1.75]"},
		{"min-h-[0.5px] min-h-[0]", "min-h-[0]"},
		{"text-[0.5px] text-[color:0]", "text-[0.5px] text-[color:0]"},
		{"text-[0.5px] text-(--my-0)", "text-[0.5px] text-(--my-0)"},
		{"hover:m-[2px] hover:m-[length:var(--c)]", "hover:m-[length:var(--c)]"},
		{"hover:focus:m-[2px] focus:hover:m-[length:var(--c)]", "focus:hover:m-[length:var(--c)]"},
		{"grid-rows-[1fr,auto] grid-rows-2", "grid-rows-2"},
		{"grid-rows-[repeat(20,minmax(0,1fr))] grid-rows-3", "grid-rows-3"},
		{"mt-2 mt-[calc(theme(fontSize.4xl)/1.125)]", "mt-[calc(theme(fontSize.4xl)/1.125)]"},
		{"mt-2 mt-[length:theme(someScale.someValue)]", "mt-[length:theme(someScale.someValue)]"},
		{"mt-2 mt-[theme(someScale.someValue)]", "mt-[theme(someScale.someValue)]"},
		{"text-2xl text-[length:theme(someScale.someValue)]", "text-[length:theme(someScale.someValue)]"},
		{"text-2xl text-[calc(theme(fontSize.4xl)/1.125)]", "text-[calc(theme(fontSize.4xl)/1.125)]"},
		{"font-[400] font-[600]", "font-[600]"},
		{"font-[var(--a)] font-[var(--b)]", "font-[var(--b)]"},
		{"font-[weight:var(--a)] font-[var(--b)]", "font-[var(--b)]"},
		{"font-[400] font-[weight:var(--b)]", "font-[weight:var(--b)]"},
		{"font-[weight:var(--a)] font-[weight:var(--b)]", "font-[weight:var(--b)]"},
		{"font-[family-name:var(--a)] font-[var(--b)]", "font-[family-name:var(--a)] font-[var(--b)]"},
		{"bg-red bg-(--other-red) bg-bottom bg-(position:-my-pos)", "bg-(--other-red) bg-(position:-my-pos)"},
		{"font-(--a) font-(--b)", "font-(--b)"},
		{"font-(weight:--a) font-(--b)", "font-(--b)"},
		{"font-(family-name:--a) font-(--b)", "font-(family-name:--a) font-(--b)"},

		// --- arbitrary-properties ---
		{"[paint-order:markers] [paint-order:normal]", "[paint-order:normal]"},
		{"[paint-order:markers] [--my-var:2rem] [paint-order:normal] [--my-var:4px]", "[paint-order:normal] [--my-var:4px]"},
		{"[paint-order:markers] hover:[paint-order:normal]", "[paint-order:markers] hover:[paint-order:normal]"},
		{"hover:[paint-order:markers] hover:[paint-order:normal]", "hover:[paint-order:normal]"},
		{"hover:focus:[paint-order:markers] focus:hover:[paint-order:normal]", "focus:hover:[paint-order:normal]"},
		{"[paint-order:markers] [paint-order:normal] [--my-var:2rem] lg:[--my-var:4px]", "[paint-order:normal] [--my-var:2rem] lg:[--my-var:4px]"},
		{"[-unknown-prop:::123:::] [-unknown-prop:url(https://hi.com)]", "[-unknown-prop:url(https://hi.com)]"},
		{"![some:prop] [some:other]", "![some:prop] [some:other]"},
		{"![some:prop] [some:other] [some:one] ![some:another]", "[some:one] ![some:another]"},

		// --- arbitrary-variants ---
		{"[p]:underline [p]:line-through", "[p]:line-through"},
		{"[&>*]:underline [&>*]:line-through", "[&>*]:line-through"},
		{"[&>*]:underline [&>*]:line-through [&_div]:line-through", "[&>*]:line-through [&_div]:line-through"},
		{"supports-[display:grid]:flex supports-[display:grid]:grid", "supports-[display:grid]:grid"},
		{"dark:lg:hover:[&>*]:underline dark:lg:hover:[&>*]:line-through", "dark:lg:hover:[&>*]:line-through"},
		{"dark:lg:hover:[&>*]:underline dark:hover:lg:[&>*]:line-through", "dark:hover:lg:[&>*]:line-through"},
		{"hover:[&>*]:underline [&>*]:hover:line-through", "hover:[&>*]:underline [&>*]:hover:line-through"},
		{"[&[data-open]]:underline [&[data-open]]:line-through", "[&[data-open]]:line-through"},
		{"[&>*]:[&_div]:underline [&>*]:[&_div]:line-through", "[&>*]:[&_div]:line-through"},
		{"[&>*]:[&_div]:underline [&_div]:[&>*]:line-through", "[&>*]:[&_div]:underline [&_div]:[&>*]:line-through"},
		{"[&>*]:[color:red] [&>*]:[color:blue]", "[&>*]:[color:blue]"},

		// --- modifiers (prefix / postfix / sorting) ---
		{"hover:block hover:focus:inline", "hover:block hover:focus:inline"},
		{"hover:block hover:focus:inline focus:hover:inline", "hover:block focus:hover:inline"},
		{"focus-within:inline focus-within:block", "focus-within:block"},
		{"text-lg/7 text-lg/8", "text-lg/8"},
		{"text-lg/none leading-9", "text-lg/none leading-9"},
		{"leading-9 text-lg/none", "text-lg/none"},
		{"w-full w-1/2", "w-1/2"},
		{"c:d:e:block d:c:e:inline", "d:c:e:inline"},
		{"*:before:block *:before:inline", "*:before:inline"},
		{"*:before:block before:*:inline", "*:before:block before:*:inline"},
		{"x:y:*:z:block y:x:*:z:inline", "y:x:*:z:inline"},

		// --- wonky-inputs ---
		{" block", "block"},
		{"block ", "block"},
		{" block ", "block"},
		{"  block  px-2     py-4  ", "block px-2 py-4"},
		{"block\npx-2", "block px-2"},
		{"\nblock\npx-2\n", "block px-2"},

		// --- tailwind-css-versions: v3.3 / v3.4 ---
		{"text-red text-lg/7 text-lg/8", "text-red text-lg/8"},
		{"hyphens-auto hyphens-manual", "hyphens-manual"},
		{"from-0% from-10% from-[12.5%]", "from-[12.5%]"},
		{"from-0% from-red", "from-0% from-red"},
		{"caption-top caption-bottom", "caption-bottom"},
		{"line-clamp-2 line-clamp-none line-clamp-[10]", "line-clamp-[10]"},
		{"delay-150 delay-0 duration-150 duration-0", "delay-0 duration-0"},
		{"justify-normal justify-center justify-stretch", "justify-stretch"},
		{"content-normal content-center content-stretch", "content-stretch"},
		{"whitespace-nowrap whitespace-break-spaces", "whitespace-break-spaces"},
		{"h-svh h-dvh w-svw w-dvw", "h-dvh w-dvw"},
		{"has-[[data-potato]]:p-1 has-[[data-potato]]:p-2 group-has-[:checked]:grid group-has-[:checked]:flex", "has-[[data-potato]]:p-2 group-has-[:checked]:flex"},
		{"text-wrap text-pretty", "text-pretty"},
		{"w-5 h-3 size-10 w-12", "size-10 w-12"},
		{"grid-cols-2 grid-cols-subgrid grid-rows-5 grid-rows-subgrid", "grid-cols-subgrid grid-rows-subgrid"},
		{"min-w-0 min-w-50 min-w-px max-w-0 max-w-50 max-w-px", "min-w-px max-w-px"},
		{"float-start float-end clear-start clear-end", "float-end clear-end"},
		{"*:p-10 *:p-20 hover:*:p-10 hover:*:p-20", "*:p-20 hover:*:p-20"},

		// --- tailwind-css-versions: v4.0 / v4.1 ---
		{"transform-3d transform-flat", "transform-flat"},
		{"rotate-12 rotate-x-2 rotate-none rotate-y-3", "rotate-x-2 rotate-none rotate-y-3"},
		{"perspective-dramatic perspective-none perspective-midrange", "perspective-midrange"},
		{"bg-linear-to-r bg-linear-45", "bg-linear-45"},
		{"ring-4 ring-orange inset-ring inset-ring-3 inset-ring-blue", "ring-4 ring-orange inset-ring-3 inset-ring-blue"},
		{"field-sizing-content field-sizing-fixed", "field-sizing-fixed"},
		{"scheme-normal scheme-dark", "scheme-dark"},
		{"col-span-full col-2 row-span-3 row-4", "col-2 row-4"},
		{"items-baseline items-baseline-last", "items-baseline-last"},
		{"wrap-break-word wrap-normal wrap-anywhere", "wrap-anywhere"},
		{"text-shadow-none text-shadow-2xl", "text-shadow-2xl"},
		{"mask-add mask-subtract", "mask-subtract"},
		{"shadow-md shadow-lg/25 text-shadow-md text-shadow-lg/25", "shadow-lg/25 text-shadow-lg/25"},

		// --- tailwind-css-versions: v4.2 / v4.3 ---
		{"inset-s-1 inset-s-2", "inset-s-2"},
		{"start-1 inset-s-2", "inset-s-2"},
		{"inset-s-1 start-2", "start-2"},
		{"inset-s-1 inset-e-2 inset-bs-3 inset-be-4 inset-0", "inset-0"},
		{"inset-0 inset-s-1 inset-bs-1", "inset-0 inset-s-1 inset-bs-1"},
		{"pbs-1 pbs-2", "pbs-2"},
		{"pt-1 pbs-2", "pt-1 pbs-2"},
		{"pbs-1 pbe-1 p-0", "p-0"},
		{"border-bs-1 border-bs-2", "border-bs-2"},
		{"border-2 border-bs-4 border-be-6", "border-2 border-bs-4 border-be-6"},
		{"border-bs-4 border-be-6 border-2", "border-2"},
		{"inline-1/2 inline-3/4", "inline-3/4"},
		{"size-10 inline-20 block-30", "size-10 inline-20 block-30"},
		{"aspect-8/11 aspect-8.5/11", "aspect-8.5/11"},
		{"w-8/11 w-8.5/11", "w-8.5/11"},
		{"scrollbar-auto scrollbar-thin scrollbar-none", "scrollbar-none"},
		{"@container @container-normal @container-size", "@container-size"},
		{"@container @container-size/sidebar", "@container-size/sidebar"},
		{"@container-size/sidebar @container", "@container-size/sidebar @container"},
		{"zoom-50 zoom-100", "zoom-100"},
		{"zoom-50 scale-125", "zoom-50 scale-125"},
		{"tab-2 tab-8", "tab-8"},
		{"tab-4 tabular-nums", "tab-4 tabular-nums"},
	}

	for _, c := range cases {
		if got := Merge(c.in); got != c.want {
			t.Errorf("Merge(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
