package main

import (
	"embed"
	"flag"
	"image"
	"image/jpeg"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"text/template"
	"time"

	"golang.org/x/image/draw"
)

const smallThumbSize = 200
const bigThumbSize = 800

type PageData struct {
	Images []string
	Path   string
}

type byName []os.DirEntry

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

	startTime := time.Now()

	dir := filepath.Join(".", *path)
	images := getImages(dir)

	generateThumbs(dir, images)

	log.Printf("Finished thumbnails in %v\n", time.Now().Sub(startTime))

	if !*output {
		startSever(publicBox, templateBox, path, &images, address)
	} else {
		outputFiles(publicBox, templateBox, path, &images)
	}
}

func getImages(dir string) []string {

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	sort.Sort(byName(files))

	var images []string
	for _, f := range files {
		ext := filepath.Ext(f.Name())
		if ext == ".jpg" || ext == ".png" {
			images = append(images, f.Name())
		}
	}
	return images
}

func generateThumbs(dir string, images []string) {

	targetPath := filepath.Join(dir, ".thumb")
	if _, err := os.Stat(targetPath); err != nil {
		if err := os.Mkdir(targetPath, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	fileNames := make(chan string, len(images))

	var wg sync.WaitGroup

	for range runtime.GOMAXPROCS(0) {
		go func() {
			for j := range fileNames {
				createThumbs(dir, targetPath, j)
				wg.Done()
			}
		}()
	}

	for _, f := range images {
		wg.Add(1)
		fileNames <- f
	}

	close(fileNames)

	wg.Wait()
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
			if _, err := os.Stat(target); err != nil {
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

func outputHTML(templateBox embed.FS, file string, pageData *PageData, wr io.Writer) error {
	t, err := template.ParseFS(templateBox, "template/index.html")

	if err != nil {
		log.Fatal(err)
	}

	return t.Execute(wr, *pageData)
}

func createThumbs(path, targetPath string, name string) {
	var small, big bool

	thumbSmallPath := filepath.Join(targetPath, name+".small.thumb")
	if inf, _ := os.Stat(thumbSmallPath); inf == nil {
		small = true
	}

	thumbBigPath := filepath.Join(targetPath, name+".big.thumb")
	if inf, _ := os.Stat(thumbBigPath); inf == nil {
		big = true
	}

	if small || big {
		img := openImage(filepath.Join(path, name))
		if small {
			createThumb(img, smallThumbSize, thumbSmallPath)
		}
		if big {
			createThumb(img, bigThumbSize, thumbBigPath)
		}
	}
}

func openImage(path string) image.Image {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)

	if err != nil {
		log.Fatal(err)
	}

	return img
}

func createThumb(src image.Image, size int, target string) {
	out, err := os.Create(target)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	dst := image.NewRGBA(image.Rect(0, 0, (src.Bounds().Max.X*size)/src.Bounds().Max.Y, size))

	draw.BiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	jpeg.Encode(out, dst, nil)
}
