/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

var port string
var username string
var password string
var dest string
var host string
// receiveCmd represents the receive command
var receiveCmd = &cobra.Command{
	Use:   "receive",
	Short: "Starts a server to receive files",
	Long: `Starts a server to receive files.`,
	Run: func(cmd *cobra.Command, args []string) {
		receive()
	},
}

func uploadHandler(uploadDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			tmpl, err := template.ParseFiles("src/html/upload.html")
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

			// if dest != "" {
			// 	if _, err := os.Stat(dest); os.IsNotExist(err) {
			// 		fmt.Println("Destination folder does not exist")
			// 		return
			// 	}
			// 	if dest[len(dest)-1:] == "/" {
			// 		dest = dest[:len(dest)-1]
			// 	}
			// } else{
			// 	dest = "./uploads"
			// }
			dst, err := os.Create(filepath.Join(uploadDir, handler.Filename))
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
}

// func getLocalIP() (string, error) {
// 	addrs, err := net.InterfaceAddrs()
// 	if err != nil {
// 		return "", err
// 	}

// 	address := ""

// 	for _, addr := range addrs {
// 		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
// 			if ipnet.IP.To4() != nil {
// 				fmt.Println(ipnet.IP.String())
// 				address = ipnet.IP.String()
// 				// return ipnet.IP.String(), nil
// 			}
// 		}
// 	}
// 	if address != "" {
// 		return address, nil
// 	}
// 	return "", fmt.Errorf("cannot find local IP address")
// }

func receive() {
	if host == "" {
		host = "localhost"
	}
	// Get the current user's home directory
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error getting current user:", err)
		return
	}
	uploadDir := filepath.Join(usr.HomeDir, "uploads")

	// Create the uploads directory if it doesn't exist
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err = os.Mkdir(uploadDir, 0755)
		if err != nil {
			fmt.Println("Error creating uploads directory:", err)
			return
		}
	}

	// Print the destination folder
	fmt.Printf("Destination folder: %s\n", uploadDir)

    // Serve static files
    fs := http.FileServer(http.Dir("src/static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

	if username != "" && password != "" {
		http.Handle("/", basicAuth(uploadHandler(uploadDir)))	
	} else {
		http.HandleFunc("/", uploadHandler(uploadDir))
	}
	if username == "" && password == "" {
		fmt.Println("Authentication disabled (password and username not provided)")
	} else {
		fmt.Println("Authentication enabled with username: " + username + " and password: " + password)
	}

	fmt.Println("\nServer started on: http://" + host + ":" + port)
	http.ListenAndServe(host + ":" + port, nil)
}

func basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != username || pass != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func init() {
	rootCmd.AddCommand(receiveCmd)
	receiveCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port number for the server")
	receiveCmd.Flags().StringVarP(&username, "username", "u", "", "Username for basic auth (to be entered by the sender in browser)")
	receiveCmd.Flags().StringVarP(&password, "password", "P", "", "Password for basic auth (to be entered by the sender in browser)")
	receiveCmd.MarkFlagsRequiredTogether("username", "password")
	receiveCmd.Flags().StringVarP(&dest, "dest", "d", "", "Destination folder (should exist) to save the files")
	receiveCmd.Flags().StringVarP(&host, "host", "H", "", "Host address or Local IP to bind the server to (default is localhost)")
}