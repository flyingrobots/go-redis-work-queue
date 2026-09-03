package collaborativesession

import (
	"encoding/json"
	"testing"
)

func TestColorJSONUsesIndependentChannels(t *testing.T) {
	color := Color{R: 17, G: 34, B: 51}

	encoded, err := json.Marshal(color)
	if err != nil {
		t.Fatalf("marshal Color: %v", err)
	}

	const want = `{"r":17,"g":34,"b":51,"is_default":false}`
	if string(encoded) != want {
		t.Fatalf("marshal Color = %s, want %s", encoded, want)
	}

	var decoded Color
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal Color: %v", err)
	}
	if decoded != color {
		t.Fatalf("unmarshal Color = %+v, want %+v", decoded, color)
	}
}
