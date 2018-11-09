package index

import (
	"fmt"
	"regexp"

	"github.com/blevesearch/bleve"
	_ "github.com/blevesearch/bleve/analysis/analyzer/keyword"
	_ "github.com/blevesearch/bleve/analysis/analyzer/simple"
	_ "github.com/blevesearch/bleve/analysis/lang/it"
	"github.com/blevesearch/bleve/mapping"
	_ "github.com/blevesearch/bleve/search/highlight/format/ansi"
	_ "github.com/blevesearch/bleve/search/highlight/format/html"
	"github.com/vigliag/isamuni-go/model"
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

func pageToDoc(p *model.Page) doc {
	var d doc
	d.Content = headersRegex.ReplaceAllString(p.Content, "")
	d.Type = p.Type.CatName()
	d.Sector = p.Sector
	d.City = p.City
	return d
}

var indexMapping *mapping.IndexMappingImpl

func init() {

	simpleTextFieldMapping := func() *mapping.FieldMapping {
		fieldMapping := bleve.NewTextFieldMapping()
		fieldMapping.Analyzer = "simple"
		return fieldMapping
	}

	keywordTextFieldMapping := func() *mapping.FieldMapping {
		fieldMapping := bleve.NewTextFieldMapping()
		fieldMapping.Analyzer = "keyword"
		return fieldMapping
	}

	indexMapping = bleve.NewIndexMapping()
	indexMapping.DefaultAnalyzer = "it"

	// create a mapping for pages
	pageMapping := bleve.NewDocumentMapping()

	pageMapping.AddFieldMappingsAt("name", simpleTextFieldMapping())
	pageMapping.AddFieldMappingsAt("type", keywordTextFieldMapping())
	pageMapping.AddFieldMappingsAt("city", keywordTextFieldMapping())
	pageMapping.AddFieldMappingsAt("sector", keywordTextFieldMapping())

	contentFieldMapping := bleve.NewTextFieldMapping()
	contentFieldMapping.Analyzer = "it"
	pageMapping.AddFieldMappingsAt("content", contentFieldMapping)

	indexMapping.AddDocumentMapping("page", pageMapping)
}

type Index struct {
	fname string
	idx   bleve.Index
}

// New initializes an index with a given filename
func New(fname string) (*Index, error) {
	idx, err := bleve.New(fname, indexMapping)
	if err != nil {
		return nil, err
	}
	return &Index{fname, idx}, nil
}

// IndexPage puts a page in the index
func (i Index) IndexPage(page model.Page) error {
	d := pageToDoc(&page)
	return i.idx.Index(fmt.Sprintf("%d", page.ID), d)
}

// SearchPageByQueryString search a page by a bleve query string
func (i Index) SearchPageByQueryString(querystring string) (*bleve.SearchResult, error) {
	query := bleve.NewQueryStringQuery(querystring)
	search := bleve.NewSearchRequest(query)

	search.AddFacet("sector", bleve.NewFacetRequest("sector", 10))
	search.AddFacet("city", bleve.NewFacetRequest("city", 10))
	search.AddFacet("type", bleve.NewFacetRequest("type", 10))

	search.Highlight = bleve.NewHighlight()
	search.Highlight.AddField("content")

	searchResults, err := i.idx.Search(search)
	if err != nil {
		return nil, err
	}
	return searchResults, nil
}
