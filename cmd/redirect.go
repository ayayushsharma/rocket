package cmd

import (
	"ayayushsharma/rocket/constants"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var redirectCmd = &cobra.Command{
	Use:   "redirect",
	Short: "Redirects application to a more appropriate port",
	Long: "Main rocket application runs on a port available in user space.\n" +
		"To use a different port to access the applications, you listen on \n" +
		"on a more accessible port and redirect the traffic rocket's main application",

	Run: func(cmd *cobra.Command, args []string) {
		applicationURL := fmt.Sprintf("http://localhost:%d", constants.ApplicationPort)
		proxyTarget, err := url.Parse(applicationURL)
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(proxyTarget)

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		})

		portNumber := viper.GetInt("port")
		port := fmt.Sprintf(":%d", portNumber)

		log.Printf("Starting reverse proxy on %s, forwarding to %s", port, proxyTarget)
		err = http.ListenAndServe(port, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}

	},
}

func init() {
	rootCmd.AddCommand(redirectCmd)

	redirectCmd.Flags().Int("port", 8080, "Port to run the server on")
}
