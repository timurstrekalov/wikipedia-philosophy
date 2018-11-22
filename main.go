package main

import (
	"encoding/json"
	"github.com/timurstrekalov/wikipedia-philosophy/wikipedia"
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	http.HandleFunc("/api/path", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		from := query.Get("from")
		to := query.Get("to")

		path, err := getPath(from, to)

		if err != nil {
			log.Printf("error getting path from %s to %s: %v", from, to, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(path)
		if err != nil {
			log.Printf("error encoding JSON: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	log.Println("Listening...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

type PathElement struct {
	Title  string `json:"title"`
	PageId string `json:"pageId"`
}

type pathFinder struct {
	path   []PathElement
	parser *wikipedia.PageParser
}

func getPath(from string, to string) ([]PathElement, error) {
	f := pathFinder{}
	f.path = make([]PathElement, 0, 8)
	f.parser = wikipedia.NewPageParser()

	err := f.findPath(from, to)
	if err != nil {
		return nil, err
	}

	return f.path, nil
}

func (f *pathFinder) findPath(from string, to string) (error) {
	resp, err := http.Get(wikipedia.ToURLString(from))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	page, err := f.parser.ParsePage(resp.Body)
	if err != nil {
		return err
	}

	f.addPage(page)

	if from == to {
		return nil
	}

	return f.findPath(page.ValidLinks[0], to)
}

func (f *pathFinder) addPage(page *wikipedia.Page) {
	f.path = append(f.path, PathElement{
		Title:  page.Title,
		PageId: page.Id,
	})
}
