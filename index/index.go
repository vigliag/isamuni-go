package index

import (
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/blevesearch/bleve"
	_ "github.com/blevesearch/bleve/analysis/analyzer/keyword"
	_ "github.com/blevesearch/bleve/analysis/analyzer/simple"
	_ "github.com/blevesearch/bleve/analysis/lang/it"
	"github.com/blevesearch/bleve/mapping"
	_ "github.com/blevesearch/bleve/search/highlight/format/ansi"
	_ "github.com/blevesearch/bleve/search/highlight/format/html"
	"github.com/vigliag/isamuni-go/model"
)

type Doc map[string]string

type SearchResult struct {
	Fragments map[string][]string
	Page      *model.Page
}

type ByKindSearchResult struct {
	ProfessionalsResults []SearchResult
	CommunitiesResults   []SearchResult
	CompaniesResults     []SearchResult
}

func (Doc) BleveType() string {
	return "page"
}

var headersRegex = regexp.MustCompile(`(?m)^#+.+$`)

func PageToDoc(p *model.Page) Doc {
	d := model.ParseContent(p.Content)
	d["name"] = p.Title
	return d
}

var indexMapping *mapping.IndexMappingImpl

type Index struct {
	fname string
	idx   bleve.Index
}

var Idx *Index

func DefaultIndex() *Index {
	var err error

	if Idx == nil {
		Idx, err = Open("data/index.bleve")
	}

	if err != nil {
		panic(err)
	}

	return Idx
}

// New initializes an index with a given filename
func New(fname string) (*Index, error) {
	idx, err := bleve.New(fname, indexMapping)
	if err != nil {
		return nil, err
	}
	return &Index{fname, idx}, nil
}

func Open(fname string) (*Index, error) {
	idx, err := bleve.Open(fname)
	if err != nil {
		return nil, err
	}
	return &Index{fname, idx}, nil
}

// IndexPage puts a page in the index
func (i Index) IndexPage(page *model.Page) error {
	d := PageToDoc(page)
	return i.idx.Index(fmt.Sprintf("%d", page.ID), d)
}

func (i Index) searchPageByQueryString(querystring string) (*bleve.SearchResult, error) {
	query := bleve.NewQueryStringQuery(querystring)
	search := bleve.NewSearchRequest(query)

	search.AddFacet("sector", bleve.NewFacetRequest("sector", 10))
	search.AddFacet("city", bleve.NewFacetRequest("city", 10))
	search.AddFacet("type", bleve.NewFacetRequest("type", 10))

	search.Highlight = bleve.NewHighlight()
	search.Highlight.AddField("short")
	search.Highlight.AddField("skills")
	search.Highlight.AddField("tags")
	search.Highlight.AddField("description")

	searchResults, err := i.idx.Search(search)
	if err != nil {
		return nil, err
	}
	return searchResults, nil
}

// SearchPageByQueryString search a page by a bleve query string
func (i Index) SearchPagesByQueryString(queryString string) ([]SearchResult, error) {
	// Create a map pageid -> page
	pages, err := model.AllPages()
	if err != nil {
		return nil, err
	}
	pagemap := make(map[uint]*model.Page)
	for _, p := range pages {
		pagemap[p.ID] = p
	}

	matches, err := i.searchPageByQueryString(queryString)
	if err != nil {
		return nil, err
	}

	searchResuls := make([]SearchResult, len(matches.Hits))

	for j, hit := range matches.Hits {
		id, _ := strconv.Atoi(hit.ID)
		p, _ := pagemap[uint(id)]
		if p == nil {
			log.Println("Error, invalid ID in index")
			continue
		}
		searchResuls[j] = SearchResult{
			Fragments: hit.Fragments,
			Page:      p,
		}
	}

	return searchResuls, nil
}

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
	pageMapping.AddFieldMappingsAt("tags", simpleTextFieldMapping())
	pageMapping.AddFieldMappingsAt("skills", simpleTextFieldMapping())

	indexMapping.AddDocumentMapping("page", pageMapping)
}
