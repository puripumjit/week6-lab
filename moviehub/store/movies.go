package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// Movie โครงสร้างข้อมูลพื้นฐานของภาพยนตร์
type Movie struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	TitleTh     string  `json:"title_th"`
	Genre       string  `json:"genre"`
	Year        int     `json:"year"`
	Rating      float64 `json:"rating"`
	Poster      *string `json:"poster"`
	Backdrop    *string `json:"backdrop,omitempty"`
	ReleaseDate *string `json:"release_date,omitempty"`
	VoteCount   int     `json:"vote_count,omitempty"`
	Popularity  float64 `json:"popularity,omitempty"`
}

// MovieDetail คือหนังหนึ่งเรื่องพร้อมเรื่องย่อ ใช้กับหน้ารายละเอียด
type MovieDetail struct {
	Movie
	Detail *string `json:"detail"`
}

// SearchMovies ค้นหารายชื่อภาพยนตร์ตามชื่อเรื่อง
func (s *Store) SearchMovies(ctx context.Context, query string) ([]Movie, error) {
	const q = `
		SELECT id,
		       original_title,
		       COALESCE(title, original_title),
		       poster_path,
		       release_date::text,
		       vote_average,
		       vote_count,
		       popularity
		FROM public.movies
		WHERE original_title ILIKE '%' || $1 || '%'
		   OR title ILIKE '%' || $1 || '%'
		ORDER BY popularity DESC
		LIMIT 20`

	rows, err := s.Pool.Query(ctx, q, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movies := []Movie{}
	for rows.Next() {
		var m Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.TitleTh, &m.Poster,
			&m.ReleaseDate, &m.Rating, &m.VoteCount, &m.Popularity); err != nil {
			return nil, err
		}
		movies = append(movies, m)
	}
	return movies, rows.Err()
}

// GetMovie ดึงหนังหนึ่งเรื่องพร้อมเรื่องย่อ
func (s *Store) GetMovie(ctx context.Context, id int64) (*MovieDetail, error) {
	const q = `
		SELECT id,
		       original_title,
		       COALESCE(title, original_title),
		       poster_path,
		       backdrop_path,
		       release_date::text,
		       vote_average,
		       vote_count,
		       popularity,
		       COALESCE(overview_th, overview)
		FROM public.movies
		WHERE id = $1`

	var d MovieDetail
	err := s.Pool.QueryRow(ctx, q, id).Scan(&d.ID, &d.Title, &d.TitleTh,
		&d.Poster, &d.Backdrop, &d.ReleaseDate, &d.Rating,
		&d.VoteCount, &d.Popularity, &d.Detail)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}