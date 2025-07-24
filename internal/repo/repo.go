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
	ID          uint32 `json:"id"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type ArticlesRepository struct {
	pool *pgxpool.Pool
}

// NewArticlesRepository creates new repo
func NewArticlesRepository(ctx context.Context) (*ArticlesRepository, error) {

	url := "postgres://user1:12345678@localhost:16432/db1?sslmode=disable"

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

func (r *ArticlesRepository) Select(ctx context.Context, id int) (*Article, error) {
	sql := `select id, url, title, description from articles where id=$1`

	res, err := r.pool.Query(ctx, sql, id)

	if err != nil {
		return nil, err
	}
	notFound := true
	art := Article{}
	for notFound && res.Next() {
		err := res.Scan(&art.ID, &art.URL, &art.Title, &art.Description)
		if err != nil {
			fmt.Printf("unable to scan row: %v", err)
			return nil, err
		}
		notFound = false
	}
	if notFound {
		return nil, fmt.Errorf("Article id=%d not found", id)
	}
	return &art, nil
}

func (repo *ArticlesRepository) UpsertArticles(ctx context.Context, articles []*Article) ([]*Article, error) {
	tx, err := repo.pool.BeginTx(context.TODO(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.TODO())
		} else {
			tx.Commit(context.TODO())
		}
	}()
	sql := `INSERT INTO articles (id, url, title, description)
        VALUES ($1, $2, $3, $4) 
      ON CONFLICT (url)
        DO UPDATE SET
          title = EXCLUDED.title,
          description = EXCLUDED.description
      RETURNING id;
      `

	batch := &pgx.Batch{}
	for _, a := range articles {
		batch.Queue(sql, a.ID, a.URL, a.Title, a.Description)
	}

	rows := tx.SendBatch(ctx, batch)
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("Failed to close batch results for cards: Error: %v\n", err)
		}
	}()

	for i := range articles {
		var id uint32
		err := rows.QueryRow().Scan(&id)
		if err != nil {
			return articles, fmt.Errorf("failed to upsert article with url %s. err: %w", articles[i].URL, err)
		}
		articles[i].ID = id
	}

	return articles, nil
}
