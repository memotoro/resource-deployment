package clients

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Client - basic client for http calls
type Client interface {
	BaseEndpoint() string
	ExecuteCall(req *http.Request) (*http.Response, []byte, error)
}

// Auth - basic http authentication
type Auth struct {
	Token string
}

type client struct {
	httpClient http.Client
	host       string
	port       string
	auth       Auth
	timeout    time.Duration
	verbose    bool
}

// NewHTTPClient returns a Client for future http calls
func NewHTTPClient(host, port string, timeout time.Duration, auth Auth, verbose bool) Client {
	return &client{httpClient: http.Client{}, host: host, port: port, auth: auth, timeout: timeout, verbose: verbose}
}

// BaseEndpoint returns base endpoint for http calls
func (c *client) BaseEndpoint() string {
	return fmt.Sprintf("https://%s:%s", c.host, c.port)
}

// ExecuteCall executes the http call with a given request
func (c *client) ExecuteCall(req *http.Request) (*http.Response, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req = req.WithContext(ctx)

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	if c.auth.Token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.auth.Token))
	}

	req.Header.Add("Content-type", "application/json")

	if c.verbose {
		log.Printf("REQUEST METHOD [%s] URL [%s]", req.Method, req.URL)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	if c.verbose {
		log.Printf("RESPONSE [%v]", string(data))
	}

	return resp, data, nil
}
