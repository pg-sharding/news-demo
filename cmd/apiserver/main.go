package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/denchick/news-aggregator/internal/repo"
)

func articleHandler(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		http.Error(w, "Bad Request: id not found", http.StatusBadRequest)
	}
	if idVal, err := strconv.Atoi(id); err != nil {
		http.Error(w, fmt.Sprintf("Bad Request: id %s is incorrect", id), http.StatusBadRequest)
	} else {
		ctx := context.Background()
		repos, err := repo.NewArticlesRepository(ctx)
		if err != nil {
			log.Fatal(err)
		}
		if art, err := repos.Select(ctx, idVal); err != nil {
			http.Error(w, fmt.Sprintf("Bad Request: id %s is incorrect", id), http.StatusBadRequest)
		} else {
			fmt.Println(`id := `, art.ID)
			json.NewEncoder(w).Encode(art)
		}
	}

}

func main() {
	http.HandleFunc("/article/{id}", articleHandler)

	fmt.Println("Starting server at port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting the server:", err)
	}
}
