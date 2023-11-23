package dockerhub

import (
	"context"
	"fmt"
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

	for _, layer := range manifest.Layers {
		fmt.Println("Pulling layer: ", layer.Digest)
		if err := pullAndUnpackLayer(ctx, token, img, layer.Digest, targetDir); err != nil {
			return err
		}
	}

	return nil
}
