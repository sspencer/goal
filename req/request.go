package req

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

const (
	// JSONContentType is the http content type for json
	JSONContentType = "application/json"

	// URLEncodededContentType is the http content type for form urlencoded data
	URLEncodededContentType = "application/x-www-form-urlencoded"
)

// From a TJ Holowaychuk tweet:
//     TIL Go json syntax errors give you the offset, so you
//     can provide more context if you want, the single char
//     gets a little confusing
//
// see SyntaxError and Unmarshall below
type SyntaxError struct {
	*json.SyntaxError
	input []byte
}

// RequestFunc allows variable numbers of args in New to configure requests.
// For example:
//   r0 := req.New()
//   r1 := req.New(req.Curl(true), req.SkipRedirects(true))
//   r2 := req.New(req.CurlHeader(true))
type RequestFunc func(*Request)

// Request is used to set some configuration options on the HTTP request.
type Request struct {
	curl          bool
	curlHeader    bool
	timeout       time.Duration
	skipRedirects bool
}

// New creates a new Request struct.  Defaults are:
//   curl (body): false
//   curl header (and body): false
//   timeout: 30 seconds
//   skip redirects: false
func New(options ...func(*Request)) *Request {
	r := &Request{}
	r.curl = false
	r.curlHeader = false
	r.timeout = 30 * time.Second
	r.skipRedirects = false

	for _, opt := range options {
		opt(r)
	}

	return r
}

// Curl enables or disables Curl logging
func Curl(b bool) RequestFunc { return func(r *Request) { r.curl = b } }

// CurlHeader enables or disables Curl logging with extra Header printout
func CurlHeader(b bool) RequestFunc { return func(r *Request) { r.curlHeader = b } }

// Timeout changes the default request timeout (30 seconds)
func Timeout(d time.Duration) RequestFunc { return func(r *Request) { r.timeout = d } }

// SkipRedirects enables or disables the skip redirects directive
func SkipRedirects(b bool) RequestFunc { return func(r *Request) { r.skipRedirects = b } }

// IsSuccess returns TRUE if the status code is 2XX
func IsSuccess(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode <= http.StatusIMUsed
}

// Error implements the Error method for SyntaxErrors
func (e SyntaxError) Error() string {
	return fmt.Sprintf("syntax error near: `%s`", string(e.input[e.Offset-1:]))
}

// Unmarshal unmarshals a successful http response (and closes it)
func Unmarshal(body io.ReadCloser, v interface{}) error {
	defer body.Close()
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &v)

	if e, ok := err.(*json.SyntaxError); ok {
		return SyntaxError{e, data}
	}

	return err
}

// Get performs a HTTP GET
func (c *Request) Get(url string) (*http.Response, error) {
	return c.request(http.MethodGet, url, "", nil)
}

// Get performs a HTTP HEAD
func (c *Request) Head(url string) (*http.Response, error) {
	return c.request(http.MethodHead, url, "", nil)
}

// Get performs a HTTP DELETE
func (c *Request) Delete(url string) (*http.Response, error) {
	return c.request(http.MethodDelete, url, "", nil)
}

// Get performs a HTTP POST
func (c *Request) Post(url string, values url.Values) (*http.Response, error) {
	return c.request(http.MethodPost, url, URLEncodededContentType, strings.NewReader(values.Encode()))
}

// Get performs a HTTP PUT
func (c *Request) Put(url string, values url.Values) (*http.Response, error) {
	return c.request(http.MethodPost, url, URLEncodededContentType, strings.NewReader(values.Encode()))
}

// request does all the work of the above HTTP method functions
func (c *Request) request(method, url, contentType string, data io.Reader) (*http.Response, error) {

	var buf bytes.Buffer
	var err error
	var req *http.Request

	if data != nil {
		tee := io.TeeReader(data, &buf) // TeeRequest for curl output
		req, err = http.NewRequest(method, url, tee)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	client := &http.Client{Timeout: c.timeout}
	if c.skipRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return errors.New("Skip redirects")
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// TeeBuffer is empty until request is sent. Data is copied to writer as it is read.
	if c.curl || c.curlHeader {
		c.logger(req, resp, &buf)
	}

	if resp.StatusCode >= http.StatusOK && resp.StatusCode <= http.StatusIMUsed {
		return resp, nil
	}

	// NOT OK - return error body
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("Error making HTTP request.  HTTP Status %d: %v", resp.StatusCode, string(body))
}

func (c *Request) logger(r *http.Request, resp *http.Response, data io.Reader) {
	if !c.curl && !c.curlHeader {
		return
	}

	indent := strings.Repeat(" ", 4)

	// -s silences output (progress meter and errors)
	// -S "unsilences" errors
	buf := bytes.NewBufferString("\ncurl -sS")

	if c.skipRedirects {
		buf.WriteString(" -L")
	}

	buf.WriteString(" -X")
	buf.WriteString(r.Method) // GET, POST, PUT, etc.
	buf.WriteString(" \\\n")

	for n, v := range r.Header {
		buf.WriteString(indent)
		buf.WriteString("-H'")
		buf.WriteString(n)
		buf.WriteString(": ")
		buf.WriteString(v[0])
		buf.WriteString("' \\\n")
	}

	if data != nil {
		b, err := ioutil.ReadAll(data)
		if err == nil {
			str := string(b)
			if str != "" {
				buf.WriteString(indent)
				buf.WriteString("-d'")
				buf.WriteString(strings.TrimSpace(str))
				buf.WriteString("' \\\n")
			}
		}
	}

	// curl -XGET ...
	//     "<THE URL>"   <--
	buf.WriteString(indent)
	buf.WriteString("\"")
	buf.WriteString(r.URL.String())
	buf.WriteString("\"")

	// that's it for the actual curl command,
	// now log the response
	if dump, err := httputil.DumpResponse(resp, true); err == nil {
		// split header from body
		parts := bytes.SplitN(dump, []byte("\r\n\r\n"), 2)

		if len(parts) > 1 {
			header := parts[0]
			body := parts[1]

			buf.WriteString("\n\n")
			if c.curlHeader {
				buf.WriteString(string(header))
				buf.WriteString("\n\n")
			} else {
				buf.WriteString(resp.Proto) // e.g. "HTTP/1.0"
				buf.WriteString(" ")
				buf.WriteString(resp.Status) // e.g. "200 OK"
				buf.WriteString("\n")
			}

			if json, err := indentJSON(body); err != nil {
				buf.WriteString(string(body))
			} else {
				buf.WriteString(string(json))
			}
			buf.WriteString("\n")
		}
	}

	// are we just logging this ?
	log.Println(buf.String())
}

func indentJSON(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "   ")
	return out.Bytes(), err
}
