package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/blevesearch/bleve/mapping"

	"github.com/blevesearch/bleve"
	_ "github.com/blevesearch/bleve/analysis/analyzer/keyword"
	_ "github.com/blevesearch/bleve/analysis/analyzer/simple"
	_ "github.com/blevesearch/bleve/analysis/lang/it"
	_ "github.com/blevesearch/bleve/search/highlight/format/ansi"
	_ "github.com/blevesearch/bleve/search/highlight/format/html"
	"github.com/vigliag/isamuni-go/db"
)

type doc struct {
	Content string `json:"content"`
	Type    string `json:"type"`
	City    string `json:"city"`
	Sector  string `json:"sector"`
}

func (doc) BleveType() string {
	return "page"
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

func simpleTextFieldMapping() *mapping.FieldMapping {
	fieldMapping := bleve.NewTextFieldMapping()
	fieldMapping.Analyzer = "simple"
	return fieldMapping
}

func keywordTextFieldMapping() *mapping.FieldMapping {
	fieldMapping := bleve.NewTextFieldMapping()
	fieldMapping.Analyzer = "keyword"
	return fieldMapping
}

func main() {

	// open a new index
	fname := "index.bleve"

	// create mapping, defaulting to treating contents as text in the italian language
	mapping := bleve.NewIndexMapping()
	mapping.DefaultAnalyzer = "it"

	// create a mapping for pages
	pageMapping := bleve.NewDocumentMapping()

	pageMapping.AddFieldMappingsAt("name", simpleTextFieldMapping())
	pageMapping.AddFieldMappingsAt("type", keywordTextFieldMapping())
	pageMapping.AddFieldMappingsAt("city", keywordTextFieldMapping())
	pageMapping.AddFieldMappingsAt("sector", keywordTextFieldMapping())

	contentFieldMapping := bleve.NewTextFieldMapping()
	contentFieldMapping.Analyzer = "it"
	pageMapping.AddFieldMappingsAt("content", contentFieldMapping)

	mapping.AddDocumentMapping("page", pageMapping)

	os.RemoveAll(fname)
	index, err := bleve.New(fname, mapping)
	if err != nil {
		panic(err)
	}

	db.Connect()

	var pages []db.Page
	db.Db.Find(&pages)

	// index some data
	for _, p := range pages {
		d := pageToDoc(&p)
		err = index.Index(fmt.Sprintf("%d", p.ID), d)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	numIndexed, _ := index.DocCount()
	fmt.Printf("indexed %v documents \n", numIndexed)

	if len(os.Args) < 2 {
		return
	}

	// search for some text
	querystring := os.Args[1]
	query := bleve.NewQueryStringQuery(querystring)
	search := bleve.NewSearchRequest(query)

	search.AddFacet("sector", bleve.NewFacetRequest("sector", 10))
	search.AddFacet("city", bleve.NewFacetRequest("city", 10))
	search.AddFacet("type", bleve.NewFacetRequest("type", 10))

	search.Highlight = bleve.NewHighlight()
	search.Highlight.AddField("content")

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
