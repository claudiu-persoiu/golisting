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
	_ , err := os.Stat(path)

	return !os.IsNotExist(err)
}