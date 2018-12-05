package main

import (
	"encoding/json"
	"github.com/timurstrekalov/wikipedia-philosophy/parsing"
	"log"
	"net/http"
)

const basePageURL = "https://en.wikipedia.org/wiki/"

func toPageURL(page string) string {
	return basePageURL + page
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
	parser := parsing.NewPageParser()

	for {
		page, err := loadPage(parser, from)
		if err != nil {
			return nil, err
		}

		path = append(path, PathElement{
			PageId: page.Id,
			Title:  page.Title,
		})

		if from == to {
			break
		}

		from = page.ValidLinks[0]
	}

	return path, nil
}

func loadPage(parser *parsing.PageParser, page string) (*parsing.Page, error) {
	pageURL := toPageURL(page)

	log.Printf("requesting page %s", pageURL)

	resp, err := http.Get(pageURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parser.ParsePage(resp.Body)
}
