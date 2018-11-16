package cmd

import (
	"crypto/tls"
	"fmt"
	"path"

	"github.com/spf13/viper"
	"github.com/vigliag/isamuni-go/index"
	"github.com/vigliag/isamuni-go/mail"
	"github.com/vigliag/isamuni-go/model"
	"github.com/vigliag/isamuni-go/web"
)

func GetMailer() *mail.SmtpServer {
	return &mail.SmtpServer{
		Host:     viper.GetString("MAILER_HOSTNAME"),
		Password: viper.GetString("MAILER_PASSWORD"),
		Port:     viper.GetInt("MAILER_PORT"),
		User:     viper.GetString("MAILER_USERNAME"),
		TlsConfig: &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         viper.GetString("MAILER_HOST"),
		},
	}
}

func GetController() *web.Controller {
	dataPath := viper.GetString("data")
	dbPath := path.Join(dataPath, "database.db")
	indexPath := path.Join(dataPath, "index.bleve")
	appURL := viper.GetString("APP_URL")
	fmt.Println("Using data folder", dataPath)
	fmt.Println("Using app url", appURL)

	db := model.Connect(dbPath)
	m := model.New(db)
	bleveidx := index.OpenOrNewBleve(indexPath)
	idx := index.New(bleveidx, m)
	mailer := GetMailer()
	ctl := web.NewController(appURL, m, idx, mailer)
	return ctl
}
