package main

import (
	"context"
	"fmt"
	"log"

	"github.com/denchick/news-aggregator/repo"
)

func main() {
	ctx := context.Background()
	repos, err := repo.NewArticlesRepository(ctx)
	if err != nil {
		log.Fatal(err)
	}

	art := &repo.Article{
		ID:          92929,
		URL:         "https://www.example.com",
		Title:       "do electric sheep dream of androids quotes",
		Description: "No, they do not",
	}

	err = repos.Create(ctx, art)
	if err != nil {
		log.Fatal(err)
	}

	res, err := repos.Select(ctx, art)
	if err != nil {
		log.Fatal(err)
	} else {
		arts := []repo.Article{}
		for res.Next() {
			art := repo.Article{}
			err := res.Scan(&art.ID, &art.URL, &art.Title, &art.Description)
			if err != nil {
				fmt.Printf("unable to scan row: %v", err)
				return
			}
			arts = append(arts, art)
		}
		fmt.Printf("<%v", arts)
	}

}
