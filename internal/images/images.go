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

	return filterAndSortImages(files)
}

func filterAndSortImages(files []os.DirEntry) []string {

	var images []string
	for _, f := range files {
		ext := filepath.Ext(f.Name())
		if !f.IsDir() && (ext == ".jpg" || ext == ".png") {
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

func FilterOutImages(allImages, errorImages []string) []string {
	return slices.DeleteFunc(allImages, func(image string) bool {
		return slices.Contains(errorImages, image)
	})
}
