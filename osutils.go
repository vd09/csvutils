package csvutils

import (
	"fmt"
	"os"
)

// openOrCreateFile opens the file for append if it exists, or creates a new file if it doesn't exist.
func openOrCreateFile(filePath string) (*os.File, error) {
	var file *os.File

	// Check if the file exists
	if _, err := os.Stat(filePath); err == nil {
		// Open file for append
		file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open file for append: %w", err)
		}
	} else {
		// Create new file for write
		file, err = os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file: %w", err)
		}
	}

	return file, nil
}
