package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/gosimple/slug"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vigliag/isamuni-go/model"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "isamuni"
	password = "your-password"
	dbname   = "isamuni_prod"
)

// indexCmd represents the index command
var importIsamuniCmd = &cobra.Command{
	Use:   "importisamuni",
	Short: "Re-create the index, and optionally search for a term",
	Run:   importIsamuniRun,
}

func getTags(isamunidb *sql.DB) map[int]string {

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

	return tags
}

func copyUsers(isamunidb *sql.DB, m *model.Model) {
	tags := getTags(isamunidb)

	rows, err := isamunidb.Query("select id, name, occupation, description, projects, links, tags, uid from users where public_profile= true;")
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
			fbid        sql.NullString
		)
		if err := rows.Scan(&id, &name, &occupation, &description, &projects, &links, &oldtags, &fbid); err != nil {
			log.Fatal(err)
		}

		userTags, hasTags := tags[id]
		if !hasTags {
			userTags = oldtags.String
		}

		u := model.User{
			FacebookID: &fbid.String,
			Username:   name.String,
		}
		err := m.SaveUser(&u)

		p := model.Page{
			Title:   name.String,
			Short:   occupation.String,
			Type:    model.PageUser,
			Slug:    slug.Make(name.String),
			OwnerID: u.ID,
		}
		var b strings.Builder
		if s := strings.TrimSpace(occupation.String); s != "" {
			b.WriteString("### In breve\n\n")
			b.WriteString(s)
		}
		if s := strings.TrimSpace(description.String); s != "" {
			b.WriteString("\n\n### Descrizione\n\n")
			b.WriteString(s)
		}
		if s := strings.TrimSpace(projects.String); s != "" {
			b.WriteString("\n\n### Progetti\n\n")
			b.WriteString(s)
		}
		if s := strings.TrimSpace(links.String); s != "" {
			b.WriteString("\n\n### Links\n\n")
			b.WriteString(s)
		}
		if s := strings.TrimSpace(userTags); s != "" {
			b.WriteString("\n\n### Tags\n\n")
			b.WriteString(s)
		}
		p.Content = strings.Replace(b.String(), "\r\n", "\n", -1)

		res := m.Db.Save(&p)
		if res.Error != nil {
			panic(res.Error)
		}
		fmt.Printf("user %s: %s\n%s\n\n", p.Title, p.Short, p.Content)

		if err != nil {
			panic(err)
		}

		cv := model.ContentVersion{
			Content: p.Content,
			PageID:  p.ID,
			UserID:  u.ID,
		}

		res = m.Db.Save(&cv)
		if res.Error != nil {
			fmt.Println(res.Error)
			continue
		}

		p.ApprovedVersionID = cv.ID
		res = m.Db.Save(&p)
		if res.Error != nil {
			fmt.Println(res.Error)
			continue
		}
	}
}

func copyPages(isamunidb *sql.DB, m *model.Model) {
	q := `select name, kind, website,
				concat_ws(chr(10), links, fbpage, twitterpage) as links,
				location, province, description, sector
		 from pages;`
	rows, err := isamunidb.Query(q)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var (
			name        string
			kind        int
			website     sql.NullString
			links       sql.NullString
			location    sql.NullString
			province    sql.NullString
			description sql.NullString
			sector      sql.NullString
		)

		err = rows.Scan(&name, &kind, &website, &links, &location, &province, &description, &sector)
		if err != nil {
			fmt.Println(err)
			continue
		}

		linksString := strings.TrimSpace(links.String)
		var linksLines []string
		for _, line := range strings.Fields(linksString) {
			linksLines = append(linksLines, "- "+line)
		}
		linksString = strings.Join(linksLines, "\n")

		var b strings.Builder
		b.WriteString(description.String)
		b.WriteString("\n\n### Links\n\n")
		b.WriteString(linksString)
		b.WriteString("\n\n### Dati\n\n")
		if s := strings.TrimSpace(website.String); s != "" {
			fmt.Fprintf(&b, "- Sito web: %v\n", s)
		}
		if s := strings.TrimSpace(location.String); s != "" {
			fmt.Fprintf(&b, "- Posizione: %v\n", s)
		}
		if s := strings.TrimSpace(province.String); s != "" {
			fmt.Fprintf(&b, "- Città: %v\n", s)
		}
		if s := strings.TrimSpace(sector.String); s != "" {
			fmt.Fprintf(&b, "- Settore: %v\n", s)
		}
		b.WriteString("\n")

		content := b.String()
		content = strings.Replace(content, "\r\n", "\n", -1)

		p := model.Page{
			Title:   name,
			Short:   sector.String,
			Slug:    slug.Make(name),
			Content: content,
			Sector:  sector.String,
			Website: website.String,
			Type:    kindToPageType(kind),
			City:    province.String,
		}

		res := m.Db.Save(&p)
		if res.Error != nil {
			fmt.Println(res.Error)
			continue
		}

		cv := model.ContentVersion{
			Content: content,
			PageID:  p.ID,
		}
		res = m.Db.Save(&cv)
		if res.Error != nil {
			fmt.Println(res.Error)
			continue
		}

		p.ApprovedVersionID = cv.ID

		res = m.Db.Save(&p)
		if res.Error != nil {
			fmt.Println(res.Error)
			continue
		}

	}
}

func importIsamuniRun(cmd *cobra.Command, args []string) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	dbname := path.Join(viper.GetString("data"), "database.db")
	m := &model.Model{model.Connect(dbname)}
	defer m.Close()

	isamunidb, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	defer isamunidb.Close()

	err = isamunidb.Ping()
	if err != nil {
		panic(err)
	}

	model.Connect(dbname)

	copyUsers(isamunidb, m)
	copyPages(isamunidb, m)
}

func kindToPageType(kind int) model.PageType {
	switch kind {
	case 1:
		return model.PageCommunity
	case 0:
		return model.PageCompany
	}
	return 0
}

func init() {
	rootCmd.AddCommand(importIsamuniCmd)
}
