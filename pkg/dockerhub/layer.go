package dockerhub

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func pullAndUnpackLayer(ctx context.Context, token, img, digest, targetDir string) error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(layerURL, img, digest), nil)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}

	req = req.WithContext(ctx)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("image manifest request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download layer, status code: %d", res.StatusCode)
	}

	return unpackLayer(res.Body, targetDir)
}

func unpackLayer(r io.Reader, targetDir string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {
		case err == io.EOF:
			return nil // End of tarball
		case err != nil:
			return err // Handle other errors
		case header == nil:
			continue // No more headers
		}

		// Target location where the dir/file should be created
		target := filepath.Join(targetDir, header.Name)

		// Check the file type and handle accordingly
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory and preserve permissions
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			// Create file and write contents
			outFile, err := os.OpenFile(
				target,
				os.O_CREATE|os.O_RDWR|os.O_TRUNC,
				os.FileMode(header.Mode),
			)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
}
