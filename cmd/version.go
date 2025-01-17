package cmd

import (
    "fmt"

    "github.com/helviojunior/sprayshark/internal/ascii"
    "github.com/helviojunior/sprayshark/internal/version"
    "github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Get the sprayshark version",
    Long:  ascii.LogoHelp(`Get the sprayshark version.`),
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println(ascii.Logo())
        fmt.Printf("\nsprayshark: %s\ngit hash: %s\nbuild env: %s\nbuild time: %s\n",
            version.Version, version.GitHash, version.GoBuildEnv, version.GoBuildTime)
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
}