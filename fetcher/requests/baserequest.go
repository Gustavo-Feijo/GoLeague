package requests

import (
	"encoding/json"
	"fmt"
	"goleague/pkg/config"
	"goleague/pkg/messages"
	"log"
	"net/http"
	"net/url"
)

// AuthRequest make a authenticated request to the Riot API.
// Return the respose.
func AuthRequest(uri string, method string, params map[string]string) (*http.Response, error) {
	// Parse the URL.
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	// Add the query params to the url.
	query := u.Query()
	for key, val := range params {
		query.Add(key, val)
	}
	u.RawQuery = query.Encode()

	// Create the request for the given url.
	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return nil, err
	}

	if config.ApiKey == "" {
		panic("Can't do a authenticated request without the API Key.")
	}
	// Add the token from the .env.
	req.Header.Set("X-Riot-Token", config.ApiKey)
	client := &http.Client{}
	return client.Do(req)
}

// Request creates a simple request and return it.
func Request(url string, method string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return nil, err
	}
	client := &http.Client{}
	return client.Do(req)
}

// HandleAuthRequest works with generics to abstract the decoding process.
func HandleAuthRequest[T any](url string, method string, params map[string]string) (T, error) {
	var zero T
	resp, err := AuthRequest(url, method, params)
	if err != nil {
		return zero, fmt.Errorf(messages.RequestFailedMsg+": %w", url, err)
	}

	defer resp.Body.Close()

	// Check the status code.
	if resp.StatusCode != http.StatusOK {
		return zero, fmt.Errorf(messages.BadStatusCodeMsg, resp.StatusCode, url)
	}

	// Parse the match timeline.
	var respData T
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return zero, fmt.Errorf(messages.FailedToParseMsg+": %w", err)
	}

	// Return the timeline.
	return respData, nil
}
