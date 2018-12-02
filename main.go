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
	pp := parser.NewPageParser()
	path := make([]PathElement, 0, 8)
	return findPathRecursively(from, to, path, pp)
}

func findPathRecursively(from string, to string, path []PathElement, pp *parser.PageParser) ([]PathElement, error) {
	resp, err := http.Get(toWikipediaURL(from))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	page, err := pp.ParsePage(resp.Body)
	path = append(path, PathElement{
		PageId: page.Id,
		Title: page.Title,
	})

	if from == to {
		return path, nil
	}

	return findPathRecursively(page.ValidLinks[0], to, path, pp)
}
