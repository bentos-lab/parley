package pretty

import "testing"

func TestFormatRoundMessage_Italic(t *testing.T) {
	got := formatRoundMessage("*a*")
	want := roundBoldStyle.Render("a")
	if got != want {
		t.Fatalf("italic render mismatch: got %q want %q", got, want)
	}
}

func TestFormatRoundMessage_Bold(t *testing.T) {
	got := formatRoundMessage("**a**")
	want := roundBoldItalicStyle.Render("a")
	if got != want {
		t.Fatalf("bold render mismatch: got %q want %q", got, want)
	}
}

func TestFormatRoundMessage_PauseTags(t *testing.T) {
	got := formatRoundMessage("x <pause300> y <pause500> z <pause1000>")
	want := "x " + roundBoldStyle.Render("[pause]") + " y " + roundBoldStyle.Render("[pause]") + " z " + roundBoldStyle.Render("[pause]")
	if got != want {
		t.Fatalf("pause render mismatch: got %q want %q", got, want)
	}
}

func TestFormatRoundMessage_BracketsBold(t *testing.T) {
	got := formatRoundMessage("[note]")
	want := roundBoldStyle.Render("[note]")
	if got != want {
		t.Fatalf("bracket render mismatch: got %q want %q", got, want)
	}
}

func TestFormatRoundMessage_Mixed(t *testing.T) {
	got := formatRoundMessage("hi *a* and **b** [c]")
	want := "hi " + roundBoldStyle.Render("a") + " and " + roundBoldItalicStyle.Render("b") + " " + roundBoldStyle.Render("[c]")
	if got != want {
		t.Fatalf("mixed render mismatch: got %q want %q", got, want)
	}
}

func TestFormatRoundMessage_UnmatchedMarkers(t *testing.T) {
	got := formatRoundMessage("hello *world and **bold")
	want := "hello *world and **bold"
	if got != want {
		t.Fatalf("unmatched render mismatch: got %q want %q", got, want)
	}
}
