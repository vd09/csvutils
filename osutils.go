package csvutils

import (
	"fmt"
	"os"
)

// openOrCreateFile opens the file for append if it exists, or creates a new file if it doesn't exist.
func openOrCreateFile(filePath string) (file *os.File, err error) {
	// Check if the file exists
	if fileExists(filePath) {
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

	return
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
