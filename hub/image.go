package hub

import (
	"context"
	"fmt"
	"os"
)

func PullImage(ctx context.Context, img, ver, targetDir string) error {
	// get authentication token
	token, err := GetAuthToken(ctx, img)
	if err != nil {
		return err
	}

	// get manifest
	manifest, err := GetManifest(ctx, token, img, ver)
	if err != nil {
		return err
	}

	// make temporal directory for layers
	layersDir, err := os.MkdirTemp("", "layers")
	if err != nil {
		return fmt.Errorf("failed to create temp directory for layers: %w", err)
	}
	defer os.RemoveAll(layersDir)

	layerFiles, err := manifest.PullLayers(ctx, token, img, layersDir)

	for _, layer := range manifest.Layers {
		layerFile := layerFiles[layer.Digest]
		if err := UnpackLayer(layerFile, targetDir); err != nil {
			return err
		}
	}

	return nil
}
