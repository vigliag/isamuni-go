package web

import (
	"html/template"
	"io"
	"log"
	"strings"
	"time"

	"github.com/labstack/echo-contrib/session"

	rice "github.com/GeertJohan/go.rice"
	"github.com/microcosm-cc/bluemonday"
	blackfriday "github.com/russross/blackfriday/v2"

	"github.com/labstack/echo/v4"
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
			s.Save(c.Request(), c.Response())
		}
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

func (t *Template) RenderString(name string, data interface{}) (string, error) {
	w := strings.Builder{}
	err := t.templates.ExecuteTemplate(&w, name, data)
	if err != nil {
		return "", err
	}
	return w.String(), nil
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

func LoadTemplates() *Template {
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
		"datetime": func(t time.Time) string {
			return t.Format("2 Jan 2006 15:04")
		},
	})

	loadTemplateFromBox(templateBox, t, "__footer.html")
	loadTemplateFromBox(templateBox, t, "__header.html")
	loadTemplateFromBox(templateBox, t, "home.html")
	loadTemplateFromBox(templateBox, t, "login.html")
	loadTemplateFromBox(templateBox, t, "pageEdit.html")
	loadTemplateFromBox(templateBox, t, "pageIndex.html")
	loadTemplateFromBox(templateBox, t, "pageShow.html")
	loadTemplateFromBox(templateBox, t, "pageSearch.html")
	loadTemplateFromBox(templateBox, t, "profileEdit.html")
	loadTemplateFromBox(templateBox, t, "privacy.html")
	loadTemplateFromBox(templateBox, t, "error.html")
	loadTemplateFromBox(templateBox, t, "exampleProfessional.html")
	loadTemplateFromBox(templateBox, t, "exampleCommunity.html")
	loadTemplateFromBox(templateBox, t, "exampleCompany.html")
	loadTemplateFromBox(templateBox, t, "admin.html")

	return &Template{templates: t}
}

//RenderMarkdown renders markdown to safe HTML for use in a template
func RenderMarkdown(m string) template.HTML {
	m = strings.Replace(m, "\r\n", "\n", -1)
	unsafe := blackfriday.Run([]byte(m), blackfriday.WithExtensions(blackfriday.Autolink))
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	return template.HTML(html)
}
