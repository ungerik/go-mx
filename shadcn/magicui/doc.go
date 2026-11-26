// Package magicui is a Go port of Magic UI (https://magicui.design), the
// animated components and effects that build on shadcn/ui's design tokens,
// rendered with go-mx and the [shadcn] package's class merging.
//
// Upstream ships React components that animate with Tailwind keyframes and
// the motion library. This port emits markup and Tailwind classes only:
// every animation is CSS, defined in [CSS], which an app adds to its Tailwind
// input after @import "tailwindcss". Without it the components render
// static. Components whose upstream animation needs a script or a
// third-party library are not ported yet; shadcn/TODOS.md lists them.
//
// Deliberate differences from upstream, all listed in README.md:
//
//   - Reduced motion: under prefers-reduced-motion: reduce every animation is
//     switched off and each component shows its final state. Upstream
//     ignores the setting.
//   - Motion rewrites: components built on the motion library upstream
//     (BorderBeam, BlurFade, TextAnimate, …) are re-expressed as CSS
//     keyframes, transitions and scroll-driven animations.
//   - Randomness: positions upstream draws with Math.random in the browser
//     are drawn on the server from a seeded generator, so the same options
//     render the same markup.
//
// Caller-supplied CSS values (colors, lengths) are validated: one holding a
// character that could end the declaration or open a url() makes the
// component return a render error instead of injecting CSS.
package magicui
