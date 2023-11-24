package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sync"
)

const (
	manifestURL              = "https://registry-1.docker.io/v2/library/%s/manifests/%s"
	manifestListAcceptHeader = "application/vnd.docker.distribution.manifest.list.v2+json"
)

type Manifest struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Config        struct {
		MediaType string `json:"mediaType"`
		Size      int    `json:"size"`
		Digest    string `json:"digest"`
	} `json:"config"`
	Layers []struct {
		MediaType string `json:"mediaType"`
		Size      int    `json:"size"`
		Digest    string `json:"digest"`
	} `json:"layers"`
}

type dlRes struct {
	digest string
	path   string
}

func (manifest *Manifest) parallelPull(
	ctx context.Context,
	token, img, targetDir string,
) (<-chan dlRes, <-chan error) {
	errChan := make(chan error, len(manifest.Layers))
	resChan := make(chan dlRes, len(manifest.Layers))

	go func() {
		defer close(errChan)
		defer close(resChan)

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var wg sync.WaitGroup

		for _, layer := range manifest.Layers {
			wg.Add(1)
			go func(digest string) {
				defer wg.Done()
				path, err := PullLayer(ctx, token, img, digest, targetDir)
				if err == nil {
					resChan <- dlRes{digest: digest, path: path}
				} else {
					errChan <- err
					cancel()
				}
			}(layer.Digest)
		}
		wg.Wait()
	}()

	return resChan, errChan
}

func (manifest *Manifest) PullLayers(ctx context.Context,
	token, img, targetDir string,
) (map[string]string, error) {
	layerFileChan, errChan := manifest.parallelPull(ctx, token, img, targetDir)

	for err := range errChan {
		return nil, err
	}

	layers := map[string]string{}
	for download := range layerFileChan {
		layers[download.digest] = download.path
	}

	return layers, nil
}

type ManifestMeta struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
	Platform  struct {
		Architecture string `json:"architecture"`
		Os           string `json:"os"`
	} `json:"platform"`
}

type ManifestList struct {
	SchemaVersion int             `json:"schemaVersion"`
	MediaType     string          `json:"mediaType"`
	Manifests     []*ManifestMeta `json:"manifests"`
}

func GetManifestList(ctx context.Context, token, img, ver string) (*ManifestList, error) {
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

	var manifest ManifestList
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, fmt.Errorf("could not parse image manifest: %w", err)
	}

	return &manifest, nil
}

func findRuntimeMeta(list *ManifestList) (*ManifestMeta, error) {
	for _, m := range list.Manifests {
		if m.Platform.Os == runtime.GOOS && m.Platform.Architecture == runtime.GOARCH {
			return m, nil
		}
	}
	return nil, fmt.Errorf("digest not found")
}

func getManifestByMetadata(
	ctx context.Context,
	token, img string,
	meta *ManifestMeta,
) (*Manifest, error) {
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

	var manifest Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, fmt.Errorf("could not parse image manifest: %w", err)
	}

	return &manifest, nil
}

func GetManifest(ctx context.Context, token, img, ver string) (*Manifest, error) {
	// Get manifest list
	list, err := GetManifestList(ctx, token, img, ver)
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
