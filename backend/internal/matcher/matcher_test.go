package matcher

import "testing"

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		name string
		input string
		want  string
	}{
		{"lowercase", "METALLICA", "metallica"},
		{"trim spaces", "  Opeth  ", "opeth"},
		{"strip the prefix", "The Black Dahlia Murder", "black dahlia murder"},
		{"remove special chars", "AC/DC", "acdc"},
		{"keep digits", "Agent Orange 77", "agent orange 77"},
		{"unicode letters preserved", "Amon Amarth", "amon amarth"},
		{"remove punctuation", "Mötley Crüe!", "mötley crüe"},
		{"empty string", "", ""},
		{"only the", "the", "the"},
		{"the with space", "the ", "the"},
		{"mixed case the prefix", "The Haunted", "haunted"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeName(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
