package webby

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"encoding/csv"
	"net/url"
	"strings"
	"net/http/cookiejar"
	"bytes"
)

const (
	userAgentValue    = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/58.0.3029.110 Chrome/58.0.3029.110 Safari/537.36"
	languageValue     = "en-US,en;q=0.8"
	cacheControlValue = "max-age=0"
)

type Api struct {
	Client *http.Client
}

// Initialize cookie jar
func (c *Api) EnableCookies() {
	if c.Client == nil {
		c.Client = &http.Client{}
	}
	c.Client.Jar, _ = cookiejar.New(nil)
}

// Return the JSON content from the api into "v"
func (c *Api) GetJSON(uri string, v interface{}) error {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return err
	}

	c.setDefaultHeaders(req, "application/json")

	resp, err := c.do(req)
	if err != nil {
		return err

	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(v)
	resp.Body.Close()

	if err == io.EOF {
		return nil
	}

	return err
}

// Return the CSV content from the api into "acceptRecord"
func (c *Api) GetCSV(uri string, acceptRecord func([]string) error) error {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return err
	}

	c.setDefaultHeaders(req, "text/csv")

	resp, err := c.do(req)
	if err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %s", resp.Status)
	}

	err = readCSV(resp.Body, acceptRecord)
	resp.Body.Close()
	return err
}

// Return the CSV content from the api into "writer"
func (c *Api) GetBody(uri string, writer io.Writer) error {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return err
	}

	c.setDefaultHeaders(req, "")

	resp, err := c.do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %s", resp.Status)
	}

	_, err = io.Copy(writer, resp.Body)
	resp.Body.Close()
	return err
}

func (c *Api) do(req *http.Request) (*http.Response, error) {
	if c.Client == nil {
		c.Client = &http.Client{}
	}
	return c.Client.Do(req)
}

// Set dome default headers
func (c *Api) setDefaultHeaders(req *http.Request, contentType string) {
	req.Header.Set("User-Agent", userAgentValue)
	req.Header.Set("Accept-Language", languageValue)
	req.Header.Set("Cache-Control", cacheControlValue)
	if contentType != "" {
		req.Header.Set("Accept", contentType)
	}
}

func readCSV(r io.Reader, acceptRecord func([]string) error) (err error) {
	var reader = csv.NewReader(r)

	for {
		var record []string
		record, err = reader.Read()
		if err != nil {
			break
		}

		err = acceptRecord(record)
		if err != nil {
			break
		}
	}

	if err == io.EOF {
		return nil
	}
	return
}

type CSVRows struct {
	Rows [][]string
}

func (c *CSVRows) Add(row []string) error {
	c.Rows = append(c.Rows, row)
	return nil
}

// GetLastURLSegment will get the last path segment of a url.
// For ex. with url "https://some.domain/path1/path2/path3?a=b,
// 'path3' will be returned.
func GetLastURLSegment(fullUrl string) (string, error) {
	URL, err := url.Parse(fullUrl)
	if err != nil {
		return "", err
	}

	path := URL.Path
	if path == "" {
		return "", nil
	}

	segments := strings.Split(path, "/")
	return segments[len(segments)-1], nil
}

type UrlBuilder struct {
	base   string
	path   string
	params map[string]string
}

func (u *UrlBuilder) Base(base string) *UrlBuilder {
	u.base = base
	return u
}

func (u *UrlBuilder) Path(path string) *UrlBuilder {
	u.path = path
	return u
}

func (u *UrlBuilder) Param(key, value string) *UrlBuilder {
	if value == "" {
		return u
	}

	if u.params == nil {
		u.params = make(map[string]string)
	}

	u.params[key] = value
	return u
}

func (u *UrlBuilder) Build() string {
	buf := bytes.NewBufferString(u.base)
	buf.WriteString(u.path)

	if len(u.params) > 0 {
		sep := "?"
		for k, v := range u.params {
			buf.WriteString(sep)
			buf.WriteString(k)
			buf.WriteString("=")
			buf.WriteString(v)
			sep = "&"
		}
	}

	return buf.String()
}
