package scraper

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestParseTracks(t *testing.T) {
	html := `<html><body>
	<table class="table_lyrics">
	<tbody>
		<tr class="odd">
			<td>1.</td>
			<td>The Leper Affinity</td>
			<td>10:23</td>
			<td></td>
		</tr>
		<tr class="displayNone">
			<td colspan="4">Lyrics hidden here</td>
		</tr>
		<tr class="even">
			<td>2.</td>
			<td>Bleak</td>
			<td>09:18</td>
			<td></td>
		</tr>
		<tr class="sideRow">
			<td colspan="4">Side B</td>
		</tr>
		<tr class="odd">
			<td>3.</td>
			<td>Harvest</td>
			<td>06:01</td>
			<td></td>
		</tr>
	</tbody>
	</table>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("parse html: %v", err)
	}

	tracks := parseTracks(doc)

	if len(tracks) != 3 {
		t.Fatalf("got %d tracks, want 3", len(tracks))
	}

	tests := []struct {
		num   int
		title string
		dur   string
	}{
		{1, "The Leper Affinity", "10:23"},
		{2, "Bleak", "09:18"},
		{3, "Harvest", "06:01"},
	}

	for i, tt := range tests {
		if tracks[i].TrackNumber != tt.num {
			t.Errorf("track %d: number = %d, want %d", i, tracks[i].TrackNumber, tt.num)
		}
		if tracks[i].Title != tt.title {
			t.Errorf("track %d: title = %q, want %q", i, tracks[i].Title, tt.title)
		}
		if tracks[i].Duration != tt.dur {
			t.Errorf("track %d: duration = %q, want %q", i, tracks[i].Duration, tt.dur)
		}
	}
}

func TestParseTracks_Empty(t *testing.T) {
	html := `<html><body><table class="table_lyrics"><tbody></tbody></table></body></html>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))

	tracks := parseTracks(doc)
	if len(tracks) != 0 {
		t.Errorf("got %d tracks, want 0", len(tracks))
	}
}

func TestParseTracks_SkipsInvalidRows(t *testing.T) {
	html := `<html><body>
	<table class="table_lyrics">
	<tbody>
		<tr class="odd">
			<td>not-a-number.</td>
			<td>Bad Track</td>
			<td>00:00</td>
			<td></td>
		</tr>
		<tr class="even">
			<td>1.</td>
			<td>Good Track</td>
			<td>03:45</td>
			<td></td>
		</tr>
	</tbody>
	</table>
	</body></html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	tracks := parseTracks(doc)

	if len(tracks) != 1 {
		t.Fatalf("got %d tracks, want 1", len(tracks))
	}
	if tracks[0].Title != "Good Track" {
		t.Errorf("title = %q, want Good Track", tracks[0].Title)
	}
}
