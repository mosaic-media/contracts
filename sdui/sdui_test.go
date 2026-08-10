package sdui_test

import (
	"encoding/json"
	"testing"

	"github.com/mosaic-media/contracts/sdui"
)

// TestActionsAreCleanPerKind pins the JSON shape of each Action kind — the
// envelope that rides inside a node's open props bag (contracts#6). Screens
// themselves are authored and exercised in the ui package; here we only assert
// the action encoding, which producers depend on being stable.
func TestActionsAreCleanPerKind(t *testing.T) {
	cases := map[string]sdui.Action{
		`{"kind":"navigate","screen":"home"}`:              sdui.Navigate("home", nil),
		`{"kind":"playPart","partId":"p1"}`:                sdui.Play("p1"),
		`{"kind":"toast","message":"hi","tone":"success"}`: sdui.Toast("hi", sdui.ToneSuccess),
		`{"kind":"invoke","mutation":"importContent"}`:     sdui.Invoke("importContent", nil),
		`{"kind":"back"}`:                                  sdui.Back(),
	}
	for want, a := range cases {
		b, err := json.Marshal(a)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if string(b) != want {
			t.Errorf("action marshalled to %s, want %s", b, want)
		}
	}
}
