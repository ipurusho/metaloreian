package store

import (
	"database/sql"
	"time"

	"github.com/imman/metaloreian/internal/models"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

const cacheTTL = 7 * 24 * time.Hour

// GetBand returns a cached band or nil if not found / stale.
func (s *Store) GetBand(maID int64) (*models.Band, error) {
	var b models.Band
	err := s.db.QueryRow(`
		SELECT ma_id, name, genre, country, status, themes, formed_in, years_active, logo_url, photo_url, scraped_at
		FROM bands WHERE ma_id = $1`, maID).Scan(
		&b.MAID, &b.Name, &b.Genre, &b.Country, &b.Status, &b.Themes,
		&b.FormedIn, &b.YearsActive, &b.LogoURL, &b.PhotoURL, &b.ScrapedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &b, err
}

// IsBandFresh checks if cached data is within TTL.
func (s *Store) IsBandFresh(b *models.Band) bool {
	return time.Since(b.ScrapedAt) < cacheTTL
}

// UpsertBand inserts or updates a band record.
func (s *Store) UpsertBand(b *models.Band) error {
	_, err := s.db.Exec(`
		INSERT INTO bands (ma_id, name, genre, country, status, themes, formed_in, years_active, logo_url, photo_url, scraped_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW())
		ON CONFLICT (ma_id) DO UPDATE SET
			name=$2, genre=$3, country=$4, status=$5, themes=$6, formed_in=$7,
			years_active=$8, logo_url=$9, photo_url=$10, scraped_at=NOW()`,
		b.MAID, b.Name, b.Genre, b.Country, b.Status, b.Themes,
		b.FormedIn, b.YearsActive, b.LogoURL, b.PhotoURL,
	)
	return err
}

// UpsertAlbum inserts or updates an album record.
func (s *Store) UpsertAlbum(a *models.Album) error {
	_, err := s.db.Exec(`
		INSERT INTO albums (album_id, band_id, name, type, release_date, label, format, cover_url, scraped_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())
		ON CONFLICT (album_id) DO UPDATE SET
			band_id=$2, name=$3, type=$4, release_date=$5, label=$6, format=$7, cover_url=$8, scraped_at=NOW()`,
		a.AlbumID, a.BandID, a.Name, a.Type, a.ReleaseDate, a.Label, a.Format, a.CoverURL,
	)
	return err
}

// UpsertMember inserts or updates a member record.
func (s *Store) UpsertMember(m *models.Member) error {
	_, err := s.db.Exec(`
		INSERT INTO members (member_id, name) VALUES ($1, $2)
		ON CONFLICT (member_id) DO UPDATE SET name=$2`,
		m.MemberID, m.Name,
	)
	return err
}

// UpsertBandLineup inserts a band lineup entry.
func (s *Store) UpsertBandLineup(bandID, memberID int64, instrument, lineupType, years string) error {
	_, err := s.db.Exec(`
		INSERT INTO band_lineup (band_id, member_id, instrument, lineup_type, years)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (band_id, member_id, lineup_type) DO UPDATE SET
			instrument=$3, years=$5`,
		bandID, memberID, instrument, lineupType, years,
	)
	return err
}

// UpsertAlbumLineup inserts an album lineup entry.
func (s *Store) UpsertAlbumLineup(albumID, memberID int64, instrument string) error {
	_, err := s.db.Exec(`
		INSERT INTO album_lineup (album_id, member_id, instrument)
		VALUES ($1,$2,$3)
		ON CONFLICT (album_id, member_id) DO UPDATE SET instrument=$3`,
		albumID, memberID, instrument,
	)
	return err
}

// UpsertMemberBand inserts a member's other-band link.
func (s *Store) UpsertMemberBand(mb *models.MemberBand) error {
	_, err := s.db.Exec(`
		INSERT INTO member_bands (member_id, band_id, band_name)
		VALUES ($1,$2,$3)
		ON CONFLICT DO NOTHING`,
		mb.MemberID, mb.BandID, mb.BandName,
	)
	return err
}

// ReplaceTracks replaces all tracks for an album.
func (s *Store) ReplaceTracks(albumID int64, tracks []models.Track) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM tracks WHERE album_id = $1`, albumID)
	if err != nil {
		return err
	}

	for _, t := range tracks {
		_, err = tx.Exec(`INSERT INTO tracks (album_id, track_number, title, duration) VALUES ($1,$2,$3,$4)`,
			albumID, t.TrackNumber, t.Title, t.Duration)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetBandLineup returns members for a band by lineup type.
func (s *Store) GetBandLineup(bandID int64, lineupType string) ([]models.Member, error) {
	rows, err := s.db.Query(`
		SELECT m.member_id, m.name, bl.instrument, bl.lineup_type, bl.years
		FROM band_lineup bl
		JOIN members m ON m.member_id = bl.member_id
		WHERE bl.band_id = $1 AND bl.lineup_type = $2
		ORDER BY bl.id`, bandID, lineupType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.Member
	for rows.Next() {
		var m models.Member
		if err := rows.Scan(&m.MemberID, &m.Name, &m.Instrument, &m.LineupType, &m.Years); err != nil {
			return nil, err
		}

		// Load other bands for each member
		m.OtherBands, _ = s.GetMemberBands(m.MemberID)
		members = append(members, m)
	}
	return members, rows.Err()
}

// GetMemberBands returns the "See also" bands for a member.
func (s *Store) GetMemberBands(memberID int64) ([]models.MemberBand, error) {
	rows, err := s.db.Query(`
		SELECT id, member_id, band_id, band_name FROM member_bands WHERE member_id = $1`,
		memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bands []models.MemberBand
	for rows.Next() {
		var mb models.MemberBand
		if err := rows.Scan(&mb.ID, &mb.MemberID, &mb.BandID, &mb.BandName); err != nil {
			return nil, err
		}
		bands = append(bands, mb)
	}
	return bands, rows.Err()
}

// GetDiscography returns all albums for a band.
func (s *Store) GetDiscography(bandID int64) ([]models.Album, error) {
	rows, err := s.db.Query(`
		SELECT album_id, band_id, name, type, release_date, label, format, cover_url, scraped_at
		FROM albums WHERE band_id = $1 ORDER BY release_date`, bandID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var albums []models.Album
	for rows.Next() {
		var a models.Album
		if err := rows.Scan(&a.AlbumID, &a.BandID, &a.Name, &a.Type, &a.ReleaseDate, &a.Label, &a.Format, &a.CoverURL, &a.ScrapedAt); err != nil {
			return nil, err
		}
		albums = append(albums, a)
	}
	return albums, rows.Err()
}

// GetAlbum returns a cached album or nil.
func (s *Store) GetAlbum(albumID int64) (*models.Album, error) {
	var a models.Album
	err := s.db.QueryRow(`
		SELECT album_id, band_id, name, type, release_date, label, format, cover_url, scraped_at
		FROM albums WHERE album_id = $1`, albumID).Scan(
		&a.AlbumID, &a.BandID, &a.Name, &a.Type, &a.ReleaseDate, &a.Label, &a.Format, &a.CoverURL, &a.ScrapedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &a, err
}

// IsAlbumFresh checks if cached album data is within TTL.
func (s *Store) IsAlbumFresh(a *models.Album) bool {
	return time.Since(a.ScrapedAt) < cacheTTL
}

// GetAlbumTracks returns all tracks for an album.
func (s *Store) GetAlbumTracks(albumID int64) ([]models.Track, error) {
	rows, err := s.db.Query(`
		SELECT id, album_id, track_number, title, duration FROM tracks WHERE album_id = $1 ORDER BY track_number`, albumID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []models.Track
	for rows.Next() {
		var t models.Track
		if err := rows.Scan(&t.ID, &t.AlbumID, &t.TrackNumber, &t.Title, &t.Duration); err != nil {
			return nil, err
		}
		tracks = append(tracks, t)
	}
	return tracks, rows.Err()
}

// GetAlbumLineup returns lineup for a specific album.
func (s *Store) GetAlbumLineup(albumID int64) ([]models.Member, error) {
	rows, err := s.db.Query(`
		SELECT m.member_id, m.name, al.instrument
		FROM album_lineup al
		JOIN members m ON m.member_id = al.member_id
		WHERE al.album_id = $1
		ORDER BY al.id`, albumID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.Member
	for rows.Next() {
		var m models.Member
		if err := rows.Scan(&m.MemberID, &m.Name, &m.Instrument); err != nil {
			return nil, err
		}
		m.OtherBands, _ = s.GetMemberBands(m.MemberID)
		members = append(members, m)
	}
	return members, rows.Err()
}

// SearchBandsByName performs a trigram search on band names.
func (s *Store) SearchBandsByName(query string) ([]models.BandSearchResult, error) {
	rows, err := s.db.Query(`
		SELECT ma_id, name, genre, country FROM bands
		WHERE name % $1 OR name ILIKE '%' || $1 || '%'
		ORDER BY similarity(name, $1) DESC
		LIMIT 20`, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.BandSearchResult
	for rows.Next() {
		var r models.BandSearchResult
		if err := rows.Scan(&r.MAID, &r.Name, &r.Genre, &r.Country); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
