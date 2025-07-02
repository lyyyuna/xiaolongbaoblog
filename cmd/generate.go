package cmd

import (
	"github.com/lyyyuna/xiaolongbaoblog/pkg/config"
	"github.com/lyyyuna/xiaolongbaoblog/pkg/site"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use: "g",
	Run: runGenerate,
}

var (
	configPath string
	indexnow   bool
)

func init() {
	generateCmd.Flags().StringVarP(&configPath, "config", "c", "_config.yml", "specify the path of the config file")
	generateCmd.Flags().BoolVarP(&indexnow, "index", "i", false, "submit indexnow")
	rootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, args []string) {
	c := config.NewConfig(configPath)
	s := site.NewGenerate(c)
	s.Output()

	if indexnow {
		s.SubmitIndexNow()
	}
}
