package scraper

import "testing"

func TestExtractLinkText(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			"standard link",
			`<a href="https://www.metal-archives.com/bands/Metallica/125">Metallica</a>`,
			"Metallica",
		},
		{
			"link with extra attributes",
			`<a href="/bands/Opeth/482" class="bold">Opeth</a>`,
			"Opeth",
		},
		{
			"no closing tag",
			`<a href="/bands/Test/1">Just text`,
			"Just text",
		},
		{
			"no anchor tag at all",
			"plain text",
			"plain text",
		},
		{
			"empty link text",
			`<a href="/bands/X/1"></a>`,
			"",
		},
		{
			"link with whitespace",
			`<a href="/bands/Darkthrone/69">  Darkthrone  </a>`,
			"Darkthrone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractLinkText(tt.html)
			if got != tt.want {
				t.Errorf("extractLinkText(%q) = %q, want %q", tt.html, got, tt.want)
			}
		})
	}
}

func TestBandIDRegex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantID  string
		wantOK  bool
	}{
		{
			"standard band URL",
			`/bands/Metallica/125`,
			"125",
			true,
		},
		{
			"URL with full domain",
			`https://www.metal-archives.com/bands/Death/141`,
			"141",
			true,
		},
		{
			"band name with special chars",
			`/bands/AC%2FDC/105`,
			"105",
			true,
		},
		{
			"no match",
			`/artists/someone/123`,
			"",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := bandIDRegex.FindStringSubmatch(tt.input)
			if tt.wantOK {
				if len(matches) < 2 {
					t.Fatalf("expected match, got none")
				}
				if matches[1] != tt.wantID {
					t.Errorf("band ID = %q, want %q", matches[1], tt.wantID)
				}
			} else {
				if len(matches) >= 2 {
					t.Errorf("expected no match, got %v", matches)
				}
			}
		})
	}
}
