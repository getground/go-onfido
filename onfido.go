package onfido

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	// "io/ioutil"

	"github.com/tomnomnom/linkheader"
)

// Constants
const (
	ClientVersion   = "0.1.0"
	DefaultEndpoint = "https://api.onfido.com/v2"
	TokenEnv        = "ONFIDO_TOKEN"
)

// Client represents an Onfido API client
type Client struct {
	Endpoint   string
	HTTPClient HTTPRequester
	Token      Token
}

// HTTPRequester represents an HTTP requester
type HTTPRequester interface {
	Do(*http.Request) (*http.Response, error)
}

// Error represents an Onfido API error response
type Error struct {
	Resp *http.Response
	Err  struct {
		ID     string              `json:"id,omitempty"`
		Type   string              `json:"type,omitempty"`
		Msg    string              `json:"message,omitempty"`
		Fields map[string][]string `json:"fields,omitempty"`
	} `json:"error"`
}

func (e *Error) Error() string {
	if e.Err.Msg != "" {
		return e.Err.Msg
	}
	if e.Resp != nil {
		return fmt.Sprintf("http request failed with status code %d", e.Resp.StatusCode)
	}
	return "an unknown error occurred"
}

// Token is an Onfido authentication token
type Token string

// String returns the token as a string.
func (t Token) String() string {
	return string(t)
}

// Prod checks if this is a production token or not.
func (t Token) Prod() bool {
	return !strings.HasPrefix(string(t), "test_")
}

// NewClientFromEnv creates a new Onfido client using configuration
// from environment variables.
func NewClientFromEnv() (*Client, error) {
	token := os.Getenv(TokenEnv)
	if token == "" {
		return nil, fmt.Errorf("onfido token not found in environmental variable `%s`", TokenEnv)
	}
	return NewClient(token), nil
}

// NewClient creates a new Onfido client.
func NewClient(token string) *Client {
	return &Client{
		Endpoint:   DefaultEndpoint,
		HTTPClient: http.DefaultClient,
		Token:      Token(token),
	}
}

func (c *Client) newRequest(method, uri string, body io.Reader) (*http.Request, error) {
	if !strings.HasPrefix(uri, "http") {
		if !strings.HasPrefix(uri, "/") {
			uri = "/" + uri
		}
		uri = c.Endpoint + uri
	}

	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Go-Onfido/"+ClientVersion)
	req.Header.Set("Authorization", "Token token="+c.Token.String())
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	req.WithContext(ctx)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return nil, err
		}
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if c := resp.StatusCode; c < 200 || c > 299 {
		fmt.Println("+++++++++++++++++++++++++++++++")
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		fmt.Println(bodyString)
		fmt.Println("+++++++++++++++++++++++++++++++")
		return nil, handleResponseErr(resp)
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			// fmt.Println("io copy")
			_, err = io.Copy(w, resp.Body)
		} else if isJSONResponse(resp) {
			// fmt.Println("is json resp")
			// data, err := ioutil.ReadAll(resp.Body)
			// if err != nil {
			// 	fmt.Printf("err in do json is %v", err)
			// 	return resp, err
			// }
			// err = json.Unmarshal(data, v)
			err = json.NewDecoder(resp.Body).Decode(v)
		} else {
			err = fmt.Errorf("unable to parse respose body into %T", v)
		}
	}

	// if err != nil {
	// 	bodyText, _ := ioutil.ReadAll(resp.Body)
	// 	fmt.Printf("Body is %v", string(bodyText))
	// 	fmt.Printf("Status code is %d\n", resp.StatusCode)
	//
	//
	// }

	// err = json.NewDecoder(resp.Body).Decode(v)
	// fmt.Printf("err in do is %v", err)
	// var bodyMap map[string]interface{}
	// err = json.Unmarshal(body, &bodyMap)
	// if err != nil {
	//   return err
	// }

	return resp, err
}

func (c *Client) download(ctx context.Context, req *http.Request) ([]byte, error) {
	req.WithContext(ctx)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return nil, err
		}
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if c := resp.StatusCode; c < 200 || c > 299 {
		return nil, handleResponseErr(resp)
	}

	return ioutil.ReadAll(resp.Body)
}

func isJSONResponse(resp *http.Response) bool {
	return strings.Contains(resp.Header.Get("Content-Type"), "application/json")
}

func handleResponseErr(resp *http.Response) error {
	var onfidoErr Error
	if resp.Body != nil && isJSONResponse(resp) {
		defer resp.Body.Close()

		// bodyText, _ := ioutil.ReadAll(resp.Body)
		// fmt.Printf("Body is %v", string(bodyText))

		if err := json.NewDecoder(resp.Body).Decode(&onfidoErr); err != nil {
			fmt.Printf("Error decoding Onfido error: %v\n", err)
			return err
		}
	} else {
		onfidoErr = Error{}
	}
	onfidoErr.Resp = resp

	fmt.Printf("Onfido error:\n%+v\n", onfidoErr)

	return &onfidoErr
}

type iter struct {
	c       *Client
	nextURL string
	handler iterHandler

	values []interface{}
	cur    interface{}
	err    error
}

type iterHandler func(body []byte) ([]interface{}, error)

func (it *iter) Current() interface{} {
	return it.cur
}

func (it *iter) Err() error {
	return it.err
}

func (it *iter) Next(ctx context.Context) bool {
	if it.err != nil {
		return false
	}
	if len(it.values) == 0 && it.nextURL != "" {
		req, err := it.c.newRequest("GET", it.nextURL, nil)
		if err != nil {
			it.err = err
			return false
		}

		var body bytes.Buffer
		resp, err := it.c.do(ctx, req, &body)
		if err != nil {
			it.err = err
			return false
		}
		if !isJSONResponse(resp) {
			it.err = errors.New("non json response")
			return false
		}

		values, err := it.handler(body.Bytes())
		if err != nil {
			it.err = err
			return false
		}
		it.values = values

		links := linkheader.Parse(resp.Header.Get("Link"))
		links = links.FilterByRel("next")
		if len(links) > 0 {
			it.nextURL = links[0].URL
		} else {
			it.nextURL = ""
		}
	}
	if len(it.values) == 0 {
		return false
	}

	it.cur = it.values[0]
	it.values = it.values[1:]
	return true
}
