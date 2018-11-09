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
	"os"

	"github.com/spf13/cobra"
	"github.com/vigliag/isamuni-go/index"
	"github.com/vigliag/isamuni-go/model"
)

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "A brief description of your command",
	Long:  `A longer description`,
	Run:   indexCommandRun,
}

func indexCommandRun(cmd *cobra.Command, args []string) {

	// open a new index
	fname := "index.bleve"
	os.RemoveAll(fname)
	idx, err := index.New(fname)
	if err != nil {
		panic(err)
	}

	model.Connect("data/database.db")

	pages, err := model.AllPages()
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

	if len(os.Args) < 2 {
		return
	}
	querystring := os.Args[1]

	searchResults, err := idx.SearchPageByQueryString(querystring)

	if err != nil {
		panic(err)
	}
	fmt.Println(searchResults)
}

func init() {
	rootCmd.AddCommand(indexCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// indexCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// indexCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
