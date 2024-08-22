/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		serveFile()
	},
}

type FileInfo struct {
	LineNumber int
	Name string
	DisplayName string
	FileSize string
}

func truncateFileName(name string, length int) string {
	if utf8.RuneCountInString(name) > length {
		return name[:length] + "..."
	}
	return name
}

func FileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
}

func listFiles(w http.ResponseWriter, r *http.Request) {
	folderPath := "./" // folder to serve files from

	var files []FileInfo

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			displayName := truncateFileName(info.Name(), 30)
			files = append(files, FileInfo{
				Name: info.Name(),
				DisplayName: displayName,
				FileSize: FileSize(info.Size()),
				LineNumber: len(files) + 1,
			})
		}
		return nil
	})
	if err != nil {
		http.Error(w, "Unable to list files", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("serve").Parse(ServeHTML)
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
  
	tmpl.Execute(w, files)
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")
	filePath := fmt.Sprintf("./files/%s", fileName)

	http.ServeFile(w, r, filePath)
}

func serveFile() {
	http.HandleFunc("/", listFiles)
	http.HandleFunc("/download", downloadFile)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("cmd/src/static"))))

	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe("localhost:8080", nil)
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
