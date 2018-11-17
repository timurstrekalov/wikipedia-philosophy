package wikipedia

import (
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"net/http"
	"regexp"
	"strings"
)

const wikiPrefix = "/wiki/"

var wikiLinkRe *regexp.Regexp

func init() {
	var err error
	wikiLinkRe, err = regexp.Compile(".+/wiki/(.+)")
	if err != nil {
		panic(err)
	}
}

type finder struct {
	from    string
	to      string
	path    []PathElement
	visited map[string]bool
}

type PathElement struct {
	Title string `json:"title"`
	Uri   string `json:"uri"`
}

func GetPath(from string, to string) ([]PathElement, error) {
	f := &finder{from: from, to: to}
	f.path = make([]PathElement, 0, 8)
	f.visited = make(map[string]bool, 0)

	return f.getPath(f.from, f.to)
}

func (f *finder) getPath(from string, to string) ([]PathElement, error) {
	candidateLink, err := f.findCandidateLink(from)
	if from == to {
		return f.path, nil
	}

	if err != nil {
		return f.path, err
	}

	return f.getPath(candidateLink, to)
}

// FIXME side-effecty & hacky
func (f *finder) findCandidateLink(from string) (string, error) {
	resp, err := http.Get("https://en.wikipedia.org/wiki/" + from)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)

	depth := 0
	openItalics := 0
	openParens := 0
	openParagraphs := 0
	content := false
	contentDepth := -1
	inTitle := false
	title := ""

	for {
		tokenType := z.Next()
		token := z.Token()

		switch tokenType {
		case html.ErrorToken:
			return "", z.Err()
		case html.TextToken:
			if inTitle {
				title = strings.Replace(token.Data, " - Wikipedia", "", 1)
			} else {
				openParens += strings.Count(token.Data, "(") - strings.Count(token.Data, ")")
			}
		case html.SelfClosingTagToken:
			if token.DataAtom == atom.Link {
				if ok, rel := getAttributeValue(token, "rel"); ok && rel == "canonical" {
					if ok, href := getAttributeValue(token, "href"); ok {
						uri := wikiLinkRe.ReplaceAllString(href, "$1")

						f.visited[uri] = true
						f.path = append(f.path, PathElement{
							Title: title,
							Uri:   uri,
						})
					}
				}
			}
		case html.StartTagToken:
			depth++

			if content {
				switch token.DataAtom {
				case atom.I:
					openItalics++
				case atom.P:
					openParagraphs++
				case atom.A:
					if content && openItalics == 0 && openParens == 0 && openParagraphs > 0 {
						if ok, href := getAttributeValue(token, "href"); ok {
							if strings.HasPrefix(href, wikiPrefix) && !strings.Contains(href, "wiktionary.org") {
								page := strings.TrimPrefix(href, wikiPrefix)

								if isValidPage(page) && !f.visited[page] {
									return page, nil
								}
							}
						}
					}
				}
			} else {
				switch token.DataAtom {
				case atom.Title:
					inTitle = true
				case atom.Div:
					if ok, id := getAttributeValue(token, "id"); ok && id == "mw-content-text" {
						content = true
						contentDepth = depth
					}
				}
			}
		case html.EndTagToken:
			depth--

			if content {
				if depth == contentDepth {
					content = false
					contentDepth = -1
				} else {
					switch token.DataAtom {
					case atom.I:
						openItalics--
					case atom.P:
						openParagraphs--
					}
				}
			} else {
				switch token.DataAtom {
				case atom.Title:
					inTitle = false
				}
			}
		}
	}

	return "", fmt.Errorf("could not find candidate link on page %s", from)
}

func getAttributeValue(t html.Token, attrName string) (bool, string) {
	for _, a := range t.Attr {
		if a.Key == attrName {
			return true, a.Val
		}
	}

	return false, ""
}

func isValidPage(page string) bool {
	return !strings.HasPrefix(page, "Help:") &&
		!strings.HasPrefix(page, "File:") &&
		!strings.HasPrefix(page, "Wikipedia:")
}
