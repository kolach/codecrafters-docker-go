package hub

import (
	"context"
	"fmt"
	"os"
	"sync"
)

type dlRes struct {
	digest string
	path   string
}

func parallelPull(
	ctx context.Context,
	man *manifest,
	token, img, targetDir string,
) (<-chan dlRes, <-chan error) {
	errChan := make(chan error, len(man.Layers))
	dlDone := make(chan dlRes, len(man.Layers))

	go func() {
		pullCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		var wg sync.WaitGroup

		for _, layer := range man.Layers {
			layer := layer
			wg.Add(1)
			go func() {
				defer wg.Done()

				layerFile, err := pullAndStoreLayer(pullCtx, token, img, layer.Digest, targetDir)
				if err != nil {
					errChan <- err
					cancel()
				} else {
					dlDone <- dlRes{digest: layer.Digest, path: layerFile}
				}
			}()
		}

		wg.Wait()
		close(errChan)
		close(dlDone)
	}()

	return dlDone, errChan
}

func PullImage(ctx context.Context, img, ver, targetDir string) error {
	// get authentication token
	token, err := getAuthToken(ctx, img)
	if err != nil {
		return err
	}

	// get manifest
	manifest, err := getManifest(ctx, token, img, ver)
	if err != nil {
		return err
	}

	// make temporal directory for layers
	layersDir, err := os.MkdirTemp("", "layers")
	if err != nil {
		return fmt.Errorf("failed to create temp directory for layers: %w", err)
	}
	defer os.RemoveAll(layersDir)

	for _, layer := range manifest.Layers {
		layerReader, err := pullLayer(ctx, token, img, layer.Digest)
		if err != nil {
			return err
		}

		layerFile, err := storeLayer(layerReader, layer.Digest, layersDir)
		if err != nil {
			return err
		}

		if err := unpackLayerFromFile(layerFile, targetDir); err != nil {
			return err
		}
	}

	return nil
}

func PullImage2(ctx context.Context, img, ver, targetDir string) error {
	// get authentication token
	token, err := getAuthToken(ctx, img)
	if err != nil {
		return err
	}

	// get manifest
	manifest, err := getManifest(ctx, token, img, ver)
	if err != nil {
		return err
	}

	// make temporal directory for layers
	layersDir, err := os.MkdirTemp("", "layers")
	if err != nil {
		return fmt.Errorf("failed to create temp directory for layers: %w", err)
	}
	defer os.RemoveAll(layersDir)

	layerFileChan, errChan := parallelPull(ctx, manifest, token, img, layersDir)
	for err := range errChan {
		return err
	}

	m := map[string]string{}
	for download := range layerFileChan {
		m[download.digest] = download.path
	}

	for _, layer := range manifest.Layers {
		layerFile := m[layer.Digest]
		if err := unpackLayerFromFile(layerFile, targetDir); err != nil {
			return err
		}
	}

	return nil
}
