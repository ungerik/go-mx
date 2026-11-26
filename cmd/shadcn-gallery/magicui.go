package main

import magicuiexamples "github.com/ungerik/go-mx/cmd/shadcn-gallery/examples/magicui"

// magicUIGroup is the ComponentDoc.Group of the Magic UI pages.
const magicUIGroup = "Magic UI"

// magicUIDocs is the Magic UI part of the catalog (package shadcn/magicui),
// alphabetical like the shadcn/ui pages, with descriptions transcribed from
// the Magic UI registry.
func magicUIDocs() []ComponentDoc {
	docs := []ComponentDoc{
		{
			Slug:        "android",
			Title:       "Android",
			Description: "A mockup of an Android device.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.AndroidDemo},
				{Name: "Video", Func: magicuiexamples.AndroidVideo},
				{Name: "Empty", Func: magicuiexamples.AndroidEmpty},
			},
		},
		{
			Slug:        "animated-circular-progress-bar",
			Title:       "Animated Circular Progress Bar",
			Description: "Animated Circular Progress Bar is a component that displays a circular gauge with a percentage value.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.AnimatedCircularProgressBarDemo},
			},
		},
		{
			Slug:        "animated-gradient-text",
			Title:       "Animated Gradient Text",
			Description: "An animated gradient background which transitions between colors for text.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.AnimatedGradientTextDemo},
				{Name: "Fast", Func: magicuiexamples.AnimatedGradientTextFast},
			},
		},
		{
			Slug:        "animated-list",
			Title:       "Animated List",
			Description: "A list that animates each item in sequence with a delay. Used to showcase notifications or events on your landing page.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.AnimatedListDemo},
			},
		},
		{
			Slug:        "animated-shiny-text",
			Title:       "Animated Shiny Text",
			Description: "A light glare effect which pans across text making it appear as if it is shimmering.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.AnimatedShinyTextDemo},
			},
		},
		{
			Slug:        "aurora-text",
			Title:       "Aurora Text",
			Description: "A beautiful aurora text effect.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.AuroraTextDemo},
			},
		},
		{
			Slug:        "avatar-circles",
			Title:       "Avatar Circles",
			Description: "Overlapping circles of avatars.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.AvatarCirclesDemo},
			},
		},
		{
			Slug:        "backlight",
			Title:       "Backlight",
			Description: "A backlight glow effect for videos, images, and SVGs.",
			Examples: []Example{
				{Name: "Image", Func: magicuiexamples.BacklightImage},
				{Name: "SVG", Func: magicuiexamples.BacklightSVG},
			},
		},
		{
			Slug:        "blur-fade",
			Title:       "Blur Fade",
			Description: "Blur fade in and out animation. Used to smoothly fade in and out content.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.BlurFadeDemo},
				{Name: "On Load", Func: magicuiexamples.BlurFadeOnLoad},
			},
		},
		{
			Slug:        "border-beam",
			Title:       "Border Beam",
			Description: "An animated beam of light which travels along the border of its container.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.BorderBeamDemo},
				{Name: "Reverse", Func: magicuiexamples.BorderBeamReverse},
				{Name: "Music Player", Func: magicuiexamples.BorderBeamMusicPlayer},
			},
		},
		{
			Slug:        "comic-text",
			Title:       "Comic Text",
			Description: "Comic text animation.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.ComicTextDemo},
			},
		},
		{
			Slug:        "dot-pattern",
			Title:       "Dot Pattern",
			Description: "A background dot pattern made with SVGs, fully customizable using Tailwind CSS.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.DotPatternDemo},
				{Name: "Linear Gradient", Func: magicuiexamples.DotPatternLinearGradient},
				{Name: "With Glow Effect", Func: magicuiexamples.DotPatternWithGlowEffect},
			},
		},
		{
			Slug:        "glare-hover",
			Title:       "Glare Hover",
			Description: "A diagonal glare on hover using a ::before gradient and CSS variables (angle, size, duration, color).",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.GlareHoverDemo},
				{Name: "CTA", Func: magicuiexamples.GlareHoverCTA},
				{Name: "Alerts", Func: magicuiexamples.GlareHoverAlerts},
			},
		},
		{
			Slug:        "grid-pattern",
			Title:       "Grid Pattern",
			Description: "A background grid pattern made with SVGs, fully customizable using Tailwind CSS.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.GridPatternDemo},
				{Name: "Linear Gradient", Func: magicuiexamples.GridPatternLinearGradient},
				{Name: "Dashed", Func: magicuiexamples.GridPatternDashed},
			},
		},
		{
			Slug:        "hero-video-dialog",
			Title:       "Hero Video Dialog",
			Description: "A hero video dialog component.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.HeroVideoDialogDemo},
				{Name: "Top In Bottom Out", Func: magicuiexamples.HeroVideoDialogTopInBottomOut},
			},
		},
		{
			Slug:        "hexagon-pattern",
			Title:       "Hexagon Pattern",
			Description: "A background hexagon pattern made with SVGs, fully customizable using Tailwind CSS.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.HexagonPatternDemo},
				{Name: "Dashed", Func: magicuiexamples.HexagonPatternDashed},
				{Name: "Spacing", Func: magicuiexamples.HexagonPatternSpacing},
			},
		},
		{
			Slug:        "interactive-hover-button",
			Title:       "Interactive Hover Button",
			Description: "A button whose dot fills it on hover while an arrow slides in.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.InteractiveHoverButtonDemo},
			},
		},
		{
			Slug:        "iphone",
			Title:       "iPhone",
			Description: "A mockup of the iPhone.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.IphoneDemo},
				{Name: "Video", Func: magicuiexamples.IphoneVideo},
				{Name: "Empty", Func: magicuiexamples.IphoneEmpty},
			},
		},
		{
			Slug:        "kinetic-text",
			Title:       "Kinetic Text",
			Description: "A text component that animates font weight of characters on hover.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.KineticTextDemo},
			},
		},
		{
			Slug:        "light-rays",
			Title:       "Light Rays",
			Description: "A component with animated light rays which shine down from above.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.LightRaysDemo},
			},
		},
		{
			Slug:        "line-shadow-text",
			Title:       "Line Shadow Text",
			Description: "A text component with a moving line shadow.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.LineShadowTextDemo},
			},
		},
		{
			Slug:        "marquee",
			Title:       "Marquee",
			Description: "An infinite scrolling component that can be used to display text, images, or videos.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.MarqueeDemo},
				{Name: "Vertical", Func: magicuiexamples.MarqueeVertical},
			},
		},
		{
			Slug:        "meteors",
			Title:       "Meteors",
			Description: "A meteor shower effect.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.MeteorsDemo},
			},
		},
		{
			Slug:        "noise-texture",
			Title:       "Noise Texture",
			Description: "An SVG fractal noise layer using feTurbulence, desaturation, and contrast controls for subtle texture overlays.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.NoiseTextureDemo},
				{Name: "Card", Func: magicuiexamples.NoiseTextureCard},
				{Name: "Button", Func: magicuiexamples.NoiseTextureButton},
			},
		},
		{
			Slug:        "orbiting-circles",
			Title:       "Orbiting Circles",
			Description: "A collection of circles which move in orbit along a circular path.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.OrbitingCirclesDemo},
			},
		},
		{
			Slug:        "progressive-blur",
			Title:       "Progressive Blur",
			Description: "The Progressive Blur component adds a smooth blur gradient effect to scrollable content, indicating more content below or above.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.ProgressiveBlurDemo},
			},
		},
		{
			Slug:        "rainbow-button",
			Title:       "Rainbow Button",
			Description: "An animated button with a rainbow effect.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.RainbowButtonDemo},
				{Name: "Outline", Func: magicuiexamples.RainbowButtonOutline},
			},
		},
		{
			Slug:        "ripple",
			Title:       "Ripple",
			Description: "An animated ripple effect typically used behind elements to emphasize them.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.RippleDemo},
			},
		},
		{
			Slug:        "safari",
			Title:       "Safari",
			Description: "A safari browser mockup to showcase your website.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.SafariDemo},
				{Name: "Video", Func: magicuiexamples.SafariVideo},
				{Name: "Simple", Func: magicuiexamples.SafariSimple},
			},
		},
		{
			Slug:        "scroll-progress",
			Title:       "Scroll Progress",
			Description: "Animated Scroll Progress for your pages.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.ScrollProgressDemo},
			},
		},
		{
			Slug:        "shimmer-button",
			Title:       "Shimmer Button",
			Description: "A button with a shimmering light which travels around the perimeter.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.ShimmerButtonDemo},
			},
		},
		{
			Slug:        "shine-border",
			Title:       "Shine Border",
			Description: "Shine border is an animated background border effect.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.ShineBorderDemo},
				{Name: "Monochrome", Func: magicuiexamples.ShineBorderMonochrome},
			},
		},
		{
			Slug:        "shiny-button",
			Title:       "Shiny Button",
			Description: "A shiny button component with dynamic styles in the dark mode or light mode.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.ShinyButtonDemo},
			},
		},
		{
			Slug:        "sparkles-text",
			Title:       "Sparkles Text",
			Description: "A dynamic text that generates continuous sparkles with smooth transitions, perfect for highlighting text with animated stars.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.SparklesTextDemo},
			},
		},
		{
			Slug:        "spinning-text",
			Title:       "Spinning Text",
			Description: "The Spinning Text component animates text in a circular motion with customizable speed, direction, color, and transitions for dynamic and engaging effects.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.SpinningTextDemo},
				{Name: "Reverse", Func: magicuiexamples.SpinningTextReverse},
			},
		},
		{
			Slug:        "striped-pattern",
			Title:       "Striped Pattern",
			Description: "A background striped pattern made with SVGs, fully customizable using Tailwind CSS.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.StripedPatternDemo},
				{Name: "Dashed", Func: magicuiexamples.StripedPatternDashed},
				{Name: "Right", Func: magicuiexamples.StripedPatternRight},
			},
		},
		{
			Slug:        "text-3d-flip",
			Title:       "Text 3D Flip",
			Description: "A text effect that flips each letter in 3D with a staggered animation on hover.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.Text3DFlipDemo},
				{Name: "From Center", Func: magicuiexamples.Text3DFlipFromCenter},
			},
		},
		{
			Slug:        "text-animate",
			Title:       "Text Animate",
			Description: "A text animation component that animates text using a variety of different animations.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.TextAnimateDemo},
				{Name: "Blur In Heading", Func: magicuiexamples.TextAnimateBlurInHeading},
				{Name: "Slide Up By Word", Func: magicuiexamples.TextAnimateSlideUpByWord},
			},
		},
		{
			Slug:        "text-reveal",
			Title:       "Text Reveal",
			Description: "Fade in text as you scroll down the page.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.TextRevealDemo},
			},
		},
		{
			Slug:        "warp-background",
			Title:       "Warp Background",
			Description: "A card with a time warping background effect.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.WarpBackgroundDemo},
			},
		},
		{
			Slug:        "word-rotate",
			Title:       "Word Rotate",
			Description: "A vertical rotation of words.",
			Examples: []Example{
				{Name: "Demo", Func: magicuiexamples.WordRotateDemo},
			},
		},
	}
	for i := range docs {
		docs[i].Group = magicUIGroup
	}
	return docs
}
