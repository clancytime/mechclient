package mechclient

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
)

// Helper method to find all available forms on parsed page
func (b *Body) forms() *goquery.Selection {
	if b.document != nil {
		return b.document.Find("form")
	} else {
		panic(fmt.Sprint("Response has not been parsed yet!"))
	}
}

// Helper method for FormWith. Utilizes goquery FilterFunction to find forms within
// parsed page whose meta-info matches the descriptors. Name and action are the most
// available options, may add more based on need.
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

// Finds the first form that describes the selectors, then retreives the input
// values inside that form for availablility when using PostForm request. Input
// values can either be changed in PostForm request or by using url.Values methods
// i.e. form.FormValues.Set("foo", "bar")
func (b *Body) FormWith(selector, value string) *Body {
	b.selection = b.formWith(selector, value)
	if len(b.selection.Nodes) < 1 {
		panic(fmt.Sprintf("No forms were found with %v set to %v", selector, value))
	} else {
		b.selection = b.selection.First()
		b.FormValues = b.getValues()
	}
	return b
}

// Gets the input values of the form selected
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
	for _, attr := range b.selection.Nodes[0].Attr {
		if attr.Key == "action" {
			path = attr.Val
		}
	}
	historyLength := len(b.client.history)
	u, _ := url.Parse(b.client.history[historyLength-1])
	if string(path[0]) == "/" {
		u.Path = path
	} else {
		u.Path = "/" + path
	}
	return u.String()
}

// Wrapper for HTTP PostForm method, either what was selected from FormWith or the first available
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
