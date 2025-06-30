package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spaolacci/murmur3"
)

const seed = 0x12345678

// Article model
type Article struct {
	ID          uint32
	URL         string
	Title       string
	Description string
}

type ArticlesRepository struct {
	pool *pgxpool.Pool
}

// NewArticlesRepository creates new repo
func NewArticlesRepository(ctx context.Context) (*ArticlesRepository, error) {

	url := "postgres://user1:admin@localhost:6432/db1?sslmode=disable"

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return &ArticlesRepository{pool}, nil
}

// Create creates article
func (repo *ArticlesRepository) Create(ctx context.Context, a *Article) error {
	a.ID = murmur3.Sum32WithSeed([]byte(a.URL), seed) >> 1
	sql := `INSERT INTO articles (id, url, title, description) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`

	_, err := repo.pool.Exec(ctx, sql, a.ID, a.URL, a.Title, a.Description)

	if err != nil {
		return err
	}

	return nil
}
func (repo *ArticlesRepository) CreateArticle(ctx context.Context, a *Article) error {
	sql := `INSERT INTO articles (id, url, title, description) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`

	_, err := repo.pool.Exec(ctx, sql, a.ID, a.URL, a.Title, a.Description)

	if err != nil {
		return err
	}

	return nil
}

// Create creates article
func (repo *ArticlesRepository) Select(ctx context.Context, a *Article) (pgx.Rows, error) {
	sql := `select id, url, title, description from articles where id=$1`

	res, err := repo.pool.Query(ctx, sql, a.ID)

	if err != nil {
		return nil, err
	}

	return res, nil
}
