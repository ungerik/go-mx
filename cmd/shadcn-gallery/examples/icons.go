package examples

import "github.com/ungerik/go-mx"

// The lucide icons of the chat/AI examples (Marker, Trace, Attachment,
// CodeBlock, Message, MessageScroller), drawn with navIcon. Circles are
// written as two arcs so that every shape is a path.

func iconFileText() mx.Component {
	return navIcon("M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z", "M14 2v4a2 2 0 0 0 2 2h4", "M10 9H8", "M16 13H8", "M16 17H8")
}

func iconSearch() mx.Component {
	return navIcon("M19 11a8 8 0 1 1-16 0 8 8 0 0 1 16 0", "m21 21-4.3-4.3")
}

func iconTerminal() mx.Component { return navIcon("m4 17 6-6-6-6", "M12 19h8") }

func iconGlobe() mx.Component {
	return navIcon("M22 12a10 10 0 1 1-20 0 10 10 0 0 1 20 0", "M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20", "M2 12h20")
}

func iconCheck() mx.Component { return navIcon("M20 6 9 17l-5-5") }

func iconClock() mx.Component {
	return navIcon("M22 12a10 10 0 1 1-20 0 10 10 0 0 1 20 0", "M12 6v6l4 2")
}

func iconChevronRight() mx.Component { return navIcon("m9 18 6-6-6-6") }

func iconX() mx.Component { return navIcon("M18 6 6 18", "m6 6 12 12") }

func iconRetry() mx.Component {
	return navIcon("M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8", "M3 3v5h5")
}

func iconImage() mx.Component {
	return navIcon("M5 3h14a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2z", "M11 9a2 2 0 1 1-4 0 2 2 0 0 1 4 0", "m21 15-3.086-3.086a2 2 0 0 0-2.828 0L6 21")
}

func iconCopy() mx.Component {
	return navIcon("M10 8h10a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H10a2 2 0 0 1-2-2V10a2 2 0 0 1 2-2z", "M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2")
}
