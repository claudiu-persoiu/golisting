package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/claudiu-persoiu/golisting/internal/images"
	"github.com/claudiu-persoiu/golisting/internal/server"
	"github.com/claudiu-persoiu/golisting/internal/static"
)

//go:embed public
var publicBox embed.FS

//go:embed template
var templateBox embed.FS

func main() {

	path := flag.String("path", "", "path to images")
	help := flag.Bool("help", false, "help menu")
	address := flag.String("address", ":5000", "server address")
	output := flag.Bool("output", false, "output location if html is generated and server will not start")
	flag.Parse()

	if *path == "" || *help {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	dir := filepath.Join(".", *path)
	imgs := images.GetImages(dir)

	noThumbs := images.GenerateThumbs(dir, imgs)

	if len(noThumbs) > 0 {
		imgs = images.FilterOutImages(imgs, noThumbs)
	}



	if !*output {
		server.StartSever(publicBox, templateBox, path, imgs, *address)
	} else {
		static.OutputFiles(publicBox, templateBox, path, imgs)
	}
}
