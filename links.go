package mechclient

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
)

// TODO: Change back to links
func (b *Body) links() *goquery.Selection {
	if b.document != nil {
		return b.document.Find("a")
	} else {
		panic(fmt.Sprint("Response has not been parsed yet!"))
	}
}

func (b *Body) linksWith(selector, search string) *goquery.Selection {
	switch selector {
	case "text":
		links := b.links()
		return links.FilterFunction(func(i int, s *goquery.Selection) bool {
			if s.Text() == search {
				return true
			} else {
				return false
			}
		})
	}
	panic(fmt.Sprintf("%v is not an available selector for links", selector))
}

func (b *Body) LinksWith(selector, search string) *Body {
	b.selection = b.linksWith(selector, search)
	return b
}

// Same as above, except it grabs the first link that matches the search and adds that to
// Body
func (b *Body) LinkWith(selector, search string) *Body {
	b.selection = b.linksWith(selector, search)
	if len(b.selection.Nodes) < 1 {
		panic(fmt.Sprint("There are no nodes to select"))
	} else {
		b.selection = b.selection.First()
	}
	return b
}

func (b *Body) Click() (resp *http.Response, err error) {
	if len(b.selection.Nodes) > 0 {
		path := b.selection.Nodes[0].Attr[0].Val
		if _, err := url.Parse(path); err != nil {
			return b.client.Get(path)
		} else {
			historyLength := len(b.client.history)
			u, _ := url.Parse(b.client.history[historyLength-1])
			if string(path[0]) == "/" {
				u.Path = path
			} else {
				u.Path = "/" + path
			}
			return b.client.Get(u.String())
		}
	} else {
		panic(fmt.Sprint("There are no Nodes to click!"))
	}
}
