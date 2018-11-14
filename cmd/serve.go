package cmd

import (
	"fmt"
	"log"
	"net/http"
	"path"

	"github.com/vigliag/isamuni-go/index"

	"github.com/spf13/viper"

	"github.com/labstack/echo"
	"github.com/spf13/cobra"
	"github.com/vigliag/isamuni-go/model"
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
	dataPath := viper.GetString("data")
	dbPath := path.Join(dataPath, "database.db")

	fmt.Println("Using data folder", dataPath)
	fmt.Println("Using app url", viper.GetString("APP_URL"))

	model.Connect(dbPath)
	index.DefaultIndex() //initialize default index

	r := web.CreateServer(echo.New())

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
