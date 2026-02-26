package scraper

// CSS selectors for Metal Archives HTML pages.
const (
	// Band page
	SelBandName    = "h1.band_name a"
	SelBandStats   = "#band_stats"
	SelBandLogo    = "a#logo"
	SelBandPhoto   = "a#photo"
	SelStatLabel   = "dt"
	SelStatValue   = "dd"

	// Lineup
	SelCurrentLineup = "#band_tab_members_current"
	SelPastLineup    = "#band_tab_members_past"
	SelLiveLineup    = "#band_tab_members_live"
	SelLineupRow     = "tr.lineupRow"
	SelLineupBands   = "tr.lineupBandsRow"
	SelMemberLink    = "a[href*=\"/artists/\"]"
	SelBandLink      = "a[href*=\"/bands/\"]"

	// Discography
	SelDiscogTable = "table.display.discog"
	SelDiscogRow   = "tbody tr"

	// Album page
	SelAlbumInfo    = "#album_info"
	SelTracklist    = "table.table_lyrics"
	SelTrackRow     = "tr.odd, tr.even"
	SelAlbumLineup  = "#album_members_lineup"
	SelAlbumCover   = "#cover"
)
