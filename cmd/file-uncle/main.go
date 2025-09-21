package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/theadarshsaxena/file-uncle/internal/config"
	"github.com/theadarshsaxena/file-uncle/internal/server"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
    defer logger.Sync()
	
	rootCmd := &cobra.Command{
        Use:   "file-uncle",
        Short: "File Uncle - Fast, secure file transfer",
    }
	rootCmd.PersistentFlags().StringVarP(&config.Shared.Host, "host", "H", "localhost", "Host address or Local IP to bind the server to (default is localhost)")
	rootCmd.PersistentFlags().StringVarP(&config.Shared.Port, "port", "p", "8080", "Port number for the server")
	rootCmd.PersistentFlags().BoolVarP(&config.Shared.WithNgrok, "with-ngrok", "n", false, "Start an ngrok tunnel for public access")

    // Serve command
    serveCmd := &cobra.Command{
        Use:   "serve",
        Short: "Serve files for download",
        Run: func(cmd *cobra.Command, args []string) {
            if err := server.RunServe(logger); err != nil {
                logger.Fatal("serve failed", zap.Error(err))
            }
        },
    }
    serveCmd.Flags().StringVarP(&config.Shared.Directory, "directory", "d", "./", "Directory to serve")

    // Receive command
    receiveCmd := &cobra.Command{
        Use:   "receive",
        Short: "Receive files via upload",
        Run: func(cmd *cobra.Command, args []string) {
            if err := server.RunReceive(logger); err != nil {
                logger.Fatal("receive failed", zap.Error(err))
            }
        },
    }
	receiveCmd.Flags().StringVarP(&config.Shared.Username, "username", "u", "", "Username for basic auth (to be entered by the sender in browser)")
	receiveCmd.Flags().StringVarP(&config.Shared.Password, "password", "P", "", "Password for basic auth (to be entered by the sender in browser)")
	receiveCmd.MarkFlagsRequiredTogether("username", "password")
	receiveCmd.Flags().StringVarP(&config.Shared.Destination, "dest", "d", "", "Destination folder (should exist) to save the files")
	
    rootCmd.AddCommand(serveCmd, receiveCmd)

    if err := rootCmd.Execute(); err != nil {
        logger.Fatal("command failed", zap.Error(err))
        os.Exit(1)
    }
}
