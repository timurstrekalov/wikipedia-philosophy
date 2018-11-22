package wikipedia

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"io/ioutil"
	"net/url"
	"strings"
)

const wikiPrefix = "/wiki/"

type Page struct {
	Title      string
	Id         string
	ValidLinks []string
}

type PageParser struct {
	visited map[string]bool
}

func NewPageParser() *PageParser {
	pp := &PageParser{}
	pp.visited = make(map[string]bool, 0)

	return pp
}

func (pp *PageParser) ParsePage(r io.Reader) (*Page, error) {
	defer io.Copy(ioutil.Discard, r)

	z := html.NewTokenizer(r)

	depth := 0
	openItalics := 0
	openParens := 0
	openParagraphs := 0
	inContent := false
	contentDepth := -1
	inTitle := false

	page := Page{}
	page.ValidLinks = make([]string, 0, 8)

	for {
		tokenType := z.Next()
		token := z.Token()

		switch tokenType {
		case html.ErrorToken:
			if z.Err() == io.EOF {
				goto done
			}

			return nil, z.Err()
		case html.TextToken:
			if inTitle {
				page.Title = strings.Replace(token.Data, " - Wikipedia", "", 1)
			} else {
				openParens += strings.Count(token.Data, "(") - strings.Count(token.Data, ")")
			}
		case html.SelfClosingTagToken:
			if token.DataAtom == atom.Link {
				if ok, rel := getAttributeValue(token, "rel"); ok && rel == "canonical" {
					if ok, href := getAttributeValue(token, "href"); ok {
						pageUrl, err := url.Parse(href)
						if err != nil {
							return nil, err
						}

						pageId := strings.TrimPrefix(pageUrl.Path, wikiPrefix)
						pp.visited[pageId] = true
						page.Id = pageId
					}
				}
			}
		case html.StartTagToken:
			depth++

			if inContent {
				switch token.DataAtom {
				case atom.I:
					openItalics++
				case atom.P:
					openParagraphs++
				case atom.A:
					if inContent && openItalics == 0 && openParens == 0 && openParagraphs > 0 {
						if ok, href := getAttributeValue(token, "href"); ok {
							if strings.HasPrefix(href, wikiPrefix) &&
								!strings.Contains(href, "wiktionary.org") &&
								!strings.Contains(href, "#cite-note") {

								nextPageId := strings.SplitAfter(strings.TrimPrefix(href, wikiPrefix), "#")[0]

								if isValidPage(nextPageId) && !pp.visited[nextPageId] {
									page.ValidLinks = append(page.ValidLinks, nextPageId)
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
						inContent = true
						contentDepth = depth
					}
				}
			}
		case html.EndTagToken:
			depth--

			if inContent {
				if depth == contentDepth {
					inContent = false
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

done:
	return &page, nil
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
