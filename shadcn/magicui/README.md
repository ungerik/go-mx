# magicui — Magic UI for go-mx

`github.com/ungerik/go-mx/shadcn/magicui` is a Go port of
[Magic UI](https://magicui.design), the animated components and effects built
on shadcn/ui's design tokens. It renders through go-mx and merges classes with
the [`shadcn`](../README.md) package's `Cn`, so the components sit next to
shadcn components and share their theme.

Upstream is a shadcn registry of React components that animate with Tailwind
keyframes and the [motion](https://motion.dev) library. This port is pinned to
[magicuidesign/magicui](https://github.com/magicuidesign/magicui) at
`magicui.UpstreamCommit` and covers the components CSS can drive on its own:
the ones that are markup and CSS upstream, and the motion-library ones whose
animation CSS can re-express. The components that need a script or a
third-party library are listed in [`../TODOS.md`](../TODOS.md#magic-ui-magicui).

This README is the reference. The docs site has a
[tutorial](https://ungerik.github.io/go-mx/shadcn/magicui/tutorial.html)
(an animated landing hero),
[how-to guides](https://ungerik.github.io/go-mx/shadcn/magicui/how-to.html)
(among them the production Tailwind build) and an
[explanation](https://ungerik.github.io/go-mx/shadcn/magicui/explanation.html)
of why the port animates with CSS alone.

## Setup

The animations are Tailwind v4 CSS: `@theme` `--animate-*` entries with their
`@keyframes`, a few `@utility` rules for the scroll-driven ones, and the
reduced-motion block. Add `magicui.CSS` to the app's Tailwind input, after
`@import "tailwindcss"`, for example by writing it to a file at build time:

```go
os.WriteFile("assets/magicui.css", []byte(magicui.CSS), 0o644)
```

```css
@import "tailwindcss";
@import "./magicui.css";
```

Without it the components render, but do not move. Tailwind also has to see
the class names in the Go source, like for `shadcn` (see
["Tailwind CSS v4 is required"](../README.md#tailwind-css-v4-is-required)).
Those names are in the go-mx source in the Go module cache, which Tailwind
does not scan on its own, so a production build needs an `@source` line
pointing there; the how-to
["Build the Magic UI CSS for production"](https://ungerik.github.io/go-mx/shadcn/magicui/how-to.html#build-the-magic-ui-css-for-production)
generates it.

## Component model

Components follow the `shadcn` package (see its
["Component model"](../README.md#component-model--how-a-react-component-maps-to-go)):

| Magic UI (React)                 | this package (Go)                                    |
|----------------------------------|------------------------------------------------------|
| `<Marquee reverse repeat={2}>`   | `Marquee(MarqueeOptions{Reverse: true, Repeat: 2}, …)` |
| prop default                     | the option's zero value                              |
| prop defaulting to `true`        | an inverted option (`HidePath`)                      |
| `className`, `style`             | `html.Class(…)`, `html.Style(…)` in the variadic     |
| `useId()`                        | an `ID` option                                       |
| `Math.random()`                  | a `Seed` option                                      |
| motion animation                 | CSS keyframes from `magicui.CSS`                     |

A component without props takes only `attribsChildren ...any`. Every root has
`data-slot="<name>"`, and caller classes win over the base classes like in
`shadcn`. A caller `style` is appended after the component's own declarations,
so it wins too.

Colors, lengths and other CSS values from options are set as CSS custom
properties in `style`. A value holding `;`, a brace, an angle bracket, a
backslash, a quote or `url(` is rejected: the component renders an error
instead of the injected CSS.

## Differences from upstream

- **Reduced motion.** Under `prefers-reduced-motion: reduce` every animation
  of the package is switched off and each component shows its final state:
  text is fully visible, a marquee stands still, a progress bar shows its
  value. Upstream ignores the setting. Hover transitions stay.
- **Motion rewrites.** Components that animate with the motion library
  upstream are CSS here. Where CSS cannot match upstream exactly, the
  component's entry below says how it differs.
- **Scroll-driven animations** (`animation-timeline: scroll()`/`view()`) are
  wrapped in `@supports`. Browsers without them (Firefox, unless its flag is
  set) show the final state.
- **Seeded randomness.** Positions and timings upstream draws with
  `Math.random()` in the browser are drawn on the server from the `Seed`
  option, so the same options render the same markup and a page is cacheable.
  They do not change over time.
- **No `asChild`.** go-mx has no Slot; where upstream offers `asChild`, the
  package exports the classes instead.
- **Buttons default to `type="button"`**, like `shadcn.Button`; upstream
  leaves the type unset, so a click submits an enclosing form.
- **Ids.** The SVG components and HeroVideoDialog take an `ID` option for
  the id upstream draws with `useId()`. Left empty, it is an `mx.UniqueID`,
  which differs between renders; pass a fixed id for byte-identical output.
- **Duplicated content is hidden from assistive technology.** Copies a
  component repeats (Marquee) are `aria-hidden`, as are purely decorative
  layers.

## Components

The gallery shows each of them with its Go source:
[ungerik.github.io/go-mx/shadcn/gallery](https://ungerik.github.io/go-mx/shadcn/gallery/).
### Text effects

- **AnimatedGradientText** — `AnimatedGradientText(opts AnimatedGradientTextOptions, attribsChildren ...any)`: text filled with a gradient that keeps flowing from `ColorFrom` through `ColorTo`; `Speed` scales `--bg-size`.
- **AnimatedShinyText** — `AnimatedShinyText(opts AnimatedShinyTextOptions, attribsChildren ...any)`: muted text swept by a glare `ShimmerWidth` pixels wide. Under reduced motion the glare rests outside the text.
- **AuroraText** — `AuroraText(opts AuroraTextOptions, attribsChildren ...any)`: text filled with a swirling gradient of `Colors`, one cycle every 10s divided by `Speed`, with an sr-only copy for screen readers. Every attribute goes to the root span (upstream passes on only `className`).
- **ComicText** — `ComicText(opts ComicTextOptions, attribsChildren ...any)`: comic-book lettering (Bangers font stack, halftone fill, black outline, skew) that pops in once; the motion spring is a CSS keyframe with upstream's cubic-bezier. Any children are accepted, not only a string.
- **KineticText** — `KineticText(opts KineticTextOptions, attribsChildren ...any)`: `opts.Text` as one aria-hidden span per letter plus an sr-only copy, in the element chosen by `As`; the letter under the pointer swells to weight 900 with a stroke, by CSS transitions. The text is split into code points, not UTF-16 units.
- **LineShadowText** — `LineShadowText(opts LineShadowTextOptions, attribsChildren ...any)`: text with a moving hatched shadow in `ShadowColor`, in the element chosen by `As`. The children must be strings, because they also fill `data-text`.
- **SparklesText** — `SparklesText(opts SparklesTextOptions, attribsChildren ...any)`: bold text with stars twinkling at positions drawn from `Seed`. The sparkles keep their places (upstream moves each one every 5–15s with `setInterval`) and are hidden under reduced motion.
- **SpinningText** — `SpinningText(opts SpinningTextOptions, attribsChildren ...any)`: text laid out on a circle that turns endlessly. The `transition` and `variants` props are not ported.
- **Text3DFlip** — `Text3DFlip(opts Text3DFlipOptions, attribsChildren ...any)`: each letter rolls over in 3D on hover with a staggered delay (from the first, last or center letter, an index, or a random order from `Seed`). As a CSS hover transition the letters stay flipped while hovered and roll back on leave, where upstream flips once on mouseenter; motion's spring is sampled into a `linear()` easing.
- **TextAnimate** — `TextAnimate(opts TextAnimateOptions, attribsChildren ...any)`: animates text in by text, word, character or line with upstream's ten presets, segment i starting at Delay + i×Duration/segments. It plays when the page shows the text; `StartOnView` ties it to the scroll position instead. The element is a flex container, so whitespace an indenting writer adds between the segments is not rendered: align it with `justify-*`, not `text-*`. Not ported: exit animations, `once`, custom `variants` (use `SegmentClass` with an own `animate-*` utility), upstream's `aria-label`; `accessible={false}` is `NoScreenReaderText`.
- **TextReveal** — `TextReveal(attribsChildren ...any)`: a 200vh section whose sticky text turns solid word by word while it scrolls through, on a named view timeline. Without scroll-driven animations the text shows solid.
- **WordRotate** — `WordRotate(opts WordRotateOptions, attribsChildren ...any)`: cycles `Words` in an h1, sliding in from above and out below. All words share one grid cell, so the width is that of the widest; reduced motion shows the first word. The caller's attributes go to the h1, as upstream's `className`.

### Buttons

- **InteractiveHoverButton** — `InteractiveHoverButton(attribsChildren ...any)`: a pill button whose dot floods it on hover while the label slides out and a copy with an arrow slides in, by Tailwind transitions only. The children render twice, so they should not carry ids.
- **RainbowButton** — `RainbowButton(opts RainbowButtonOptions, attribsChildren ...any)`: a button whose border and glow cycle through `--color-1` to `--color-5`, in a default or outline variant and four sizes. Instead of upstream's `asChild`, put `RainbowButtonClasses(variant, size)` on a link.
- **ShimmerButton** — `ShimmerButton(opts ShimmerButtonOptions, attribsChildren ...any)`: a button with a spark of light travelling around its border; the options become the `--shimmer-color`, `--cut`, `--shimmer-button-radius`, `--speed` and `--bg` variables.
- **ShinyButton** — `ShinyButton(attribsChildren ...any)`: a bordered button swept by a diagonal shine that pops in with a springy overshoot and shrinks while pressed. The motion animation is CSS: the shine position is the registered property `--shiny-button-x` (upstream's `--x` would clash with other components), the springs are `linear()` easings, and upstream's invalid `var(--primary)/10%` colors are written as `color-mix()`, so its border shine and dark-mode glow show.

### Borders and highlights

- **BorderBeam** — `BorderBeam(opts BorderBeamOptions, attribsChildren ...any)`: a beam of light circling the border of its positioned parent, a keyframe on `offset-distance` along `offset-path: rect(…)` that honors Reverse, Delay, InitialOffset, Duration, Size, the colors and BorderWidth. The caller's attributes go to the beam, as upstream's `className`. Upstream's `transition` prop (a custom motion easing) is not ported.
- **GlareHover** — `GlareHover(opts GlareHoverOptions, attribsChildren ...any)`: a diagonal `::before` glare that sweeps across on hover, pure CSS. An `Angle` of 0 means -45; pass 360 for 0°.
- **ShineBorder** — `ShineBorder(opts ShineBorderOptions, attribsChildren ...any)`: an absolutely positioned layer drawing an animated shine along the border of its positioned parent in one or more `ShineColors`.

### Backgrounds

- **Backlight** — `Backlight(opts BacklightOptions, attribsChildren ...any)`: an SVG blur-and-saturate filter that makes images, videos and SVGs glow in their own colors.
- **DotPattern** — `DotPattern(opts DotPatternOptions, attribsChildren ...any)`: an SVG dot grid with an optional pulsing `Glow`. It repeats an SVG `<pattern>` tile instead of measuring the container and drawing every dot, so with `Glow` an 8×8 tile of seeded timings repeats.
- **GridPattern** — `GridPattern(opts GridPatternOptions, attribsChildren ...any)`: an SVG grid of lines, optionally with some cells filled; `X`/`Y` are `*float64` (nil means upstream's -1), and children go inside the `<svg>`.
- **HexagonPattern** — `HexagonPattern(opts HexagonPatternOptions, attribsChildren ...any)`: a honeycomb of flat-top or pointy-top hexagon outlines with an optional gap and dashing, optionally with some hexagons filled.
- **LightRays** — `LightRays(opts LightRaysOptions, attribsChildren ...any)`: soft rays of light swinging down from above, positioned from `Seed`. A ray's resting style is its brightest state (upstream's is `opacity-0`), so the rays stay visible under reduced motion.
- **Meteors** — `Meteors(opts MeteorsOptions, attribsChildren ...any)`: meteors falling diagonally, placed from `Seed` at percentages of the parent's width (upstream uses pixels up to `window.innerWidth`). They sit in an aria-hidden wrapper that takes the caller's attributes, so restyle them with e.g. `[&>span]:bg-sky-500`.
- **NoiseTexture** — `NoiseTexture(opts NoiseTextureOptions, attribsChildren ...any)`: a grainy, desaturated noise layer from an SVG `feTurbulence` filter.
- **ProgressiveBlur** — `ProgressiveBlur(opts ProgressiveBlurOptions, attribsChildren ...any)`: stacked, masked `backdrop-filter` layers that blur an edge more strongly towards it. A caller class can override the position class.
- **Ripple** — `Ripple(opts RippleOptions, attribsChildren ...any)`: concentric circles pulsing out from the center; children are kept before the circles (upstream drops them).
- **StripedPattern** — `StripedPattern(opts StripedPatternOptions, attribsChildren ...any)`: diagonal stripes in `currentColor`.
- **WarpBackground** — `WarpBackground(opts WarpBackgroundOptions, attribsChildren ...any)`: a 3D tunnel of grids with colored beams flying outwards around the children; the beams' delays, hues and lengths come from `Seed`.

### Motion and layout

- **AnimatedList** — `AnimatedList(opts AnimatedListOptions, attribsChildren ...any)`: shows its children one after another, newest on top, each springing in after `Delay` × its index while the items below make room. All items are rendered at once; upstream's endless append loop needs a script. The gap is `--animated-list-gap` (1rem).
- **BlurFade** — `BlurFade(opts BlurFadeOptions, attribsChildren ...any)`: fades its children in from a blur while they move `Offset` px in `Direction`. `InView` ties the entrance to the scroll position (`animation-timeline: view()`), so `Duration`/`Delay` do not apply then. No exit animation and no `variant` prop.
- **Marquee** — `Marquee(opts MarqueeOptions, attribsChildren ...any)`: scrolls its children endlessly, horizontally or vertically, repeated `Repeat` times; the copies after the first are aria-hidden. Override `--duration` (40s) and `--gap` (1rem) with a class such as `[--duration:20s]`.
- **OrbitingCircles** — `OrbitingCircles(opts OrbitingCirclesOptions, attribsChildren ...any)`: children orbit the center of the positioned parent, evenly spaced, with an optional path (`HidePath`). The caller's attributes go to every orbiting item, as upstream's; under reduced motion the items rest at their starting points.
- **ScrollProgress** — `ScrollProgress(attribsChildren ...any)`: a fixed gradient bar growing with the page scroll (`animation-timeline: scroll(root)`); hidden without scroll-driven animations and under reduced motion.

### Data and media

- **AnimatedCircularProgressBar** — `AnimatedCircularProgressBar(opts AnimatedCircularProgressBarOptions, attribsChildren ...any)`: a circular gauge driven by upstream's CSS variables with the rounded percentage in the middle. Empty colors default to `var(--primary)`/`var(--muted)`; `Max` is a `*float64` so a range can end at 0 (nil means 100), and `Min == Max` is a render error.
- **AvatarCircles** — `AvatarCircles(opts AvatarCirclesOptions, attribsChildren ...any)`: a row of overlapping avatar links plus an optional "+N" circle.
- **HeroVideoDialog** — `HeroVideoDialog(opts HeroVideoDialogOptions, attribsChildren ...any)`: a thumbnail with a play button that opens an embedded video in a modal, built on `shadcn.DialogTrigger`/`DialogContent`/`DialogClose` (native `<dialog>`: Escape, focus trap, backdrop click). All eight `HeroVideoAnimationStyle` values are ported with their exits (which need the CSS `overlay` property; Chrome has it). The spring is a 300ms ease-out, the iframe loads lazily, and an `onclose` attribute reloads it to stop playback.

### Device mockups

- **Android** — `Android(opts AndroidOptions, attribsChildren ...any)`: an Android phone as an `<svg>` with an optional image, or a video in a `<foreignObject>`. The clip-path id comes from `ID` (upstream's is fixed).
- **Iphone** — `Iphone(opts IphoneOptions, attribsChildren ...any)`: an iPhone frame with upstream's SVG paths and an optional image or video on the screen; the mask id comes from `ID`.
- **Safari** — `Safari(opts SafariOptions, attribsChildren ...any)`: a browser window with a URL in the address bar, an optional image or video, and the full or simple toolbar (`SafariMode`); the mask id comes from `ID`.
