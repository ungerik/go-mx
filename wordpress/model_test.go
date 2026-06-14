package wordpress

import (
	"bytes"
	"encoding/json"
	"testing"
)

// TestModelJSONRoundTrip guards the model's headline property: a Site marshals
// to a JSON tree (no parent/child pointer cycles) and round-trips unchanged.
func TestModelJSONRoundTrip(t *testing.T) {
	site, _, err := ParseFile("testdata/sample-wxr.xml")
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	first, err := json.Marshal(site)
	if err != nil {
		t.Fatalf("marshal: %v (a pointer cycle would surface here)", err)
	}

	var back Site
	if err := json.Unmarshal(first, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	second, err := json.Marshal(&back)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}

	if !bytes.Equal(first, second) {
		t.Errorf("JSON not stable across round-trip:\n first=%s\nsecond=%s", first, second)
	}
}
