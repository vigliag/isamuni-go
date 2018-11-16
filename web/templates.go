package web

import (
	"html/template"
	"io"
	"log"
	"strings"

	"github.com/labstack/echo-contrib/session"

	"github.com/GeertJohan/go.rice"
	"github.com/microcosm-cc/bluemonday"
	blackfriday "gopkg.in/russross/blackfriday.v2"

	"github.com/labstack/echo"
)

//H is the context for a template
type H map[string]interface{}

//Template is our template type, that exposes a custom Render method
type Template struct {
	templates *template.Template
}

//Render renders a template given its name, and the data to pass to the template engine
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	s, err := session.Get("session", c)

	// add common variables from context
	if viewContext, isMap := data.(H); isMap {
		viewContext["currentUser"] = c.Get("currentUser")
		viewContext["path"] = c.Request().URL.Path
		viewContext["csrf"] = c.Get("csrf")
		if err == nil {
			viewContext["flashes"] = s.Flashes()
		}
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

func loadTemplateFromBox(box *rice.Box, parent *template.Template, name string) {
	templateString, err := box.String(name)
	if err != nil {
		log.Fatal(err)
	}
	// parse and execute the template
	_, err = parent.New(name).Parse(templateString)
	if err != nil {
		log.Fatal(err)
	}
}

func loadTemplates() *Template {
	templateBox, err := rice.FindBox("templates")
	if err != nil {
		log.Fatal(err)
	}

	t := template.New("main")
	t.Funcs(template.FuncMap{
		"sanitize": func(text string) template.HTML {
			return template.HTML(bluemonday.UGCPolicy().Sanitize(text))
		},
		"pageurl": PageURL,
		"caturl":  CatUrl,
		"catname": CatName,
	})
	loadTemplateFromBox(templateBox, t, "__footer.html")
	loadTemplateFromBox(templateBox, t, "__header.html")
	loadTemplateFromBox(templateBox, t, "home.html")
	loadTemplateFromBox(templateBox, t, "login.html")
	loadTemplateFromBox(templateBox, t, "pageEdit.html")
	loadTemplateFromBox(templateBox, t, "pageIndex.html")
	loadTemplateFromBox(templateBox, t, "pageShow.html")
	loadTemplateFromBox(templateBox, t, "pageSearch.html")
	loadTemplateFromBox(templateBox, t, "privacy.html")
	loadTemplateFromBox(templateBox, t, "error.html")

	return &Template{templates: t}
}

//RenderMarkdown renders markdown to safe HTML for use in a template
func RenderMarkdown(m string) template.HTML {
	m = strings.Replace(m, "\r\n", "\n", -1)
	unsafe := blackfriday.Run([]byte(m), blackfriday.WithExtensions(blackfriday.Autolink))
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	return template.HTML(html)
}
