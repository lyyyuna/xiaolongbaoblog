package cmd

import (
	"path/filepath"

	"github.com/lyyyuna/xiaolongbaoblog/pkg/config"
	"github.com/lyyyuna/xiaolongbaoblog/pkg/deploy"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use: "d",
	Run: runDeploy,
}

func init() {
	deployCmd.Flags().StringVarP(&configPath, "config", "c", "_config.yml", "specify the path of the config file")
	rootCmd.AddCommand(deployCmd)
}

func runDeploy(cmd *cobra.Command, args []string) {
	c := config.NewConfig(configPath)
	d := deploy.NewDeploy(c)
	d.DeployToServer(filepath.Join(".", c.OutputDir))
}
