package main

import (
	"os"
	"path"

	"github.com/spf13/viper"
	"github.com/vigliag/isamuni-go/model"
	"github.com/vigliag/isamuni-go/web"
)

func main() {
	dbname := path.Join(viper.GetString("data"), "database.db")
	m := model.New(model.Connect(dbname))
	p := m.FindPage(71, model.PageUser)
	content := p.Content
	//content = strings.Replace(content, "\r\n", "\n", -1)
	os.Stdout.WriteString(string(web.RenderMarkdown(content)))
}
