// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"log"
	"net/http"
	"path"

	"github.com/labstack/echo"
	"github.com/spf13/cobra"
	"github.com/vigliag/isamuni-go/db"
	"github.com/vigliag/isamuni-go/web"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: serveRun,
}

func serveRun(cmd *cobra.Command, args []string) {
	dbPath := path.Join(cmd.Flag("dataPath").Value.String(), "database.db")
	db.Connect(dbPath)

	r := web.CreateServer(echo.New())

	//r.Use(middleware.Recover())

	db.RegisterEmail("vigliag", "vigliag@gmail.com", "password", "admin")
	db.RegisterEmail("testuser", "testuser@example.com", "password", "user")

	listenURL := cmd.Flag("listenURL").Value.String()
	fmt.Println("Server started on " + listenURL)

	log.Fatal(http.ListenAndServe(listenURL, r))
	fmt.Println("serve called")
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().String("dataPath", "./data", "path where isamuni will store its data")
	serveCmd.Flags().String("listenURL", ":8080", "url isamuni should bind to, in the format [<ip>]:port")
}