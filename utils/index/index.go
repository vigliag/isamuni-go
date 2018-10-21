package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/blevesearch/bleve"
	_ "github.com/blevesearch/bleve/analysis/lang/it"
	_ "github.com/blevesearch/bleve/search/highlight/format/ansi"
	_ "github.com/blevesearch/bleve/search/highlight/format/html"
	"github.com/vigliag/isamuni-go/db"
)

type doc struct {
	Content string
	Type    string
	City    string
	Sector  string
}

var headersRegex = regexp.MustCompile(`(?m)^#+.+$`)

func pageToDoc(p *db.Page) doc {
	var d doc
	d.Content = headersRegex.ReplaceAllString(p.Content, "")
	d.Type = p.Type.CatName()
	d.Sector = p.Sector
	d.City = p.City
	return d
}

func main() {

	db.Connect()

	var pages []db.Page
	db.Db.Find(&pages)

	// open a new index
	fname := "index.bleve"
	mapping := bleve.NewIndexMapping()
	mapping.DefaultAnalyzer = "it"
	os.RemoveAll(fname)
	index, err := bleve.New(fname, mapping)
	if err != nil {
		panic(err)
	}

	// index some data
	for _, p := range pages {
		d := pageToDoc(&p)
		err = index.Index(fmt.Sprintf("%d", p.ID), d)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	// search for some text

	query := bleve.NewMatchQuery("javascript")
	search := bleve.NewSearchRequest(query)

	search.AddFacet("sector", bleve.NewFacetRequest("Sector", 10))
	search.AddFacet("city", bleve.NewFacetRequest("City", 10))
	search.AddFacet("type", bleve.NewFacetRequest("Type", 10))

	search.Highlight = bleve.NewHighlight()
	//search.Highlight.AddField("Content")
	//search.Fields = []string{"Content"}

	searchResults, err := index.Search(search)

	if err != nil {
		panic(err)
	}
	fmt.Println(searchResults)
	// for _, h := range searchResults.Hits {
	// 	fmt.Println("Match: ")
	// 	fmt.Println(h.)
	// }
}
