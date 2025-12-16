package images

import (
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func GetImages(dir string) []string {

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	var images []string
	for _, f := range files {
		ext := filepath.Ext(f.Name())
		if ext == ".jpg" || ext == ".png" {
			images = append(images, f.Name())
		}
	}

	slices.SortFunc(images, func(a, b string) int {
		if len(a) != len(b) {
			return len(a) - len(b)
		}

		return strings.Compare(a, b)
	})

	return images
}
