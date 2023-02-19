package main

import (
	"embed"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"text/template"

	"github.com/nfnt/resize"
)

type PageData struct {
	Images []string
	Path   string
}

type byName []os.FileInfo

func (s byName) Len() int {
	return len(s)
}

func (s byName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byName) Less(i, j int) bool {
	if len(s[i].Name()) != len(s[j].Name()) {
		return len(s[i].Name()) < len(s[j].Name())
	}

	return s[i].Name() < s[j].Name()
}

//go:embed public
var publicBox embed.FS

//go:embed template
var templateBox embed.FS

func main() {

	path := flag.String("path", "./", "path to images")
	address := flag.String("address", ":5000", "server address")
	output := flag.Bool("output", false, "output location if html is generated and server will not start")
	flag.Parse()

	images := generateThumbs(path)

	if !*output {
		startSever(publicBox, templateBox, path, &images, address)
	} else {
		outputFiles(publicBox, templateBox, path, &images)
	}
}

func generateThumbs(sourcePath *string) []string {

	dir := filepath.Join(".", *sourcePath)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	targetPath := filepath.Join(dir, ".thumb")
	if _, err := os.Stat(targetPath); err != nil {
		if err := os.Mkdir(targetPath, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	fileNames := make(chan string, 1000)

	var waitgroup sync.WaitGroup

	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		go func() {
			for j := range fileNames {
				createThumbs(dir, j)
				waitgroup.Done()
			}
		}()
	}

	var images []string

	sort.Sort(byName(files))

	for _, f := range files {
		ext := filepath.Ext(f.Name())
		if ext == ".jpg" || ext == ".png" {
			waitgroup.Add(1)
			fmt.Println(f.Name())
			images = append(images, f.Name())
			fileNames <- f.Name()
		}
	}

	close(fileNames)

	waitgroup.Wait()

	runtime.GC()

	return images
}

func startSever(publicBox, templateBox embed.FS, path *string, images *[]string, address *string) {
	http.Handle("/public/", http.FileServer(http.FS(publicBox)))
	http.Handle("/photo/", http.StripPrefix("/photo/", http.FileServer(http.Dir(*path))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := &PageData{Images: *images, Path: "photo/"}

		err := outputHTML(templateBox, "index.html", data, w)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Printf("Starting listening on %s... \n", *address)
	log.Fatal(http.ListenAndServe(*address, nil))
}

func outputFiles(publicBox, templateBox embed.FS, path *string, images *[]string) {
	data := &PageData{Images: *images}

	targetPath := filepath.Join(*path, "index.html")

	out, err := os.Create(targetPath)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	if err := outputHTML(templateBox, "index.html", data, out); err != nil {
		log.Fatal(err)
	}

	if err := copyStaticFiles(publicBox, path); err != nil {
		log.Fatal(err)
	}
}

func copyStaticFiles(publicBox embed.FS, targetPath *string) error {
	staticTarget := filepath.Join(*targetPath, "public")
	if _, err := os.Stat(staticTarget); err != nil {
		if err := os.Mkdir(staticTarget, os.ModePerm); err != nil {
			log.Fatal("aici", err)
		}
	}

	return fs.WalkDir(publicBox, ".", func(fileName string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if fileName == "." {
			return nil
		}

		if d.IsDir() {
			if _, err := os.Stat(fileName); err != nil {
				if err := os.Mkdir(fileName, os.ModePerm); err != nil {
					return err
				}
			}
		} else {
			out, err := os.Create(fileName)
			if err != nil {
				return err
			}
			defer out.Close()

			buf, err := fs.ReadFile(publicBox, fileName)
			if err != nil {
				return err
			}

			if _, err := out.Write(buf); err != nil {
				return err
			}

		}

		return nil
	})
}

func outputHTML(templateBox embed.FS, file string, pageData *PageData, wr io.Writer) error {
	t, err := template.ParseFS(templateBox, "template/index.html")

	if err != nil {
		log.Fatal(err)
	}

	return t.Execute(wr, *pageData)
}

func createThumbs(path string, name string) {
	var img image.Image

	targetPath := filepath.Join(path, ".thumb")

	thumb200Path := filepath.Join(targetPath, "/"+name+".200.thumb")
	if _, err := os.Stat(thumb200Path); err != nil {
		if img == nil {
			img = openImage(filepath.Join(path, name))
		}

		createThumb(img, 200, thumb200Path)
	}

	thumb800Path := filepath.Join(targetPath, "/"+name+".800.thumb")
	if _, err := os.Stat(thumb800Path); err != nil {
		if img == nil {
			img = openImage(filepath.Join(path, name))
		}

		createThumb(img, 800, thumb800Path)
	}
}

func openImage(path string) image.Image {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	var img image.Image

	ext := filepath.Ext(path)

	if ext == ".jpg" {
		img, err = jpeg.Decode(file)
	} else if ext == ".png" {
		img, err = png.Decode(file)
	}

	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	return img
}

func createThumb(img image.Image, size uint, target string) {
	m := resize.Resize(0, size, img, resize.Lanczos3)

	out, err := os.Create(target)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	jpeg.Encode(out, m, nil)
}
