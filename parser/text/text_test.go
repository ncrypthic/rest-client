package text_test

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/ncrypthic/rest-client/parser/text"
)

func TestExtractHttpRequest(t *testing.T) {
	t.Run("ExtractHttpRequest: must be successful when given valid url", func(t *testing.T) {
		lines := []string{`http://example.com`}
		_, method, url, err := text.ExtractHttpRequest(text.Variable{}, lines)
		if err != nil {
			t.Fail()
		}
		if url == nil {
			t.Errorf("url must not nil")
		}
		if url.String() != lines[0] {
			t.Errorf("Expect http request url to be %s, got %s instead", lines[0], url.String())
		}
		if method != "GET" {
			t.Errorf("Expect http request method to be GET, got %s instead", method)
		}
	})
	t.Run("ExtractHttpRequest: must be successful when give valid request", func(t *testing.T) {
		lines := []string{`http://example.com`, "", `GET /health`}
		_, method, url, err := text.ExtractHttpRequest(text.Variable{}, lines)
		if err != nil {
			t.Fail()
		}
		if url == nil {
			t.Errorf("url must not nil")
		}
		if method != "GET" {
			t.Errorf("Expect http request method to be GET, got %s instead", method)
		}
		if url.String() != "http://example.com/health" {
			t.Errorf("Expect http request url to be %s, got %s instead", lines[0], url.String())
		}
		if url.Path != "/health" {
			t.Errorf("Expect http request path to be /health, got %s instead", url.Path)
		}
	})
}

func TestExtractHttpHeader(t *testing.T) {
	t.Run("ExtractHttpHeaders: must be successful when give valid headers parameter", func(t *testing.T) {
		lines := []string{"authorization: BEARER token", "", "user-agent: Chrome", "", "content-type: application/json", "", ""}
		_, header, err := text.ExtractHttpHeaders(text.Variable{}, lines)
		if err != nil {
			t.Fail()
		}
		if header == nil {
			t.Errorf("header must not nil")
		}
		if header.Get("authorization") != "BEARER token" {
			t.Errorf("HTTP authorization header value must be `BEARER token`, got `%s` instead", header.Get("authorization"))
		}
		if header.Get("user-agent") != "Chrome" {
			t.Errorf("HTTP authorization header value must be `BEARER token`, got `%s` instead", header.Get("authorization"))
		}
		if len(header) != 3 {
			t.Errorf("Header length must be 3, got `%d`", len(header))
		}
	})

}

func TestExtractHttpPayload(t *testing.T) {
	t.Run("ExtractHttpPayload: must be successful extract http body", func(t *testing.T) {
		jsonStr := `{
		    "type": "text",
		    "name": "example"
		}`
		jsonLines := strings.Split(jsonStr, "\n")
		jsonReader, err := text.ExtractHttpPayload(text.Variable{}, jsonLines)
		if err != nil {
			t.Fail()
		}
		jsonPayload, err := ioutil.ReadAll(jsonReader)
		if err != nil {
			t.Fail()
		}
		if string(jsonPayload) != jsonStr {
			t.Errorf("Expected body contain `%s`, got `%s`", jsonStr, string(jsonPayload))
		}
		formStr := `user=username&type=normal&length=1`
		formLines := []string{formStr}
		formReader, err := text.ExtractHttpPayload(text.Variable{}, formLines)
		if err != nil {
			t.Fail()
		}
		formPayload, err := ioutil.ReadAll(formReader)
		if err != nil {
			t.Fail()
		}
		if string(formPayload) != formStr {
			t.Errorf("Expected body contain `%s`, got `%s`", formStr, string(formPayload))
		}
	})
}

func TestParse(t *testing.T) {
	t.Run("Parse: must be successful given a requests configuration", func(t *testing.T) {
		config := `# Global Variable
		http://example.com

		token: xyz
		user_id: 123
		post_id: abcde12345
		--

		GET /home

		authorization: :token
		--

		http://example1.com

		PUT /users/:user_id/posts/:post_id

		authorization: BEARER :token

		{
		    "username": "test",
		    "password": "topsecret"
		}

		--

		http://example2.com

		POST /users/register

		{
		    "username": "test",
		    "password": "topsecret"
		}`
		requests, _, _, err := text.Parse([]byte(config))
		if err != nil {
			t.Fail()
		}
		if len(requests) != 3 {
			t.Errorf("Expected number of requests is 3, got %d", len(requests))
		}
		// requests.0
		if requests[0].URL.String() != "http://example.com/home" {
			t.Errorf("Expected request.0.URL to be `http://example.com/home`, got %s instead", requests[0].URL.String())
		}
		if requests[0].Method != http.MethodGet {
			t.Errorf("Expected request.0.Method to be `GET`, got %s instead", requests[0].Method)
		}
		if requests[0].Header.Get("authorization") != "xyz" {
			t.Errorf("Expected request.0.Header['Token'] to be `xyz`, got `%s` instead", requests[0].Header.Get("authorization"))
		}
		// requests.1
		if requests[1].URL.String() != "http://example1.com/users/123/posts/abcde12345" {
			t.Errorf("Expected request.1.URL to be `http://example1.com/users/123/posts/abcde12345`, got %s instead", requests[1].URL.String())
		}
		if requests[1].Method != http.MethodPut {
			t.Errorf("Expected request.1.Method to be `PUT`, got %s instead", requests[1].Method)
		}
		if requests[1].Header.Get("authorization") != "BEARER xyz" {
			t.Errorf("Expected request.1.Header['Token'] to be `BEARER xyz`, got `%s` instead", requests[1].Header.Get("authorization"))
		}
		// requests.2
		if requests[2].URL.String() != "http://example2.com/users/register" {
			t.Errorf("Expected request.2.URL to be `http://example2.com/users/register`, got %s instead", requests[2].URL.String())
		}
		if requests[2].Method != http.MethodPost {
			t.Errorf("Expected request.2.Method to be `POST`, got %s instead", requests[2].Method)
		}
	})
}
