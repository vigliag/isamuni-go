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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
	"github.com/vigliag/isamuni-go/index"
	"github.com/vigliag/isamuni-go/model"
)

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Re-create the index, and optionally search for a term",
	Run:   indexRun,
}

var printindexCmd = &cobra.Command{
	Use:   "printindex",
	Short: "Prints the indexed documents",
	Run:   printIndexRun,
}

func printIndexRun(cmd *cobra.Command, args []string) {
	dbname := path.Join(viper.GetString("data"), "database.db")
	db := model.Connect(dbname)
	m := model.Model{db}

	pages, err := m.AllPages()
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range pages {
		pDoc := index.PageToDoc(p)
		pDocJSON, _ := json.Marshal(pDoc)
		fmt.Println(string(pDocJSON))
	}
}

func indexRun(cmd *cobra.Command, args []string) {
	dbname := path.Join(viper.GetString("data"), "database.db")
	m := &model.Model{model.Connect(dbname)}
	defer m.Close()

	//Remove and re-create the index
	fname := path.Join(viper.GetString("data"), "index.bleve")
	os.RemoveAll(fname)
	blvidx, err := index.NewBleve(fname)
	if err != nil {
		panic(err)
	}
	idx := index.New(blvidx, m)
	defer idx.Close()

	pages, err := m.AllPages()
	if err != nil {
		log.Fatal(err)
	}

	// index some data
	count := 0
	for _, p := range pages {
		err := idx.IndexPage(p)
		if err != nil {
			fmt.Println(err)
			continue
		}
		count++
	}

	fmt.Printf("indexed %v documents \n", count)

	if len(args) < 1 {
		return
	}
	querystring := args[0]

	searchResults, err := idx.SearchPagesByQueryString(querystring)
	for _, result := range searchResults {
		fmt.Println(result.Page.Title, result.Page.Type.CatName(), result.Fragments)
	}

}

func init() {
	rootCmd.AddCommand(indexCmd)
	rootCmd.AddCommand(printindexCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// indexCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// indexCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
