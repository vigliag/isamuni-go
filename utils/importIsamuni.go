package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/vigliag/isamuni-go/db"

	"github.com/gosimple/slug"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "isamuni"
	password = "your-password"
	dbname   = "isamuni_prod"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	isamunidb, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	defer isamunidb.Close()

	err = isamunidb.Ping()
	if err != nil {
		panic(err)
	}

	db.Connect()

	tags := make(map[int]string)

	rows, err := isamunidb.Query("select ts.taggable_id, string_agg(t.name, '; ') from tags t join taggings ts on t.id=ts.tag_id group by ts.taggable_id;")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var (
			id  int
			tgs string
		)
		err := rows.Scan(&id, &tgs)
		if err != nil {
			panic(err)
		}
		tags[id] = tgs
	}

	rows, err = isamunidb.Query("select id, name, occupation, description, projects, links, tags from users where public_profile= true;")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var (
			id          int
			name        sql.NullString
			occupation  sql.NullString
			description sql.NullString
			projects    sql.NullString
			links       sql.NullString
			oldtags     sql.NullString
		)
		if err := rows.Scan(&id, &name, &occupation, &description, &projects, &links, &oldtags); err != nil {
			log.Fatal(err)
		}

		userTags, hasTags := tags[id]
		if !hasTags {
			userTags = oldtags.String
		}

		p := db.Page{
			Title: name.String,
			Short: occupation.String,
			Type:  db.PageUser,
			Slug:  slug.Make(name.String),
		}
		var b strings.Builder

		s := strings.TrimSpace(description.String)
		b.WriteString("### Description\n")
		b.WriteString(s)

		if s := strings.TrimSpace(projects.String); s != "" {
			b.WriteString("\n\n### Projects\n")
			b.WriteString(s)
		}
		if s := strings.TrimSpace(links.String); s != "" {
			b.WriteString("\n\n### Links\n")
			b.WriteString(s)
		}
		if s := strings.TrimSpace(userTags); s != "" {
			b.WriteString("\n\n### Tags\n")
			b.WriteString(s)
		}
		p.Content = b.String()

		res := db.Db.Save(&p)
		if res.Error != nil {
			panic(res.Error)
		}
		fmt.Printf("user %s: %s\n%s\n\n", p.Title, p.Short, p.Content)
	}

	//Pages

}

func kindToPageType(kind int) db.PageType {
	switch kind {
	case 0:
		return db.PageCommunity
	case 1:
		return db.PageCompany
	}
	return 0
}
