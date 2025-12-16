package server

import (
	"embed"
	"log"
	"net/http"

	"github.com/claudiu-persoiu/golisting/internal"
)

func StartSever(publicBox, templateBox embed.FS, path *string, images []string, address string) {
	http.Handle("/public/", http.FileServer(http.FS(publicBox)))
	http.Handle("/photo/", http.StripPrefix("/photo/", http.FileServer(http.Dir(*path))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := &internal.PageData{Images: images, Path: "photo/"}

		err := internal.OutputHTML(templateBox, "index.html", data, w)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Printf("Starting listening on %s... \n", address)
	log.Fatal(http.ListenAndServe(address, nil))
}
