package requests

import (
	"goleague/pkg/config"
	"log"
	"net/http"
	"net/url"
)

// Do a authenticated request to the Riot API.
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
	// Add the token from the .env
	req.Header.Set("X-Riot-Token", config.ApiKey)
	client := &http.Client{}
	return client.Do(req)
}

// Create a simple request and return it.
func Request(url string, method string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return nil, err
	}
	client := &http.Client{}
	return client.Do(req)
}
