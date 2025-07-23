package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/denchick/news-aggregator/repo"
	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

var RSSFeeds = []string{
	"https://news.ycombinator.com/rss",
	"https://habr.com/rss/",
	"https://www.theverge.com/rss/index.xml",
	"https://www.wired.com/feed/category/science/latest/rss",
}

func run() error {
	ctx := context.Background()
	s, err := repo.NewArticlesRepository(ctx)
	if err != nil {
		return err
	}

	for _, feed := range RSSFeeds {
		log.Printf("Start %s parsing...", feed)
		articles, err := parseFeed(feed)
		if err != nil {
			return err
		}
		for _, a := range articles {
			if err := s.Create(ctx, a); err != nil {
				return fmt.Errorf("can't save article %+v: %w", a, err)
			}
		}
		log.Printf("Saved %d articles for %s.", len(articles), feed)
	}
	return nil
}

func parseFeed(feed string) ([]*repo.Article, error) {
	parser := gofeed.NewParser()
	parsedFeed, err := parser.ParseURL(feed)
	if err != nil {
		return nil, fmt.Errorf("parsing URL failed: %w", err)
	}

	var articles []*repo.Article
	for _, item := range parsedFeed.Items {
		a := &repo.Article{
			URL:         item.Link,
			Title:       item.Title,
			Description: getItemDescription(item),
		}
		if len(a.Description) > 0 {
			log.Printf("OK: %s with dess '%s'", a.URL, a.Description)
			articles = append(articles, a)
		}
	}
	log.Printf("Parsed %d articles from %s", len(articles), feed)
	return articles, nil
}

func getItemDescription(item *gofeed.Item) string {
	var description string
	if len(item.Description) > 0 {
		description = item.Description
	} else {
		description = item.Content
	}

	return cleanText(description, 500)
}

func cleanText(text string, limit int) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(text))
	if err != nil {
		fmt.Print(fmt.Errorf("NewDocumentFromReader failed: %w", err))
		return ""
	}

	text = strings.TrimSpace(doc.Text())
	text = strings.ReplaceAll(text, "\n", " ")

	runes := []rune(text)
	if len(runes) >= limit {
		return string(runes[:limit]) + "..."
	}
	return text
}
