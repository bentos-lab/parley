package native_test

import (
	"testing"

	"github.com/bentos-lab/parley/adapter/outbound/tts/native"
)

func TestNormalizeText(t *testing.T) {
	input := "<audio>Hello [world] **!!!**</audio>\n\nLine two."
	want := "Hello !!! Line two."

	got := native.NormalizeText(input)
	if got != want {
		t.Fatalf("NormalizeText() = %q, want %q", got, want)
	}
}
