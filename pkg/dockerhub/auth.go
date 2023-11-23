package dockerhub

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func getAuthToken(ctx context.Context, img string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(tokenURL, img), nil)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}

	req = req.WithContext(ctx)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var tokenResponse tokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", fmt.Errorf("could not parse response body: %w", err)
	}

	return tokenResponse.Token, nil
}
