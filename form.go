package mechclient

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
)

func (b *Body) forms() *goquery.Selection {
	if b.document != nil {
		return b.document.Find("form")
	} else {
		panic(fmt.Sprint("Response has not been parsed yet!"))
	}
}

func (b *Body) formWith(selector, value string) *goquery.Selection {
	switch selector {
	case "name", "action":
		forms := b.forms()
		return forms.FilterFunction(func(i int, s *goquery.Selection) bool {
			for _, val := range s.Nodes[0].Attr {
				if val.Key == selector {
					if val.Val == value {
						return true
					}
					break
				}
			}
			return false
		})
	}
	panic(fmt.Sprintf("%v is not a valid selector", selector))
}

func (b *Body) FormWith(selector, value string) *Body {
	b.selection = b.formWith(selector, value).First()
	b.FormValues = b.getValues()
	return b
}

// Gets the input values of the form selected by name for posting
func (b *Body) getValues() url.Values {
	inputVals := url.Values{}
	inputs := b.selection.Find("input")
	for _, input := range inputs.Nodes {
		var inputName string
		for _, name := range input.Attr {
			if name.Key == "name" {
				inputName = name.Val
				inputVals.Add(name.Val, "")
			}
			if name.Key == "value" {
				inputVals.Set(inputName, name.Val)
			}
		}
	}
	return inputVals
}

// Creates a string url for the form selected to send an http PostForm request.
// Adds authorization if needed.
func (b *Body) formAddress() string {
	var path string
	var fullPath string
	for _, attr := range b.selection.Nodes[0].Attr {
		if attr.Key == "action" {
			path = attr.Val
		}
	}
	if string(path[0]) == "/" {
		fullPath = b.client.history[0] + path[1:]
	} else {
		fullPath = b.client.history[0] + path
	}
	return fullPath
}

// Posts to a form, either what was selected from FormWith or the first available
// form on the page. If val is nil, uses default values found from that form, or values
// added/changed by setting the value i.e. b.FormValues.Set in main function.
func (b *Body) PostForm(val url.Values) (resp *http.Response, err error) {
	if b.selection == nil {
		b.selection = b.forms().First()
		if len(b.selection.Nodes) == 0 {
			panic(fmt.Sprint("There were no forms found on this page"))
		}
		b.FormValues = b.getValues()
	}
	if val != nil {
		for key, values := range val {
			if len(values) > 1 {
				b.FormValues.Del(key)
				for _, value := range values {
					b.FormValues.Add(key, value)
				}
			} else {
				b.FormValues.Set(key, values[0])
			}
		}
	}
	return b.client.postForm(b.formAddress(), b.FormValues)
}
