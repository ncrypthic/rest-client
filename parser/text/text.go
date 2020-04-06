package text

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	InvalidRequestErr   error = errors.New("Invalid request")
	InvalidParameterErr       = errors.New("Invalid parameter")
)

type Variable struct {
	URL       *url.URL
	Header    http.Header
	variables map[string]string
}

func (v *Variable) GetOrElse(key, value string) string {
	if _, ok := v.Header[key]; ok {
		return v.Header.Get(key)
	}
	return value
}

func Parse(data []byte) ([]*http.Request, []string, *Variable, error) {
	str := string(data)
	segments := skipEmptySegments(strings.Split(str, "--\n"))
	requests := make([]*http.Request, 0)
	var variable Variable
	var err error
	if len(segments) > 1 {
		lines := strings.Split(segments[0], "\n")
		variable, err = ExtractVariables(skipEmptyLine(lines))
		if err != nil {
			return nil, segments, &variable, err
		}
		segments = segments[1:]
	}
	for _, segment := range segments {
		lines := strings.Split(segment, "\n")
		if len(skipEmptyLine(lines)) == 0 {
			continue
		}
		lines, method, url, header, err := ExtractHttpRequest(variable, skipEmptyLine(lines))
		if err != nil {
			return nil, segments, &variable, err
		}
		body, err := ExtractHttpPayload(variable, skipEmptyLine(lines))
		requests = append(requests, &http.Request{
			Method: method,
			URL:    url,
			Header: header,
			Body:   body,
		})
		if err != nil {
			return nil, segments, &variable, err
		}
	}

	return requests, segments, &variable, nil
}

func ExtractVariables(lines []string) (Variable, error) {
	lines = skipEmptyLine(lines)

	lines, _, url, header, err := ExtractHttpRequest(Variable{}, skipEmptyLine(lines))
	if err != nil {
		return Variable{}, err
	}
	return Variable{URL: url, Header: header}, nil
}

func ExtractHttpRequest(variable Variable, lines []string) (leftLines []string, method string, reqURL *url.URL, header http.Header, err error) {
	hostname := ""
	if variable.URL != nil {
		hostname = variable.URL.Hostname()
	}
	port := ":80"
	credentials := ""
	scheme := "http"
	method = http.MethodGet
	path := ""
	leftLines = skipEmptyLine(lines)
	if len(leftLines) == 0 {
		return leftLines, "", nil, nil, nil
	}
	urlPattern := regexp.MustCompile("^http(s?):")
	if urlPattern.MatchString(trim(leftLines[0])) {
		reqURL, _ = url.Parse(trim(leftLines[0]))
		hostname = reqURL.Hostname()
		port = reqURL.Port()
		if reqURL.User.String() != "" {
			credentials = reqURL.User.String() + "@"
		}
		scheme = reqURL.Scheme
		if reqURL.Port() != "" {
			port = ":" + reqURL.Port()
		}
		leftLines = skipEmptyLine(leftLines[1:])
	} else if variable.URL != nil {
		hostname = variable.URL.Hostname()
		port = ":" + variable.URL.Port()
		credentials = variable.URL.User.String()
		scheme = variable.URL.Scheme
		path = variable.URL.RawPath
		leftLines = skipEmptyLine(leftLines)
	}
	leftLines, header, err = ExtractHttpHeaders(variable, leftLines)
	if len(leftLines) > 0 {
		verbs := strings.SplitN(trim(leftLines[0]), " ", 2)
		if len(verbs) == 1 {
			err := InvalidRequestErr
			return leftLines, "", nil, header, err
		}
		if extractHTTPVerb(verbs[0]) != "" {
			method = verbs[0]
			path = verbs[1]
			leftLines = leftLines[1:]
		}
	}
	urlString := fmt.Sprintf("%s://%s%s%s%s", scheme, credentials, hostname, port, applyVariables(variable, path))
	reqURL, err = url.Parse(urlString)
	if err != nil {
		return
	}
	return
}

func extractHTTPVerb(str string) string {
	switch strings.ToUpper(str) {
	case http.MethodGet:
		return http.MethodGet
	case http.MethodPost:
		return http.MethodPost
	case http.MethodPut:
		return http.MethodPut
	case http.MethodDelete:
		return http.MethodDelete
	case http.MethodHead:
		return http.MethodHead
	case http.MethodOptions:
		return http.MethodOptions
	case http.MethodConnect:
		return http.MethodConnect
	case http.MethodPatch:
		return http.MethodPatch
	case http.MethodTrace:
		return http.MethodTrace
	default:
		return ""
	}
}

func ExtractHttpHeaders(variable Variable, lines []string) (leftLines []string, headers http.Header, err error) {
	leftLines = skipEmptyLine(lines)
	headers = http.Header{}
	httpCommand := regexp.MustCompile("^(GET|POST|PUT|DELETE|HEAD|OPTION|PATCH) ")
	for {
		if len(leftLines) == 0 {
			return
		}
		line := leftLines[0]
		if httpCommand.MatchString(trim(line)) {
			break
		}
		paramSegments := strings.SplitN(line, ":", 2)
		if len(paramSegments) != 2 {
			err = InvalidParameterErr
			return
		}
		headers[trim(paramSegments[0])] = strings.Split(applyVariables(variable, trim(paramSegments[1])), ";")
		leftLines = skipEmptyLine(leftLines[1:])
	}

	return
}

func ExtractHttpPayload(variable Variable, lines []string) (io.ReadCloser, error) {
	buf := bytes.NewBuffer([]byte(strings.Join(lines, "\n")))
	return ioutil.NopCloser(buf), nil
}

func skipEmptyLine(lines []string) []string {
	nonEmptyLines := make([]string, 0)
	for idx, line := range lines {
		if strings.HasPrefix(trim(line), "#") {
			continue
		}
		if trim(line) != "" {
			nonEmptyLines = lines[idx:]
			break
		}
		idx = idx + 1
	}
	return nonEmptyLines
}

func trim(str string) string {
	if str == "" {
		return str
	}
	wsPrefixPattern := regexp.MustCompile(`^[[:space:]]*`)
	wsSuffixPattern := regexp.MustCompile(`[[:space:]]*$`)
	str = wsPrefixPattern.ReplaceAllString(str, "")
	str = wsSuffixPattern.ReplaceAllString(str, "")

	return str
}

func applyVariables(v Variable, str string) string {
	pattern := regexp.MustCompile(":[a-zA-Z0-9-_]+")
	processed := pattern.ReplaceAllStringFunc(str, func(match string) string {
		return v.GetOrElse(strings.TrimPrefix(match, ":"), ":"+match)
	})
	return processed
}

func skipEmptySegments(segments []string) []string {
	res := make([]string, 0)
	for _, segment := range segments {
		if segment == "" {
			continue
		}
		res = append(res, segment)
	}
	return res
}
