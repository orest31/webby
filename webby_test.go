package webby_test

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
	"net/http/httptest"
	"github.com/orest31/webby"
	"github.com/stretchr/testify/assert"
)

func TestGetCSV(t *testing.T) {
	t.Run("with no records", func(t *testing.T) {
		rows, err := getCSV("", http.StatusOK)
		assert.Nil(t, err)
		assert.Len(t, rows.Rows, 0)
	})

	t.Run("with 1 record", func(t *testing.T) {
		response := strings.Join([]string{`1,2,abc`}, "\n")

		rows, err := getCSV(response, http.StatusOK)
		assert.Nil(t, err)
		assert.Len(t, rows.Rows, 1)
		assert.Equal(t, rows.Rows[0], []string{"1", "2", "abc"})
	})

	t.Run("with 2 records", func(t *testing.T) {
		response := strings.Join([]string{`1,2,row1`, `4,5,row2`}, "\n")

		rows, err := getCSV(response, http.StatusOK)
		assert.Nil(t, err)
		assert.Len(t, rows.Rows, 2)
		assert.Equal(t, rows.Rows[0], []string{"1", "2", "row1"})
		assert.Equal(t, rows.Rows[1], []string{"4", "5", "row2"})
	})

	t.Run("with Server Not Found error", func(t *testing.T) {
		_, err := getCSV("", http.StatusNotFound)
		assert.NotNil(t, err)
	})
}

func getCSV(response string, status int) (*webby.CSVRows, error) {
	ts := mockWebResponse(response, status)
	api := &webby.Api{}
	rows := &webby.CSVRows{}
	err := api.GetCSV(ts.URL, rows.Add)
	ts.Close()
	return rows, err
}

func TestGetJSON(t *testing.T) {
	t.Run("with no json body", func(t *testing.T) {
		_, err := getJSON("", http.StatusOK)
		assert.Nil(t, err)
	})

	t.Run("with an empty json body", func(t *testing.T) {
		_, err := getJSON("{}", http.StatusOK)
		assert.Nil(t, err)
	})

	t.Run("with a JSON response", func(t *testing.T) {
		result, err := getJSON(`{"key":"value"}`, http.StatusOK)
		assert.Nil(t, err)
		assert.Equal(t, "value", result.Key)
	})

	t.Run("with Server Not Found error", func(t *testing.T) {
		_, err := getJSON("", http.StatusNotFound)
		assert.NotNil(t, err)
	})
}

func getJSON(response string, status int) (*struct{ Key string }, error) {
	ts := mockWebResponse(response, status)
	api := &webby.Api{}
	result := &struct{ Key string }{}
	err := api.GetJSON(ts.URL, result)
	ts.Close()
	return result, err
}

func TestGetBody(t *testing.T) {
	t.Run("with no body", func(t *testing.T) {
		_, err := getBody("", http.StatusOK)
		assert.Nil(t, err)
	})

	t.Run("with a response", func(t *testing.T) {
		result, err := getBody("<html>Some Content</html>", http.StatusOK)
		assert.Nil(t, err)
		assert.Equal(t, "<html>Some Content</html>", result)
	})

	t.Run("with Server Not Found error", func(t *testing.T) {
		_, err := getBody("", http.StatusNotFound)
		assert.NotNil(t, err)
	})
}

func getBody(response string, status int) (string, error) {
	ts := mockWebResponse(response, status)
	api := &webby.Api{}
	result := &bytes.Buffer{}
	err := api.GetBody(ts.URL, result)
	ts.Close()
	return result.String(), err
}

func TestGetLastURLSegment(t *testing.T) {
	type testCase struct{ url, expected string }

	for _, tc := range []testCase{
		{url: "https://test.com/path1/path2/filename.zip?a=b&z=3", expected: "filename.zip"},
		{url: "https://test.com/path1/path2/filename.zip", expected: "filename.zip"},
		{url: "https://test.com/filename.zip?a=b&z=3", expected: "filename.zip"},
		{url: "https://test.com/filename.zip", expected: "filename.zip"},
		{url: "https://test.com", expected: ""},
	} {
		actual, err := webby.GetLastURLSegment(tc.url)
		if err != nil {
			t.Fatalf("Getting the last url segment for %s resulted in error '%s'", tc.url, err)
		}
		if actual != tc.expected {
			t.Fatalf("Expected to get '%s' from url '%s' but got %s", tc.expected, tc.url, actual)
		}
	}
}

func mockWebResponse(response string, statusCode int) *httptest.Server {
	bResponse := []byte(response)
	mockApiHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write(bResponse)
	}

	return httptest.NewServer(http.HandlerFunc(mockApiHandler))
}
