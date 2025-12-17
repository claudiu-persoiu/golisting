package images

import (
	"io/fs"
	"os"
	"slices"
	"testing"
)

type mockDirEntry struct {
	name  string
	isDir bool
}

func (m mockDirEntry) Name() string {
	return m.name
}

func (m mockDirEntry) IsDir() bool {
	return m.isDir
}

func (m mockDirEntry) Type() fs.FileMode {
	panic("implement me")
}

func (m mockDirEntry) Info() (fs.FileInfo, error) {
	panic("implement me")
}

func TestFilterAndSortImages(t *testing.T) {

	input := []os.DirEntry{
		mockDirEntry{name: "img3.jpg"},
		mockDirEntry{name: "img10.png"},
		mockDirEntry{name: "img2.jpg"},
		mockDirEntry{name: "document.pdf"},
		mockDirEntry{name: "document.jpg", isDir: true},
	}

	output := filterAndSortImages(input)

	expected := []string{"img2.jpg", "img3.jpg", "img10.png"}

	if !slices.Equal(expected, output) {
		t.Errorf("Expected %v, but got %v", expected, output)
	}
}

func TestFilterOutImages(t *testing.T) {
	allImages := []string{"img1.jpg", "img2.png", "img3.jpg", "img4.png"}
	errorImages := []string{"img2.png", "img4.png"}

	output := FilterOutImages(allImages, errorImages)

	expected := []string{"img1.jpg", "img3.jpg"}

	if !slices.Equal(expected, output) {
		t.Errorf("Expected %v, but got %v", expected, output)
	}
}
