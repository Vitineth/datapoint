package datapoint

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

// Supplier is a simple interface for something that returns a type. This can be used to abstract over
// backing storage in an independent way
type Supplier[T any] interface {
	Get() T
}

// DataPointClient represents a client used for interacting with the MetOffice DataPoint service
type DataPointClient struct {
	apiKeySupplier *Supplier[string]
	baseUrl        string
	httpClient     *http.Client
}

// Opt is an option that can apply to a DataPointClient
type Opt interface {
	apply(client *DataPointClient)
}

type baseUriOpt struct {
	uri string
}

func (b baseUriOpt) apply(client *DataPointClient) {
	client.baseUrl = b.uri
}

// WithBaseURI will update the data point client to use a different base URI from the standard. This can be used to
// to a simple proxy if required
func WithBaseURI(uri string) Opt {
	return baseUriOpt{uri: uri}
}

type stringSupplier struct {
	value string
}

func (s stringSupplier) Get() string {
	return s.value
}

type apiKey struct {
	key string
}

func (a apiKey) apply(client *DataPointClient) {
	var sup Supplier[string] = stringSupplier{value: a.key}
	client.apiKeySupplier = &sup
}

// WithApiKey sets the API key of the client to the string provided, this cannot be updated during use
func WithApiKey(key string) Opt {
	return apiKey{key: key}
}

type apiSupplier struct {
	supplier Supplier[string]
}

func (a apiSupplier) apply(client *DataPointClient) {
	client.apiKeySupplier = &a.supplier
}

// WithApiSupplier sets the API key to this supplier, this will be called on each invocation to the DataPoint API
func WithApiSupplier(supplier Supplier[string]) Opt {
	return apiSupplier{supplier: supplier}
}

type httpClient struct {
	client *http.Client
}

func (h httpClient) apply(client *DataPointClient) {
	client.httpClient = h.client
}

// WithHttpClient replaces the HTTP client which will be used for making requests to the service. This can be used to
// configure a real proxy, or set specific timeouts, etc
func WithHttpClient(client *http.Client) Opt {
	return httpClient{client: client}
}

// NewClient returns a new DataPointClient, applying all the options set. If an API key is not provided, this will fail
// and return an error. Look for With functions for the options which can be provided
func NewClient(opt ...Opt) (*DataPointClient, error) {
	client := DataPointClient{
		baseUrl:    "http://datapoint.metoffice.gov.uk/public/data/",
		httpClient: &http.Client{},
	}

	for _, o := range opt {
		o.apply(&client)
	}

	if client.apiKeySupplier == nil {
		return nil, errors.New("no api key provided")
	}

	return &client, nil
}

func (d *DataPointClient) fetch(description string, suffix string, params map[string]string) ([]byte, string, error) {
	target, err := url.JoinPath(d.baseUrl, suffix)
	if err != nil {
		return nil, "???", fmt.Errorf("failed to generate %v url: %w", description, err)
	}

	withKey := target + "?key=" + (*d.apiKeySupplier).Get()
	if params != nil {
		for k, v := range params {
			withKey += "&" + url.QueryEscape(k) + "=" + url.QueryEscape(v)
		}
	}

	r, err := d.httpClient.Get(withKey)
	if err != nil {
		return nil, target, fmt.Errorf("failed to query %v for %v: %w", target, description, err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slog.Warn("failed to close body from query", "err", err, "desc", description)
		}
	}(r.Body)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, target, fmt.Errorf("failed to read body from response from %v for %v: %w", target, description, err)
	}

	return body, target, nil
}
