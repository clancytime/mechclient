// Basic Implementation of Mechanize for Go. Wraps http.Client
// and CookieJar and utilizes goquery (http://github.com/PuerkitoBio/goquery)
// to automate interaction within websites.
//

package mechclient

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type MechClient struct {
	client  *http.Client
	auth    *url.URL
	history []string
}

type Body struct {
	client     *MechClient
	Document   *goquery.Document
	selection  *goquery.Selection
	FormValues url.Values
}

// Initalize mechclient, adding cookiejar to client for storing
func New() *MechClient {
	j, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	m := &MechClient{}
	m.client = &http.Client{Jar: j}

	return m
}

// Adds http authentication to a url specified
func (m *MechClient) AddAuth(dom, user, pass string) {
	u, err := url.Parse(dom)
	if err != nil {
		panic(err)
	}
	u.User = url.UserPassword(user, pass)
	m.auth = u
}

// Wrapper for HTTP Get method, appends url string to history array
func (m *MechClient) Get(address string) *Body {
	u := m.addAuthTo(address)
	m.history = append(m.history, address)
	resp, err := m.client.Get(u)
	if err != nil {
		panic(err)
	}
	b := m.parse(resp)
	return b
}

//adds any authorization added from AddAuth to url before sending http request
func (m *MechClient) addAuthTo(address string) string {
	if m.auth != nil {
		// turn address string into url.URL to check host against auth host
		u, err := url.Parse(address)
		if err != nil {
			panic(err)
		}

		if u.Host == m.auth.Host {
			// add user/pass info to address url
			user := m.auth.User.Username()
			pass, _ := m.auth.User.Password()
			u.User = url.UserPassword(user, pass)
			return u.String()
		} else {
			return address
		}
	} else {
		return address
	}
}

// Produces a string array of the full history of the client, starting with the first visited
// url. Does not include authorization if used.
func (m *MechClient) History() []string {
	return m.history
}

// Returns cookies from designate URL value
func (m *MechClient) Cookies(u *url.URL) (cookies []*http.Cookie) {
	return m.client.Jar.Cookies(u)
}

// Parses the response to find links and forms, creates a wrapper for Goquery
func (m *MechClient) parse(res *http.Response) *Body {
	b := &Body{client: m}
	b.Document, _ = goquery.NewDocumentFromResponse(res)
	return b
}

// helper for Body.PostForm method
func (m *MechClient) postForm(address string, val url.Values) (resp *http.Response, err error) {
	u := m.addAuthTo(address)
	m.history = append(m.history, address)
	return m.client.PostForm(u, val)
}

// helper method for Body.LinksWith and Body.FormWith
func (b *Body) find(selector string) *goquery.Selection {
	if b.Document != nil {
		return b.Document.Find(selector)
	} else {
		panic(fmt.Sprint("Response has not been parsed yet!"))
	}
}
