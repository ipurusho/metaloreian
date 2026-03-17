package scraper

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestParseLineup(t *testing.T) {
	html := `<div id="lineup">
	<table>
		<tr class="lineupRow">
			<td><a href="https://www.metal-archives.com/artists/Mikael_%C3%85kerfeldt/1234">Mikael Åkerfeldt</a></td>
			<td>Vocals, Guitars  (1990-present) </td>
		</tr>
		<tr class="lineupBandsRow">
			<td colspan="2">
				See also: <a href="https://www.metal-archives.com/bands/Bloodbath/123">Bloodbath</a>,
				<a href="https://www.metal-archives.com/bands/Storm_Corrosion/456">Storm Corrosion</a>
			</td>
		</tr>
		<tr class="lineupRow">
			<td><a href="https://www.metal-archives.com/artists/Fredrik_%C3%85kesson/5678">Fredrik Åkesson</a></td>
			<td>Guitars (2007-present) </td>
		</tr>
	</table>
	</div>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("parse html: %v", err)
	}

	members := parseLineup(doc.Find("#lineup"))

	if len(members) != 2 {
		t.Fatalf("got %d members, want 2", len(members))
	}

	// First member
	if members[0].Name != "Mikael Åkerfeldt" {
		t.Errorf("member 0 name = %q", members[0].Name)
	}
	if members[0].MemberID != 1234 {
		t.Errorf("member 0 ID = %d, want 1234", members[0].MemberID)
	}
	if !strings.Contains(members[0].Instrument, "Vocals") {
		t.Errorf("member 0 instrument = %q, expected to contain Vocals", members[0].Instrument)
	}
	if len(members[0].OtherBands) != 2 {
		t.Errorf("member 0 other bands = %d, want 2", len(members[0].OtherBands))
	}

	// Second member
	if members[1].Name != "Fredrik Åkesson" {
		t.Errorf("member 1 name = %q", members[1].Name)
	}
	if members[1].MemberID != 5678 {
		t.Errorf("member 1 ID = %d, want 5678", members[1].MemberID)
	}
}

func TestParseOtherBands(t *testing.T) {
	html := `<html><body><table><tbody>
	<tr class="lineupBandsRow">
		<td colspan="2">
			See also: <a href="https://www.metal-archives.com/bands/Bloodbath/123">Bloodbath</a>,
			<a href="https://www.metal-archives.com/bands/Katatonia/456">Katatonia</a>,
			<a href="https://www.metal-archives.com/bands/Storm_Corrosion/789">Storm Corrosion</a>
		</td>
	</tr>
	</tbody></table></body></html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	row := doc.Find("tr.lineupBandsRow").First()

	bands := parseOtherBands(row, 42)

	if len(bands) != 3 {
		t.Fatalf("got %d bands, want 3", len(bands))
	}

	expected := []struct {
		id   int64
		name string
	}{
		{123, "Bloodbath"},
		{456, "Katatonia"},
		{789, "Storm Corrosion"},
	}

	for i, e := range expected {
		if bands[i].BandID != e.id {
			t.Errorf("band %d: id = %d, want %d", i, bands[i].BandID, e.id)
		}
		if bands[i].BandName != e.name {
			t.Errorf("band %d: name = %q, want %q", i, bands[i].BandName, e.name)
		}
		if bands[i].MemberID != 42 {
			t.Errorf("band %d: memberID = %d, want 42", i, bands[i].MemberID)
		}
	}
}

func TestParseOtherBands_NoLinks(t *testing.T) {
	html := `<html><body><table><tbody><tr class="lineupBandsRow"><td>See also: N/A</td></tr></tbody></table></body></html>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	row := doc.Find("tr.lineupBandsRow").First()

	bands := parseOtherBands(row, 1)
	if len(bands) != 0 {
		t.Errorf("got %d bands, want 0", len(bands))
	}
}

func TestParseLineup_Empty(t *testing.T) {
	html := `<div id="lineup"><table></table></div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))

	members := parseLineup(doc.Find("#lineup"))
	if len(members) != 0 {
		t.Errorf("got %d members, want 0", len(members))
	}
}
