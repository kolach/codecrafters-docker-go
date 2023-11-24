package dockerhub

import (
	"context"
	"fmt"
	"os"
)

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
