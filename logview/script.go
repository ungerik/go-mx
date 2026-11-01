package logview

import (
	"strconv"

	"github.com/ungerik/go-mx/shadcn"
)

// Behavior attributes. They are deliberately independent of [Config.Prefix]:
// the prefix names appearance classes a theme styles, while these name the
// contract between the markup and the script, the way shadcn's
// data-stick-to-bottom does. One script then serves every view on a page, no
// matter what prefixes they were rendered with.
const (
	dataPrefix = "data-mx-log-"
	attrView   = dataPrefix + "view"
	attrLines  = dataPrefix + "lines"
	attrFilter = dataPrefix + "filter"
	// attrPause marks the pause button, and its value is the label the button
	// swaps to while paused — the same trick as attrStatus and attrBadge, so
	// the script holds no user-visible string of its own.
	attrPause = dataPrefix + "pause"
	// attrStatus marks the status live region, and its value is the state it
	// reports while paused.
	attrStatus   = dataPrefix + "status"
	attrBadge    = dataPrefix + "badge"
	attrMaxLines = dataPrefix + "max-lines"
	attrLevel    = dataPrefix + "level"
	// attrStream marks the element that holds the connection attributes, and
	// attrError the error sink inside it, which a reconnect empties.
	attrStream = dataPrefix + "stream"
	attrError  = dataPrefix + "error"
	// attrSearch is set on the root when [Config.SearchParam] is, and its value
	// is the query parameter the filter is sent to the stream URL as.
	attrSearch = dataPrefix + "search"

	// attrPaused is the paused state on the view root. It is what a stylesheet
	// keys off, so the button's caption is free to name the next action rather
	// than the current state — an aria-pressed toggle would have to do both and
	// would then say one thing to the eye and the opposite to a screen reader.
	attrPaused = dataPrefix + "paused"

	// attrPending marks a line that arrived while the view was paused, and
	// attrNomatch one the filter excludes. They are two attributes rather than
	// one hidden property because they are independent reasons to hide the same
	// line: resuming must not reveal a line the filter excludes. Both are
	// hidden by a rule in [Theme.CSS] instead of by the hidden property, whose
	// UA-stylesheet rule any author display declaration beats.
	attrPending = dataPrefix + "pending"
	attrNomatch = dataPrefix + "nomatch"
)

// matchHighlightName is the CSS custom highlight the filter's matches are
// registered under, styled by the ::highlight() rule [Theme.CSS] emits for
// [ClassMatch]. Like the attributes above it is fixed rather than derived from
// [Config.Prefix], so one script definition serves every view on the page.
const matchHighlightName = "mx-log-match"

// stickyThresholdPx is how far from the bottom counts as the reader having
// scrolled up. It gates both trimming and the scroll that pauses the stream.
//
// It is the scroll area's own default rather than a number of its own, so the
// two cannot drift. The area's effective threshold is at most this — it shrinks
// for a short area — which is the safe direction: scrolling can stop the stream
// only once the area has already stopped following it, never while it is still
// scrolling itself back to the bottom.
const stickyThresholdPx = shadcn.StickToBottomDefaultThresholdPx

// filterDebounceMs delays the filter pass so that typing a word walks the lines
// once instead of once per keystroke.
const filterDebounceMs = 120

// searchDebounceMs delays the reconnect a [Config.SearchParam] search makes.
// It is longer than the filter's delay because a reconnect costs a request and
// a backlog: the client-side filter gives feedback on what is already there
// while the reader is still typing, and the search replaces it once they stop.
const searchDebounceMs = 400

