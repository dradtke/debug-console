package types_test

import (
	"encoding/json"
	"testing"

	"github.com/dradtke/debug-console/types"
	"github.com/google/go-cmp/cmp"
)

func TestNewEvaluateRequest(t *testing.T) {
	req := types.NewEvaluateRequest(types.EvaluateArguments{
		Expression: "print('hello world')",
	})

	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err = json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}

	want := map[string]any{
		"seq": float64(1),
		"type": "request",
		"command": "evaluate",
		"arguments": map[string]any{
			"expression": "print('hello world')",
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}
