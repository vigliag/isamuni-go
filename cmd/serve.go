package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/viper"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/cobra"
	"github.com/vigliag/isamuni-go/web"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the isamuni server",
	Run:   serveRun,
}

func serveRun(cmd *cobra.Command, args []string) {
	listenURL := viper.GetString("LISTEN_URL")
	ctl := GetController()
	r := web.CreateServer(echo.New(), ctl)

	// attach CSRF middleware here, so that we don't have it during testing
	r.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:csrf",
	}))

	//r.Use(middleware.Recover())

	//model.RegisterEmail("vigliag", "vigliag@gmail.com", "password", "admin")
	//model.RegisterEmail("testuser", "testuser@example.com", "password", "user")

	fmt.Println("Server started on", listenURL)

	log.Fatal(http.ListenAndServe(listenURL, r))
	fmt.Println("serve called")
}

func init() {
	serveCmd.Flags().String("LISTEN_URL", ":8080", "url isamuni should bind to, in the format [<ip>]:port")
	viper.BindPFlag("LISTEN_URL", serveCmd.Flags().Lookup("LISTEN_URL"))

	rootCmd.AddCommand(serveCmd)
}
