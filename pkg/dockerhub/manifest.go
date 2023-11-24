package dockerhub

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
)

func getManifestList(ctx context.Context, token, img, ver string) (*manifestList, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(manifestURL, img, ver), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	req = req.WithContext(ctx)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", manifestListAcceptHeader)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("image manifest request failed: %w", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image manifest body: %w", err)
	}

	// fmt.Printf("Got response body: %s\n\n", string(body))

	var manifest manifestList
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, fmt.Errorf("could not parse image manifest: %w", err)
	}

	return &manifest, nil
}

func findRuntimeMeta(list *manifestList) (*manifestMeta, error) {
	for _, m := range list.Manifests {
		if m.Platform.Os == runtime.GOOS && m.Platform.Architecture == runtime.GOARCH {
			// fmt.Printf(
			// 	"Found manifest for OS: %s, ARCH: %s\n",
			// 	m.Platform.Os,
			// 	m.Platform.Architecture,
			// )
			return m, nil
		}
	}
	return nil, fmt.Errorf("digest not found")
}

func getManifestByMetadata(
	ctx context.Context,
	token, img string,
	meta *manifestMeta,
) (*manifest, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(manifestURL, img, meta.Digest), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make manifest request: %w", err)
	}

	req = req.WithContext(ctx)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", meta.MediaType)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("image manifest request failed: %w", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image manifest body: %w", err)
	}

	// fmt.Printf("Got response body: %s\n\n", string(body))

	var manifest manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, fmt.Errorf("could not parse image manifest: %w", err)
	}

	return &manifest, nil
}

func getManifest(ctx context.Context, token, img, ver string) (*manifest, error) {
	// Get manifest list
	list, err := getManifestList(ctx, token, img, ver)
	if err != nil {
		return nil, err
	}

	// Find manifest in list that matches runtime platform
	meta, err := findRuntimeMeta(list)
	if err != nil {
		return nil, err
	}

	// Fetch manifest and return
	return getManifestByMetadata(ctx, token, img, meta)
}
