package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl, err := template.ParseFiles("upload.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	} else {
		r.ParseMultipartForm(10 << 20) // limit your max input length!
		file, handler, err := r.FormFile("uploadFile")
		if err != nil {
			fmt.Println("Error retrieving the file")
			fmt.Println(err)
			return
		}
		defer file.Close()

		fmt.Printf("Uploaded File: %s\n", handler.Filename)
		fmt.Printf("File Size: %d\n", handler.Size)
		fmt.Printf("MIME Header: %v\n", handler.Header)

		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error getting user home directory")
			fmt.Println(err)
			return
		}

		uploadDir := userHomeDir + "/uploads/"
		dst, err := os.Create(uploadDir + handler.Filename)
		if err != nil {
			fmt.Println("Error creating file")
			fmt.Println(err)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			fmt.Println("Error copying file")
			fmt.Println(err)
			return
		}
		fmt.Fprintf(w, "Successfully Uploaded File\n")
	}
}

func main() {
	http.HandleFunc("/upload", uploadHandler)

	fmt.Println("Server started on: http://localhost:8080")
	http.ListenAndServe(":8085", nil)
}
