package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	UPLOAD_DIR = "./uploads"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hello")
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	imageId := r.FormValue("id")
	imagePath := UPLOAD_DIR + "/" + imageId
	w.Header().Set("Content-Type", "image")
	http.ServeFile(w, r, imagePath)
}
func isExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return os.IsExist(err)
}
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		io.WriteString(w, "<form method=\"POST\" action=\"/upload\" "+
			" enctype=\"multipart/form-data\">"+
			"Choose an image to upload: <input name=\"image\" type=\"file\" />"+
			"<input type=\"submit\" value=\"Upload\" />"+
			"</form>")
		return
	}

	if r.Method == "POST" {
		f, h, err := r.FormFile("image")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		filename := h.Filename
		defer f.Close()

		t, err := os.Create(UPLOAD_DIR + "/" + filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer t.Close()

		http.Redirect(w, r, "/view?id="+filename, http.StatusFound)
	}
}

func main() {
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/view", viewHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe", err.Error())
	}
}
