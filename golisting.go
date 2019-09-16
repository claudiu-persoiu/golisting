package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"text/template"

	"github.com/gobuffalo/packr/v2"

	"github.com/nfnt/resize"
)

type PageData struct {
	Images []string
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

func main() {

	path := flag.String("path", "./", "path to images")
	address := flag.String("address", ":5000", "server address")
	flag.Parse()

	publicBox := packr.New("public", "./public")
	templateBox := packr.New("template", "./template")

	dir := filepath.Join(".", *path)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	targetPath := filepath.Join(dir, ".thumb")
	if _, err := os.Stat(targetPath); err != nil {
		err := os.Mkdir(targetPath, os.ModePerm)

		if err != nil {
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

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(publicBox)))
	http.Handle("/photo/", http.StripPrefix("/photo/", http.FileServer(http.Dir(*path))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		b, err := templateBox.FindString("index.html")
		if err != nil {
			log.Panic(err)
		}

		t, err := template.New("hello").Parse(b)

		if err != nil {
			log.Panic(err)
		}

		data := &PageData{Images: images}

		err = t.Execute(w, *data)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Printf("Starting listtening on %s... \n", *address)
	log.Fatal(http.ListenAndServe(*address, nil))
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
