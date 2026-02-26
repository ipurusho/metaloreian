package scraper

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/imman/metaloreian/internal/models"
)

var (
	memberIDRegex = regexp.MustCompile(`/artists/[^/]+/(\d+)`)
	bandLinkRegex = regexp.MustCompile(`/bands/([^/]+)/(\d+)`)
)

// parseLineup parses a lineup section (current, past, or live).
func parseLineup(section *goquery.Selection) []models.Member {
	var members []models.Member

	section.Find(SelLineupRow).Each(func(_ int, row *goquery.Selection) {
		member := models.Member{}

		// Member name and ID from link
		link := row.Find(SelMemberLink).First()
		member.Name = strings.TrimSpace(link.Text())
		if href, exists := link.Attr("href"); exists {
			if matches := memberIDRegex.FindStringSubmatch(href); len(matches) >= 2 {
				member.MemberID, _ = strconv.ParseInt(matches[1], 10, 64)
			}
		}

		// Instrument from second td
		tds := row.Find("td")
		if tds.Length() >= 2 {
			instrument := strings.TrimSpace(tds.Eq(1).Text())
			// Clean up whitespace in instrument text
			instrument = strings.Join(strings.Fields(instrument), " ")
			member.Instrument = instrument
		}

		// Other bands from the next sibling lineupBandsRow
		nextRow := row.Next()
		if nextRow.HasClass("lineupBandsRow") {
			member.OtherBands = parseOtherBands(nextRow, member.MemberID)
		}

		if member.MemberID > 0 {
			members = append(members, member)
		}
	})

	return members
}

// parseOtherBands extracts the "See also:" band links from a lineupBandsRow.
func parseOtherBands(row *goquery.Selection, memberID int64) []models.MemberBand {
	var bands []models.MemberBand

	row.Find(SelBandLink).Each(func(_ int, link *goquery.Selection) {
		href, exists := link.Attr("href")
		if !exists {
			return
		}

		matches := bandLinkRegex.FindStringSubmatch(href)
		if len(matches) < 3 {
			return
		}

		bandID, err := strconv.ParseInt(matches[2], 10, 64)
		if err != nil {
			return
		}

		bands = append(bands, models.MemberBand{
			MemberID: memberID,
			BandID:   bandID,
			BandName: strings.TrimSpace(link.Text()),
		})
	})

	return bands
}
