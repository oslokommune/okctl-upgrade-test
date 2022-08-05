package migration

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/afero"
)

func copyFile(fs *afero.Afero, sourcePath string, destinationPath string) error {
	content, err := fs.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("reading: %w", err)
	}

	defer func() {
		_ = content.Close()
	}()

	err = fs.WriteReader(destinationPath, content)
	if err != nil {
		return fmt.Errorf("writing: %w", err)
	}

	return nil
}

func filenameWithoutExtension(filename string) string {
	return strings.Replace(path.Base(filename), path.Ext(filename), "", 1)
}
