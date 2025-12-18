package images

import (
	"log"
	"log/slog"
	"net/http"
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

	return filterAndSortImages(files, dir, checkAllowedImageFormats)
}

func filterAndSortImages(files []os.DirEntry, dir string, imageChecker func(string) bool) []string {
	var images []string
	for _, f := range files {
		if !f.IsDir() && imageChecker(filepath.Join(dir, f.Name())) {
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

func checkAllowedImageFormats(image string) bool {
	file, err := os.Open(image)
	if err != nil {
		slog.Info("Error opening file " + image + " error: " + err.Error())
		return false
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		slog.Info("Error reading file " + image + " error: " + err.Error())
		return false
	}

	contentType := http.DetectContentType(buffer)
	allowedTypes := []string{"image/jpeg", "image/png"}

	return slices.Contains(allowedTypes, contentType)
}
