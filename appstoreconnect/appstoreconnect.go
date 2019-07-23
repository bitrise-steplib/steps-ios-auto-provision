package appstoreconnect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/bitrise-io/bitrise-add-new-project/httputil"
	"github.com/bitrise-io/go-utils/log"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/go-querystring/query"
)

const (
	baseURL = "https://api.appstoreconnect.apple.com/"
)

type service struct {
	client *Client
}

// Client communicate with the Apple API
type Client struct {
	keyID             string
	issuerID          string
	privateKeyContent []byte

	token       *jwt.Token
	signedToken string

	client  *http.Client
	BaseURL *url.URL

	common       service // Reuse a single struct instead of allocating one for each service on the heap.
	Provisioning *ProvisioningService
}

// NewClient creates a new client
func NewClient(keyID, issuerID string, privateKey []byte) (*Client, error) {
	httpClient := http.DefaultClient
	baseURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	c := &Client{
		keyID:             keyID,
		issuerID:          issuerID,
		privateKeyContent: privateKey,

		client:  httpClient,
		BaseURL: baseURL,
	}
	c.common.client = c
	c.Provisioning = (*ProvisioningService)(&c.common)

	return c, nil
}

// ensureSignedToken makes sure that the JWT auth token is not expired
// and return a signed key
func (c *Client) ensureSignedToken() (string, error) {
	if c.token != nil {
		claim, ok := c.token.Claims.(claims)
		if !ok {
			return "", fmt.Errorf("failed to cast claim for token")
		}
		expiration := time.Unix(int64(claim.Expiration), 0)
		// token is marked valid if it will not expire in the upcoming 60 sec
		if expiration.After(time.Now().Add(60 * time.Second)) {
			return c.signedToken, nil
		}
	}

	c.token = createToken(c.keyID, c.issuerID)
	var err error
	if c.signedToken, err = signToken(c.token, c.privateKeyContent); err != nil {
		return "", err
	}
	return c.signedToken, nil
}

// NewRequest creates a new http.Request
func (c *Client) NewRequest(method, endpoint string, body interface{}) (*http.Request, error) {
	endpoint = "v1/" + endpoint
	u, err := c.BaseURL.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	signedToken, err := c.ensureSignedToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+signedToken)

	return req, nil
}

func checkResponse(r *http.Response) error {
	if r.StatusCode >= 200 && r.StatusCode <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		if err := json.Unmarshal(data, errorResponse); err != nil {
			log.Errorf("Failed to unmarshal response, error: %s", err)
		}
	}
	return errorResponse
}

// Do ...
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	log.Debugf("Request:")
	if err := httputil.PrintRequest(req); err != nil {
		log.Debugf("Failed to print request: %s", err)
	}

	resp, err := c.client.Do(req)

	log.Debugf("Response:")
	if err := httputil.PrintResponse(resp); err != nil {
		log.Debugf("Failed to print response: %s", err)
	}

	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Warnf("failed to close response body, error: %s", cerr)
		}
	}()

	if err := checkResponse(resp); err != nil {
		return resp, err
	}

	if v != nil {
		decErr := json.NewDecoder(resp.Body).Decode(v)
		if decErr == io.EOF {
			decErr = nil // ignore EOF errors caused by empty response body
		}
		if decErr != nil {
			err = decErr
		}
	}

	return resp, err
}

// addOptions adds the parameters in opt as URL query parameters to s. opt
// must be a struct whose fields may contain "url" tags.
func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}
