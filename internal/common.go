package internal

import (
	"embed"
	"html/template"
	"io"
	"log"
	"os"
)

type PageData struct {
	Images []string
	Path   string
}

func OutputHTML(templateBox embed.FS, file string, pageData *PageData, wr io.Writer) error {
	t, err := template.ParseFS(templateBox, "template/index.html")

	if err != nil {
		log.Fatal(err)
	}

	return t.Execute(wr, *pageData)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)

	return !os.IsNotExist(err)
}

func CreateDirIfNotExists(path string) error {
	return createDirIfNotExistsSilent(path, FileExists, os.Mkdir)
}

func createDirIfNotExistsSilent(path string, exists func(string) bool, mkdir func(string, os.FileMode) error) error {
	if !exists(path) {
		if err := mkdir(path, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
