package bookmarks

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/hugojanuario/devdocs/internal/database"
	"github.com/hugojanuario/devdocs/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Saved struct {
	ID         string       `json:"id"`
	Question   string       `json:"question"`
	Title      string       `json:"title"`
	Technology string       `json:"technology"`
	Response   model.Answer `json:"response"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}
type Summary struct {
	ID         string    `json:"id"`
	Question   string    `json:"question"`
	Title      string    `json:"title"`
	Technology string    `json:"technology"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
type Page struct {
	Items    []Summary `json:"items"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
	Total    int64     `json:"total"`
}
type Store interface {
	Save(context.Context, model.Answer, string) (Saved, error)
	Get(context.Context, string) (Saved, error)
	List(context.Context, int, int, string, string, string) (Page, error)
	Rename(context.Context, string, string) (Saved, error)
	Delete(context.Context, string) (bool, error)
	Ready(context.Context) error
}
type Postgres struct {
	Pool    *pgxpool.Pool
	Queries *database.Queries
}

func New(pool *pgxpool.Pool) *Postgres { return &Postgres{pool, database.New(pool)} }
func id(v string) pgtype.UUID          { u, _ := uuid.Parse(v); return pgtype.UUID{Bytes: u, Valid: true} }
func saved(row database.SavedQuery, err error) (Saved, error) {
	if err != nil {
		return Saved{}, err
	}
	var response model.Answer
	if err = json.Unmarshal(row.Response, &response); err != nil {
		return Saved{}, err
	}
	return Saved{ID: uuid.UUID(row.ID.Bytes).String(), Question: row.Question, Title: row.Title, Technology: row.Technology.String, Response: response, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time}, nil
}
func (s *Postgres) Save(ctx context.Context, a model.Answer, title string) (Saved, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return Saved{}, err
	}
	return saved(s.Queries.CreateSaved(ctx, database.CreateSavedParams{ID: id(uuid.NewString()), Question: a.Query, Title: title, Technology: pgtype.Text{String: a.Technology, Valid: true}, Response: b}))
}
func (s *Postgres) Get(ctx context.Context, v string) (Saved, error) {
	return saved(s.Queries.GetSaved(ctx, id(v)))
}
func (s *Postgres) Rename(ctx context.Context, v, title string) (Saved, error) {
	return saved(s.Queries.RenameSaved(ctx, database.RenameSavedParams{ID: id(v), Title: title}))
}
func (s *Postgres) Delete(ctx context.Context, v string) (bool, error) {
	n, err := s.Queries.DeleteSaved(ctx, id(v))
	return n > 0, err
}
func (s *Postgres) Ready(ctx context.Context) error {
	var exists bool
	err := s.Pool.QueryRow(ctx, "SELECT to_regclass('public.saved_queries') IS NOT NULL").Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return pgx.ErrNoRows
	}
	return nil
}
func (s *Postgres) List(ctx context.Context, page, size int, filter, tech, order string) (Page, error) {
	p := Page{Items: []Summary{}, Page: page, PageSize: size}
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return p, err
	}
	defer tx.Rollback(ctx)
	q := s.Queries.WithTx(tx)
	p.Total, err = q.CountSaved(ctx, database.CountSavedParams{Filter: filter, Tech: tech})
	if err != nil {
		return p, err
	}
	rows, err := q.ListSaved(ctx, database.ListSavedParams{Filter: filter, Tech: tech, SortOrder: order, PageOffset: int32((page - 1) * size), PageLimit: int32(size)})
	if err != nil {
		return p, err
	}
	for _, r := range rows {
		p.Items = append(p.Items, Summary{ID: uuid.UUID(r.ID.Bytes).String(), Question: r.Question, Title: r.Title, Technology: r.Technology.String, CreatedAt: r.CreatedAt.Time, UpdatedAt: r.UpdatedAt.Time})
	}
	return p, tx.Commit(ctx)
}
