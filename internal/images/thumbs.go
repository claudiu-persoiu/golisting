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

	"github.com/claudiu-persoiu/golisting/internal"
	"golang.org/x/image/draw"
)

const smallThumbSize = 200
const bigThumbSize = 800

func GenerateThumbs(dir string, images []string) []string {

	startTime := time.Now()

	targetPath := filepath.Join(dir, ".thumb")
	if err := internal.CreateDirIfNotExists(targetPath); err != nil {
		log.Fatal(err)
	}

	fileNames := make(chan string, len(images))
	errorImagesQueue := make(chan string)
	var errorImagesSlice []string

	var wg sync.WaitGroup

	for range runtime.GOMAXPROCS(0) {
		go func() {
			for fileName := range fileNames {
				if err := createThumbs(dir, targetPath, fileName); err != nil {
					log.Println("Error creating thumbnail for ", fileName, ": ", err)
					errorImagesQueue <- fileName
				}
				wg.Done()
			}
		}()
	}

	for _, f := range images {
		wg.Add(1)
		fileNames <- f
	}

	close(fileNames)

	go func() {
		for img := range errorImagesQueue {
			errorImagesSlice = append(errorImagesSlice, img)
		}
	}()

	wg.Wait()
	close(errorImagesQueue)

	slog.Info("Finished thumbnails in " + time.Now().Sub(startTime).String())

	return errorImagesSlice
}

func createThumbs(path, targetPath string, name string) error {
	var small, big bool

	thumbSmallPath := filepath.Join(targetPath, name+".small.thumb")
	small = !internal.FileExists(thumbSmallPath)

	thumbBigPath := filepath.Join(targetPath, name+".big.thumb")
	big = !internal.FileExists(thumbBigPath)

	if small || big {
		img, err := openImage(filepath.Join(path, name))
		if err != nil {
			return err
		}
		if small {
			if err := createThumb(img, smallThumbSize, thumbSmallPath); err != nil {
				return err
			}
		}
		if big {
			if err := createThumb(img, bigThumbSize, thumbBigPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func openImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)

	if err != nil {
		return nil, err
	}

	return img, nil
}

func createThumb(src image.Image, size int, target string) error {
	out, err := os.Create(target)
	if err != nil {
		return err
	}
	defer out.Close()

	dst := image.NewRGBA(image.Rect(0, 0, (src.Bounds().Max.X*size)/src.Bounds().Max.Y, size))

	draw.BiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	if err := jpeg.Encode(out, dst, nil); err != nil {
		return err
	}
	return nil
}
