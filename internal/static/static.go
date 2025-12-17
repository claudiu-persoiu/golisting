package static

import (
	"embed"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/claudiu-persoiu/golisting/internal"
)

func OutputFiles(publicBox, templateBox embed.FS, path *string, images []string) {
	data := &internal.PageData{Images: images}

	targetPath := filepath.Join(*path, "index.html")

	out, err := os.Create(targetPath)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	if err := internal.OutputHTML(templateBox, "index.html", data, out); err != nil {
		log.Fatal(err)
	}

	if err := copyStaticFiles(publicBox, path); err != nil {
		log.Fatal(err)
	}
}

func copyStaticFiles(publicBox embed.FS, targetPath *string) error {
	staticTarget := filepath.Join(*targetPath, "public")
	if internal.FileExists(staticTarget) {
		if err := os.Mkdir(staticTarget, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	return fs.WalkDir(publicBox, ".", func(fileName string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if fileName == "." {
			return nil
		}

		target := filepath.Join(*targetPath, fileName)

		if d.IsDir() {
			if !internal.FileExists(target) {
				if err := os.Mkdir(target, os.ModePerm); err != nil {
					return err
				}
			}
		} else {
			out, err := os.Create(target)
			if err != nil {
				return err
			}
			defer out.Close()

			in, err := os.Open(fileName)
			if err != nil {
				return err
			}
			defer in.Close()

			_, err = io.Copy(out, in)
			if err != nil {
				return err
			}

			return out.Sync()
		}

		return nil
	})
}
