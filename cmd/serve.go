package cmd

import (
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/lyyyuna/xiaolongbaoblog/pkg/config"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use: "s",
	Run: runServe,
}

var port int

func init() {
	serveCmd.Flags().StringVarP(&configPath, "config", "c", "_config.yml", "specify the path of the config file")
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "specify the port of the server")
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) {
	c := config.NewConfig(configPath)

	fs := http.FileServer(http.Dir(filepath.Join(".", c.OutputDir)))
	http.Handle("/", fs)

	log.Printf("listening on :%v", strconv.Itoa(port))

	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
