package main

import (
	"os"

	"github.com/vigliag/isamuni-go/model"
	"github.com/vigliag/isamuni-go/web"
)

func main() {
	model.Connect("data/database.db")
	p := model.FindPage(71, model.PageUser)
	content := p.Content
	//content = strings.Replace(content, "\r\n", "\n", -1)
	os.Stdout.WriteString(string(web.RenderMarkdown(content)))
}
