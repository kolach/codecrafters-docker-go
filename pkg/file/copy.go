package file

import (
	"io"
	"os"
)

func Copy(src, dst string) error {
	// Open the source file for reading
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Get the source file's permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}
	sourcePerm := sourceInfo.Mode()

	// Create the destination file for writing
	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// Copy the contents of the source file to the destination file
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	// Set the destination file's permissions to match the source file
	err = destinationFile.Chmod(sourcePerm)
	if err != nil {
		return err
	}

	return nil
}
