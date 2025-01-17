package cmd

import (
	//"crypto/tls"
	//"net/http"
	"os/user"
	"os"

	"github.com/helviojunior/sprayshark/internal/ascii"
	"github.com/helviojunior/sprayshark/pkg/log"
	"github.com/helviojunior/sprayshark/pkg/runner"
	"github.com/spf13/cobra"
)

var (
	opts = &runner.Options{}
)

var rootCmd = &cobra.Command{
	Use:   "sprayshark",
	Short: "SprayShark is a modular password sprayer",
	Long:  ascii.Logo(),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		
		usr, err := user.Current()
	    if err != nil {
	       return err
	    }

	    opts.Writer.UserPath = usr.HomeDir

		if opts.Logging.Silence {
			log.EnableSilence()
		}

		if opts.Logging.Debug && !opts.Logging.Silence {
			log.EnableDebug()
			log.Debug("debug logging enabled")
		}
		return nil
	},
}

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SilenceErrors = true
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Disable Certificate Validation (Globally)
	//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	rootCmd.PersistentFlags().BoolVarP(&opts.Logging.Debug, "debug-log", "D", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVarP(&opts.Logging.Silence, "quiet", "q", false, "Silence (almost all) logging")
	rootCmd.PersistentFlags().BoolVarP(&opts.Chrome.SkipSSLCheck, "ssl-insecure", "K", true, "SSL Insecure")
	rootCmd.PersistentFlags().StringVarP(&opts.Chrome.Proxy, "proxy", "X", "", "Proxy to pass traffic through: <scheme://ip:port>")
	rootCmd.PersistentFlags().StringVarP(&opts.Chrome.ProxyUser, "proxy-user", "", "", "Proxy User")
	rootCmd.PersistentFlags().StringVarP(&opts.Chrome.ProxyPassword, "proxy-pass", "", "", "Proxy Password")

	
}
