package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
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

func main() {

	path := flag.String("path", "./", "path to images")
	address := flag.String("address", ":5000", "server address")
	output := flag.Bool("output", false, "output location if html is generated and server will not start")
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

	if !*output {
		startSever(publicBox, templateBox, path, &images, address)
	} else {
		outputFiles(publicBox, templateBox, path, &images)
	}
}

func startSever(publicBox, templateBox *packr.Box, path *string, images *[]string, address *string) {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(publicBox)))
	http.Handle("/photo/", http.StripPrefix("/photo/", http.FileServer(http.Dir(*path))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := &PageData{Images: *images, Path: "photo/"}

		err := outputHTML(templateBox, "index.html", data, w)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Printf("Starting listtening on %s... \n", *address)
	log.Fatal(http.ListenAndServe(*address, nil))
}

func outputFiles(publicBox, templateBox *packr.Box, path *string, images *[]string) {
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

func copyStaticFiles(publicBox *packr.Box, targetPath *string) error {

	staticTarget := filepath.Join(*targetPath, "static")
	if _, err := os.Stat(staticTarget); err != nil {
		if err := os.Mkdir(staticTarget, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	return publicBox.Walk(func(fileName string, fileContent packr.File) error {
		dir := path.Dir(fileName)

		targetdir := filepath.Join(staticTarget, dir)
		if _, err := os.Stat(targetdir); err != nil {
			if err := os.Mkdir(targetdir, os.ModePerm); err != nil {
				return err
			}
		}

		target := filepath.Join(staticTarget, fileName)
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		defer out.Close()

		buf := make([]byte, 1000)
		for {
			n, err := fileContent.Read(buf)
			if err != nil && err != io.EOF {
				return err
			}
			if n == 0 {
				break
			}

			if _, err := out.Write(buf[:n]); err != nil {
				return err
			}
		}
		return nil
	})
}

func outputHTML(templateBox *packr.Box, file string, pageData *PageData, wr io.Writer) error {
	b, err := templateBox.FindString(file)
	if err != nil {
		log.Panic(err)
	}

	t, err := template.New("hello").Parse(b)

	if err != nil {
		log.Panic(err)
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