// logViewScript implements pausing, filtering and the line cap for a
// [Config.View]. Like shadcn's stick-to-bottom script it defines itself once
// per page behind a guard and runs the scan unguarded, so a view swapped in
// later still wires up, and it is driven entirely by the data attributes above
// so one definition serves every instance.
//
// It watches DOM mutations rather than htmx events, so it works for content
// appended by SSE, by an ordinary swap, or by any other script.
//
// The filter also highlights what it matched, through the CSS custom highlight
// API rather than by wrapping matches in elements: a log line is a tree of
// spans that htmx swapped in and two MutationObservers are watching, and
// rewriting it on every keystroke would disturb all of that. Ranges disturb
// nothing, and one of them can span the several spans a field is split into. A
// browser without the API filters without highlighting.
//
// Pausing has two entry points that share one state: the button, and scrolling
// up — a reader who scrolls back to look at something is asking for the stream
// to hold still, and having to also press a button to get that would be a
// second step for one intent. Which one paused is remembered, because it decides
// what resumes: scrolling back to the bottom undoes a pause the scroll caused,
// while a pause the button caused was asked for explicitly and only the button
// takes it back.
//
// A search ([Config.SearchParam]) reconnects by replacing the connecting
// element with an empty copy whose sse-connect carries the query. The swap goes
// through htmx.swap rather than replaceWith, because htmx's cleanup of the old
// element is what makes the SSE extension close its EventSource, and htmx's
// processing of the new one is what opens the next. The line container and the
// scroll area are then new elements, so the observer and the scroll listener
// are bound again.
var logViewScript = /*js*/ `
if(!window.mxLogView){
window.mxLogView=function(root){
if(root.mxLog)return;root.mxLog=1;
var lines,area,observer=new MutationObserver(added);
if(!bind())return;
var param=root.getAttribute('` + attrSearch + `')||'';
var filter=root.querySelector('[` + attrFilter + `]');
var pause=root.querySelector('[` + attrPause + `]');
var status=root.querySelector('[` + attrStatus + `]');
var badge=root.querySelector('[` + attrBadge + `]');
var max=parseInt(root.getAttribute('` + attrMaxLines + `'),10);if(!(max>0))max=0;
var paused=false,byScroll=false,pending=0,query='',hl=highlighter();
var pauseLabel=pause?pause.textContent:'',resumeLabel=pause?(pause.getAttribute('` + attrPause + `')||pauseLabel):'';
var pausedLabel=status?(status.getAttribute('` + attrStatus + `')||''):'';
// A line's text is read fresh rather than cached on the element: pausing marks
// lines instead of detaching them precisely so an out-of-band swap can still
// address one, and a cached copy would go on matching the text the line used to
// have. The filter pass is debounced and the highlighter walks the line anyway,
// so the read is not worth a cache that can be wrong.
function matches(el){return !query||(el.textContent||'').toLowerCase().indexOf(query)>=0;}
function mark(el,attr,on){if(on)el.setAttribute(attr,'');else el.removeAttribute(attr);}
function highlighter(){
if(!window.CSS||!CSS.highlights||typeof Highlight!=='function')return null;
if(!window.mxLogHighlight){window.mxLogHighlight=new Highlight();CSS.highlights.set('` + matchHighlightName + `',window.mxLogHighlight);}
return window.mxLogHighlight;
}
function unmark(el){
if(!hl||!el.mxRanges)return;
for(var i=0;i<el.mxRanges.length;i++)hl.delete(el.mxRanges[i]);
el.mxRanges=null;
}
// locate maps an offset in a line's concatenated text back to the text node
// holding it, which is what lets one match span the several spans a rendered
// field is split into — "status=200" is three of them.
function locate(nodes,off){
for(var i=nodes.length-1;i>=0;i--)if(nodes[i][1]<=off)return nodes[i];
return nodes[0];
}
function remark(el){
if(!hl||!query)return;
var nodes=[],text='',w=document.createTreeWalker(el,NodeFilter.SHOW_TEXT),n;
while(n=w.nextNode()){nodes.push([n,text.length]);text+=n.nodeValue;}
if(!nodes.length)return;
var hay=text.toLowerCase(),at=0,found=[];
while((at=hay.indexOf(query,at))>=0){
var r=document.createRange(),a=locate(nodes,at),b=locate(nodes,at+query.length);
r.setStart(a[0],at-a[1]);r.setEnd(b[0],at+query.length-b[1]);
hl.add(r);found.push(r);
at+=query.length;
}
if(found.length)el.mxRanges=found;
}
// refresh re-decides one line's visibility and its highlight together, so the
// two can never disagree about what the query is.
function refresh(el){
unmark(el);
mark(el,'` + attrNomatch + `',!matches(el));
remark(el);
}
function held(){return lines.querySelectorAll('[` + attrPending + `]');}
// countPending recomputes the badge from the DOM rather than counting arrivals,
// so it cannot drift from what resuming will actually reveal — neither once the
// cap has dropped some of the held lines, nor once the filter changed while
// paused. Held lines the filter excludes are left out for the same reason: they
// stay hidden when the pending marks are cleared, so promising them would be a
// count of lines that never appear.
function countPending(){
if(!paused)return;
pending=lines.querySelectorAll('[` + attrPending + `]:not([` + attrNomatch + `])').length;
badgeText();
}
function trim(){
if(!max)return;
if(paused){
// While paused the visible lines are frozen, so the cap applies to what is
// being held back instead. Without this a view left paused grows forever.
for(var h=held(),i=0;i+max<h.length;i++){unmark(h[i]);lines.removeChild(h[i]);}
return;
}
// Trimming under a reader who has scrolled up would drag the text they are
// reading upward by a line for every line that arrives. Scrolling up pauses, so
// this is usually the paused branch above; the position is read again here
// because a scroll event is dispatched after the fact, and a line can arrive in
// between.
if(!atBottom())return;
while(lines.children.length>max){unmark(lines.firstElementChild);lines.removeChild(lines.firstElementChild);}
}
function badgeText(){
if(!badge)return;
badge.textContent=pending?pending+' '+(badge.getAttribute('` + attrBadge + `')||''):'';
}
function filtered(){
query=filter?filter.value.trim().toLowerCase():'';
for(var i=0,c=lines.children;i<c.length;i++)refresh(c[i]);
countPending();
// Hiding or revealing lines moves the bottom without firing a scroll event, so
// the position this script and stick-to-bottom each decided from is stale until
// one is dispatched: clearing a filter can leave a reader hundreds of pixels
// above the bottom while both still believe they are following the stream. Both
// listeners read live geometry, so the synthetic event re-syncs them from the
// numbers a real scroll would have given.
if(area)area.dispatchEvent(new Event('scroll'));
}
// bind finds the line container and its scroll area and attaches the observer
// and the scroll listener to them, once at start and again after a search has
// replaced both.
function bind(){
lines=root.querySelector('[` + attrLines + `]');if(!lines)return false;
area=lines.parentElement;
observer.disconnect();observer.observe(lines,{childList:true});
if(area&&!area.mxLogScroll){area.mxLogScroll=1;area.addEventListener('scroll',scrolled,{passive:true});}
return true;
}
function added(records){
var n=0;
for(var i=0;i<records.length;i++){
var nodes=records[i].addedNodes;
for(var j=0;j<nodes.length;j++){
var el=nodes[j];
if(el.nodeType!==1)continue;
refresh(el);
if(paused)el.setAttribute('` + attrPending + `','');
n++;
}
}
if(!n)return;
trim();
countPending();
}
// search reconnects the stream with the filter as a query parameter, so the
// handler answers with everything that matches rather than everything recent.
// It does nothing unless the root asks for it, the stream has a URL, and htmx
// can swap: without a connection to replace, the filter stays the client-side
// filter it always is.
function search(){
var stream=root.querySelector('[` + attrStream + `]');
if(!param||!stream||!window.htmx||!htmx.swap)return;
var attr=stream.hasAttribute('sse-connect')?'sse-connect':'data-sse-connect';
var url=stream.getAttribute(attr);if(!url)return;
var u=new URL(url,document.baseURI),q=filter?filter.value.trim():'';
if((u.searchParams.get(param)||'')===q)return;
if(q)u.searchParams.set(param,q);else u.searchParams.delete(param);
var next=stream.cloneNode(true);
next.setAttribute(attr,u.origin===location.origin?u.pathname+u.search+u.hash:u.href);
// The copy starts empty: the lines it had are the old answer, and an error
// reported on the old connection is not one of the new.
next.querySelectorAll('[` + attrLines + `],[` + attrError + `]').forEach(function(el){el.textContent='';});
// A new search is a request to see its results, so it does not stay paused
// behind a badge counting them.
setPaused(false);
for(var i=0,c=lines.children;i<c.length;i++)unmark(c[i]);
htmx.swap(stream,next.outerHTML,{swapStyle:'outerHTML',transition:false});
bind();
}
function setPaused(on,fromScroll){
if(paused===on)return;
paused=on;
byScroll=on&&!!fromScroll;
// The caption names what pressing does next, the root attribute carries the
// state for the stylesheet, and the status region says the state out loud —
// which is the only one of the three that reaches a screen reader when it was
// scrolling, not the button, that paused the stream.
mark(root,'` + attrPaused + `',on);
if(pause)pause.textContent=on?resumeLabel:pauseLabel;
if(status)status.textContent=on?pausedLabel:'';
if(on)return;
for(var h=held(),i=0;i<h.length;i++)h[i].removeAttribute('` + attrPending + `');
pending=0;badgeText();
// Catching up before trimming, so that the cap is applied from the bottom the
// view has just returned to rather than from the one it is leaving.
if(area)area.scrollTop=area.scrollHeight;
trim();
}
// atBottom is the scroll position the pause and the line cap both turn on, read
// live rather than remembered: the rendered height changes for reasons that
// never reach a scroll listener.
function atBottom(){return !area||area.scrollHeight-area.scrollTop-area.clientHeight<=` + strconv.Itoa(stickyThresholdPx) + `;}
function scrolled(){
// Scrolling up is a request to read what is on screen, so it holds the stream
// the same way the button does, and scrolling back to the bottom is the request
// to stop reading. Only a pause the scroll caused resumes that way: one the
// button caused was asked for and stays until the button takes it back.
// Trimming cannot trigger either: removing lines from the top shrinks
// scrollHeight and scrollTop together, leaving the gap unchanged.
if(!atBottom())setPaused(true,true);else if(byScroll)setPaused(false);
}
if(pause)pause.addEventListener('click',function(){setPaused(!paused);});
if(filter){
var timer,searchTimer;
filter.addEventListener('input',function(){
clearTimeout(timer);timer=setTimeout(filtered,` + strconv.Itoa(filterDebounceMs) + `);
if(param){clearTimeout(searchTimer);searchTimer=setTimeout(search,` + strconv.Itoa(searchDebounceMs) + `);}
});
filtered();
// A filter the browser restored on a back navigation is searched for at once.
search();
}
};
window.mxLogViewScan=function(r){
if(!r||!r.querySelectorAll)return;
if(r.matches&&r.matches('[` + attrView + `]'))window.mxLogView(r);
r.querySelectorAll('[` + attrView + `]').forEach(window.mxLogView);
};
document.addEventListener('DOMContentLoaded',function(){window.mxLogViewScan(document);});
document.addEventListener('htmx:load',function(e){window.mxLogViewScan(e.target);});
}
window.mxLogViewScan(document);`
