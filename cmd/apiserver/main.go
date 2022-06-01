package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/denchick/news-aggregator/repo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	repo, err := repo.NewArticlesRepository()
	if err != nil {
		return err
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		articles, err := repo.GetAll()
		if err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				fmt.Errorf("could not get any articles: %w", err),
			)
		}
		if len(articles) == 0 {
			return c.NoContent(http.StatusNotFound)
		}
		return c.JSON(http.StatusOK, articles)
	})

	s := &http.Server{
		Addr:         ":1323",
		ReadTimeout:  30 * time.Minute,
		WriteTimeout: 30 * time.Minute,
	}
	e.Logger.Print(e.StartServer(s))

	return nil
}