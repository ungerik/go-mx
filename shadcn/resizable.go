package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/svg"
)

// ResizeDirection selects whether a [ResizablePanelGroup] lays its panels out
// in a row (horizontal, the default for "") or a column (vertical).
type ResizeDirection string

const (
	// ResizeHorizontal lays the panels out in a row, resized along the x-axis (the default).
	ResizeHorizontal ResizeDirection = "horizontal"
	// ResizeVertical lays the panels out in a column, resized along the y-axis.
	ResizeVertical ResizeDirection = "vertical"
)

// resizeScript drags a [ResizableHandle] to resize its two adjacent panels by
// adjusting their flex-basis. shadcn wraps react-resizable-panels; CSS has no
// declarative split-pane resize, so this one small pointer-drag handler is the
// native equivalent (the same tradeoff as the Slider's drag script). A 24px
// minimum keeps either panel from collapsing to nothing.
const resizeScript = /*js*/ `if(!window.resizeStart){window.resizeStart=function(e,h){e.preventDefault();var g=h.parentElement;var horiz=g.dataset.direction!=='vertical';var prev=h.previousElementSibling,next=h.nextElementSibling;if(!prev||!next)return;var start=horiz?e.clientX:e.clientY;var ps=horiz?prev.offsetWidth:prev.offsetHeight;var ns=horiz?next.offsetWidth:next.offsetHeight;function mv(ev){var d=(horiz?ev.clientX:ev.clientY)-start;var np=ps+d,nn=ns-d;if(np<24||nn<24)return;prev.style.flex='1 1 '+np+'px';next.style.flex='1 1 '+nn+'px';}function up(){document.removeEventListener('pointermove',mv);document.removeEventListener('pointerup',up);}document.addEventListener('pointermove',mv);document.addEventListener('pointerup',up);};}`

// ResizablePanelGroup is the flex container of [ResizablePanel]s and
// [ResizableHandle]s. The shared resizeScript is appended once. direction may be
// "" for the default (horizontal).
func ResizablePanelGroup(direction ResizeDirection, attribsChildren ...any) *mx.Element {
	dir := "horizontal"
	if direction == ResizeVertical {
		dir = "vertical"
	}
	e := html.Div(attribsChildren...)
	if e.AttribIndex("data-direction") < 0 {
		e.Attribs = append(e.Attribs, html.DataAttr("direction", dir))
	}
	e.Children = append(e.Children, html.Script(mx.Raw(resizeScript)))
	return finish(e, "resizable-panel-group",
		"group/resizable flex h-full w-full data-[direction=vertical]:flex-col")
}

// ResizablePanel is one resizable region. It defaults to flex: 1 1 0 (equal
// share); pass html.Style("flex: 1 1 30%") to set a different initial size.
func ResizablePanel(attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex("style") < 0 {
		e.Attribs = append(e.Attribs, html.Style("flex: 1 1 0"))
	}
	return finish(e, "resizable-panel", "overflow-hidden")
}

// ResizableHandle is the draggable divider between two panels. It carries a
// grip and adapts its orientation/cursor to the group direction. Default
// onpointerdown starts the resize drag.
func ResizableHandle(attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex("onpointerdown") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("onpointerdown", "resizeStart(event,this)"))
	}
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("separator"))
	}
	if len(e.Children) == 0 {
		e.Children = mx.Components{
			html.DivClass("bg-border z-10 flex h-4 w-3 items-center justify-center rounded-xs border group-data-[direction=vertical]/resizable:h-3 group-data-[direction=vertical]/resizable:w-4 group-data-[direction=vertical]/resizable:rotate-90",
				icon("grip-vertical", "size-2.5",
					svg.Circle(svg.CX(9), svg.CY(12), svg.R(1)), svg.Circle(svg.CX(9), svg.CY(5), svg.R(1)), svg.Circle(svg.CX(9), svg.CY(19), svg.R(1)),
					svg.Circle(svg.CX(15), svg.CY(12), svg.R(1)), svg.Circle(svg.CX(15), svg.CY(5), svg.R(1)), svg.Circle(svg.CX(15), svg.CY(19), svg.R(1))),
			),
		}
	}
	return finish(e, "resizable-handle",
		"bg-border relative flex w-px shrink-0 cursor-col-resize items-center justify-center group-data-[direction=vertical]/resizable:h-px group-data-[direction=vertical]/resizable:w-full group-data-[direction=vertical]/resizable:cursor-row-resize")
}
