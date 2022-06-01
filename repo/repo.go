package repo

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Article model
type Article struct {
	ID          uint
	URL         string
	Title       string
	Description string
}

type ArticlesRepository struct {
	pool *pgxpool.Pool
}

// NewArticlesRepository creates new repo
func NewArticlesRepository() (*ArticlesRepository, error) {
	pool, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &ArticlesRepository{pool}, nil
}

// Create creates article
func (repo *ArticlesRepository) Create(a *Article) error {
	_, err :=repo.pool.Exec(context.Background(),
		"INSERT INTO articles (url, title, description) VALUES ($1, $2, $3)",
		a.Title, a.URL, a.Description)

	if err != nil {
		pgerr := err.(*pgconn.PgError)
		if pgerrcode.IsIntegrityConstraintViolation(pgerr.Code) {
			log.Printf("article already exists: %s", a.URL)
			return nil
		}
		return err
	}
	return nil
}

func (repo *ArticlesRepository) GetAll() ([]*Article, error) {
	rows, err := repo.pool.Query(context.Background(), "SELECT * FROM articles")
	if err != nil {
		return nil, fmt.Errorf("unable to SELECT: %w", err)
	}
	defer rows.Close()

	articles := []*Article{}
	for rows.Next() {
		a := Article{}
		err := rows.Scan(&a.ID, &a.URL, &a.Title, &a.Description)
		if err != nil {
			return nil, err
		}
		articles = append(articles, &a)
	}

	return articles, nil
}
