package models

import "time"

type Band struct {
	MAID        int64     `json:"ma_id"`
	Name        string    `json:"name"`
	Genre       string    `json:"genre"`
	Country     string    `json:"country"`
	Status      string    `json:"status"`
	Themes      string    `json:"themes"`
	FormedIn    string    `json:"formed_in"`
	YearsActive string    `json:"years_active"`
	LogoURL     string    `json:"logo_url"`
	PhotoURL    string    `json:"photo_url"`
	ScrapedAt   time.Time `json:"scraped_at"`
}

type BandSearchResult struct {
	MAID    int64  `json:"ma_id"`
	Name    string `json:"name"`
	Genre   string `json:"genre"`
	Country string `json:"country"`
}

type Album struct {
	AlbumID     int64     `json:"album_id"`
	BandID      int64     `json:"band_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	ReleaseDate string    `json:"release_date"`
	Label       string    `json:"label"`
	Format      string    `json:"format"`
	CoverURL    string    `json:"cover_url"`
	ScrapedAt   time.Time `json:"scraped_at"`
}

type Track struct {
	ID          int64  `json:"id"`
	AlbumID     int64  `json:"album_id"`
	TrackNumber int    `json:"track_number"`
	Title       string `json:"title"`
	Duration    string `json:"duration"`
}

type Member struct {
	MemberID   int64        `json:"member_id"`
	Name       string       `json:"name"`
	Instrument string       `json:"instrument,omitempty"`
	LineupType string       `json:"lineup_type,omitempty"`
	Years      string       `json:"years,omitempty"`
	OtherBands []MemberBand `json:"other_bands,omitempty"`
}

type MemberBand struct {
	ID       int64  `json:"id,omitempty"`
	MemberID int64  `json:"member_id"`
	BandID   int64  `json:"band_id"`
	BandName string `json:"band_name"`
}

type BandFull struct {
	Band
	CurrentLineup []Member `json:"current_lineup"`
	PastLineup    []Member `json:"past_lineup,omitempty"`
	Discography   []Album  `json:"discography"`
}

type AlbumFull struct {
	Album
	BandName  string   `json:"band_name"`
	Tracks    []Track  `json:"tracks"`
	Lineup    []Member `json:"lineup"`
}

type SimilarAlbum struct {
	AlbumID  int64   `json:"album_id"`
	Name     string  `json:"name"`
	BandName string  `json:"band_name"`
	Type     string  `json:"type"`
	Year     string  `json:"year"`
	CoverURL string  `json:"cover_url"`
	Score    float64 `json:"score"`
}
