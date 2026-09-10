package logview

import "strconv"

// Behavior attributes. They are deliberately independent of [Config.Prefix]:
// the prefix names appearance classes a theme styles, while these name the
// contract between the markup and the script, the way shadcn's
// data-stick-to-bottom does. One script then serves every view on a page, no
// matter what prefixes they were rendered with.
const (
	dataPrefix   = "data-mx-log-"
	attrView     = dataPrefix + "view"
	attrLines    = dataPrefix + "lines"
	attrFilter   = dataPrefix + "filter"
	attrPause    = dataPrefix + "pause"
	attrBadge    = dataPrefix + "badge"
	attrMaxLines = dataPrefix + "max-lines"
	attrLevel    = dataPrefix + "level"

	// attrPending marks a line that arrived while the view was paused, and
	// attrNomatch one the filter excludes. They are two attributes rather than
	// one hidden property because they are independent reasons to hide the same
	// line: resuming must not reveal a line the filter excludes. Both are
	// hidden by a rule in [Theme.CSS] instead of by the hidden property, whose
	// UA-stylesheet rule any author display declaration beats.
	attrPending = dataPrefix + "pending"
	attrNomatch = dataPrefix + "nomatch"
)

// stickyThresholdPx is how close to the bottom counts as following the stream.
// It only gates trimming, so it does not have to match the scroll area's own
// threshold.
const stickyThresholdPx = 32

// filterDebounceMs delays the filter pass so that typing a word walks the lines
// once instead of once per keystroke.
const filterDebounceMs = 120

// logViewScript implements pausing, filtering and the line cap for a
// [Config.View]. Like shadcn's stick-to-bottom script it defines itself once
// per page behind a guard and runs the scan unguarded, so a view swapped in
// later still wires up, and it is driven entirely by the data attributes above
// so one definition serves every instance.
//
// It watches DOM mutations rather than htmx events, so it works for content
// appended by SSE, by an ordinary swap, or by any other script.
var logViewScript = /*js*/ `
if(!window.mxLogView){
window.mxLogView=function(root){
if(root.mxLog)return;root.mxLog=1;
var lines=root.querySelector('[` + attrLines + `]');if(!lines)return;
var area=lines.parentElement;
var filter=root.querySelector('[` + attrFilter + `]');
var pause=root.querySelector('[` + attrPause + `]');
var badge=root.querySelector('[` + attrBadge + `]');
var max=parseInt(root.getAttribute('` + attrMaxLines + `'),10);if(!(max>0))max=0;
var paused=false,pending=0,query='',stuck=true;
function lower(el){return el.mxLc||(el.mxLc=(el.textContent||'').toLowerCase());}
function matches(el){return !query||lower(el).indexOf(query)>=0;}
function mark(el,attr,on){if(on)el.setAttribute(attr,'');else el.removeAttribute(attr);}
function held(){return lines.querySelectorAll('[` + attrPending + `]');}
function trim(){
if(!max)return;
if(paused){
// While paused the visible lines are frozen, so the cap applies to what is
// being held back instead. Without this a view left paused grows forever.
for(var h=held(),i=0;i+max<h.length;i++)lines.removeChild(h[i]);
return;
}
if(!stuck)return;
while(lines.children.length>max)lines.removeChild(lines.firstElementChild);
}
function badgeText(){
if(!badge)return;
badge.textContent=pending?pending+' '+(badge.getAttribute('` + attrBadge + `')||''):'';
mark(badge,'hidden',pending===0);
}
function filtered(){
query=filter?filter.value.trim().toLowerCase():'';
for(var i=0,c=lines.children;i<c.length;i++)mark(c[i],'` + attrNomatch + `',!matches(c[i]));
}
new MutationObserver(function(records){
var added=0;
for(var i=0;i<records.length;i++){
var nodes=records[i].addedNodes;
for(var j=0;j<nodes.length;j++){
var el=nodes[j];
if(el.nodeType!==1)continue;
mark(el,'` + attrNomatch + `',!matches(el));
if(paused)el.setAttribute('` + attrPending + `','');
added++;
}
}
if(!added)return;
trim();
// Derived from the DOM rather than counted, so the badge cannot drift from
// what resuming will actually reveal once the cap has dropped some.
if(paused){pending=held().length;badgeText();}
}).observe(lines,{childList:true});
if(area)area.addEventListener('scroll',function(){
stuck=area.scrollHeight-area.scrollTop-area.clientHeight<=` + strconv.Itoa(stickyThresholdPx) + `;
},{passive:true});
if(pause)pause.addEventListener('click',function(){
paused=!paused;
pause.setAttribute('aria-pressed',paused?'true':'false');
if(paused)return;
for(var h=held(),i=0;i<h.length;i++)h[i].removeAttribute('` + attrPending + `');
pending=0;badgeText();stuck=true;trim();
if(area)area.scrollTop=area.scrollHeight;
});
if(filter){
var timer;
filter.addEventListener('input',function(){
clearTimeout(timer);timer=setTimeout(filtered,` + strconv.Itoa(filterDebounceMs) + `);
});
filtered();
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
