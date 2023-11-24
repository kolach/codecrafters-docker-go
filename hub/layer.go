package hub

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	layerURL = "https://registry.hub.docker.com/v2/library/%s/blobs/%s"
)

func getLayer(
	ctx context.Context,
	token, img, digest string,
) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(layerURL, img, digest), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	req = req.WithContext(ctx)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("image manifest request failed: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download layer, status code: %d", res.StatusCode)
	}

	return res.Body, nil
}

func storeLayer(r io.ReadCloser, digest string, targetDir string) (string, error) {
	defer r.Close()

	filename := filepath.Join(targetDir, fmt.Sprintf("%s.tar.gz", digest))
	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create layer file: %w", err)
	}

	if _, err := io.Copy(file, r); err != nil {
		return "", fmt.Errorf("failed to write layer file: %w", err)
	}

	return filename, nil
}

func PullLayer(ctx context.Context, token, img, digest, targetDir string) (string, error) {
	layerReader, err := getLayer(ctx, token, img, digest)
	if err != nil {
		return "", err
	}
	return storeLayer(layerReader, digest, targetDir)
}

func UnpackLayer(filename string, targetDir string) error {
	cmd := exec.Command("tar", "-xzf", filename, "-C", targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to extract layer from file: %w", err)
	}
	return nil
}
