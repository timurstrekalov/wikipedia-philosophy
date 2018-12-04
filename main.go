package main

import (
	"encoding/json"
	"github.com/timurstrekalov/wikipedia-philosophy/parser"
	"log"
	"net/http"
)

const baseWikipediaURL = "https://en.wikipedia.org/wiki/"

func toWikipediaURL(page string) string {
	return baseWikipediaURL + page
}

// TODO
// - expose static resources at /
// - implement /api/path, taking query parameters "from" and "to", returning a JSON array of objects {pageId, title}
func main() {
	fs := http.FileServer(http.Dir("static"))

	http.Handle("/", fs)
	http.HandleFunc("/api/path", func(writer http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()
		from := query.Get("from")
		to := query.Get("to")

		path, err := findPath(from, to)
		if err != nil {
			log.Println(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(writer).Encode(path)
		if err != nil {
			log.Println(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	addr := ":8080"
	log.Printf("Listening on %s...\n", addr)

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}

type PathElement struct {
	PageId string `json:"pageId"`
	Title  string `json:"title"`
}

func findPath(from string, to string) ([]PathElement, error) {
	path := make([]PathElement, 0, 8)
	pp := parser.NewPageParser()

	for {
		page, err := loadPage(pp, from)
		if err != nil {
			return nil, err
		}

		path = append(path, PathElement{
			PageId: page.Id,
			Title:  page.Title,
		})

		if from == to {
			return path, nil
		} else {
			from = page.ValidLinks[0]
		}
	}
}

func loadPage(pp *parser.PageParser, pageName string) (*parser.Page, error) {
	wikipediaURL := toWikipediaURL(pageName)

	log.Printf("requesting page %s", wikipediaURL)

	resp, err := http.Get(wikipediaURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return pp.ParsePage(resp.Body)
}
