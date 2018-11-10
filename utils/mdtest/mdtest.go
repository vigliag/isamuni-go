package main

import (
	"fmt"

	"github.com/vigliag/isamuni-go/model"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

func main() {
	model.Connect("data/database.db")
	p := model.FindPage(62, model.PageUser)
	fmt.Println(p.Content)

	unsafe := blackfriday.Run([]byte(p.Content))
	fmt.Println(string(unsafe))
}
