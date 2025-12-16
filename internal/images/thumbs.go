package images

import (
	"image"
	"image/jpeg"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"golang.org/x/image/draw"
)

const smallThumbSize = 200
const bigThumbSize = 800

func GenerateThumbs(dir string, images []string) {

	startTime := time.Now()

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

	slog.Info("Finished thumbnails in " + time.Now().Sub(startTime).String())
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
