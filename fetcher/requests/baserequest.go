package requests

import (
	"fmt"
	"goleague/pkg/config"
	"net/http"
)

// Do a authenticated request to the Riot API.
// Return the respose.
func AuthRequest(url string, method string) (*http.Response, error) {
	// Create the request for the given url.
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
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
		fmt.Println("Error creating request:", err)
		return nil, err
	}
	client := &http.Client{}
	return client.Do(req)
}
