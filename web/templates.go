package web

import (
	"html/template"
	"io"
	"log"

	"github.com/GeertJohan/go.rice"

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
	// add common variables from context
	if viewContext, isMap := data.(H); isMap {
		viewContext["currentUser"] = c.Get("currentUser")
		viewContext["path"] = c.Request().URL.Path
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
	loadTemplateFromBox(templateBox, t, "__footer.html")
	loadTemplateFromBox(templateBox, t, "__header.html")
	loadTemplateFromBox(templateBox, t, "home.html")
	loadTemplateFromBox(templateBox, t, "login.html")
	loadTemplateFromBox(templateBox, t, "pageEdit.html")
	loadTemplateFromBox(templateBox, t, "pageIndex.html")
	loadTemplateFromBox(templateBox, t, "pageShow.html")

	return &Template{templates: t}
}
