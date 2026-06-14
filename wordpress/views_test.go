package wordpress

import (
	"strings"
	"testing"
)

func TestCommentOrphanPromotion(t *testing.T) {
	site := &Site{WXRVersion: "1.2"}
	v := site.Views(Options{})
	comments := []*Comment{
		{ID: 1, Author: "Top", Approved: true, ParentID: 0, Content: "top comment"},
		{ID: 2, Author: "Orphan", Approved: true, ParentID: 999, Content: "orphaned reply"}, // parent 999 was deleted
	}
	out := renderStr(t, v.CommentThread(comments))
	if !strings.Contains(out, "orphaned reply") {
		t.Errorf("reply to a missing parent was silently dropped: %s", out)
	}
	if !strings.Contains(out, "top comment") {
		t.Errorf("top-level comment missing: %s", out)
	}
}

func TestCommentCycleNoInfiniteLoop(t *testing.T) {
	site := &Site{WXRVersion: "1.2"}
	v := site.Views(Options{})
	// A → B → A cycle must not loop forever and must still render both.
	comments := []*Comment{
		{ID: 1, Author: "A", Approved: true, ParentID: 2, Content: "comment a"},
		{ID: 2, Author: "B", Approved: true, ParentID: 1, Content: "comment b"},
	}
	out := renderStr(t, v.CommentThread(comments))
	if !strings.Contains(out, "comment a") || !strings.Contains(out, "comment b") {
		t.Errorf("cyclic comments not both rendered: %s", out)
	}
}
