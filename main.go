package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"vigliag/commwiki/db"

	"github.com/gorilla/securecookie"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func getSessionKey() []byte {
	sessKey := os.Getenv("SESSION_KEY")
	if sessKey == "" {
		fmt.Println("SESSION_KEY not set, generating new session key")
		return securecookie.GenerateRandomKey(32)
	}
	return []byte(sessKey)
}

var cookieStore = cookie.NewStore(getSessionKey())

func serveTemplate(templateName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, templateName+".html", nil)
	}
}

func loginEmail(c *gin.Context) {
	tplName := "loginPage.html"
	email := c.PostForm("email")
	password := c.PostForm("password")

	if email == "" || password == "" {
		log.Println("Empty email or password")
		c.HTML(http.StatusBadRequest, tplName, gin.H{"error": "Empty email or password"})
		return
	}

	user := db.LoginEmail(email, password)
	if user == nil {
		log.Println("Invalid email or password")
		c.HTML(http.StatusNotFound, tplName, gin.H{"error": "Invalid email or password"})
		return
	}

	session := sessions.Default(c)

	session.Clear()
	session.Set("usermail", email)
	session.Set("userid", user.ID)
	session.Save()

	c.String(200, "Logged in with mail %s", user.Email)
}

func showPageHandler(ptype db.PageType) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		if idParam == "" {
			log.Println("Invalid " + idParam)
			c.AbortWithStatus(404)
			return
		}

		id, err := strconv.Atoi(idParam)
		if err != nil {
			log.Println("Invalid " + idParam)
			c.AbortWithStatus(404)
			return
		}

		page := db.FindPage(uint(id), ptype)
		if page == nil {
			log.Println("Page not found for " + idParam)
			c.AbortWithStatus(404)
			return
		}

		c.HTML(200, "showPage.html", gin.H{"page": page})
	}
}

func updatePageHandler(c *gin.Context) {
	var p db.Page
	c.BindQuery(&p)

	res := db.Db.Save(p)
	if res.Error != nil {
		log.Println(res.Error)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
}

func createServer(r *gin.Engine) {
	r.LoadHTMLGlob("templates/*")
	r.Use(sessions.Sessions("defaultSession", cookieStore))
	r.GET("/login", serveTemplate("loginPage"))
	r.POST("/login", loginEmail)

	r.GET("/professionals/:id", showPageHandler(db.PageUser))
	r.GET("/wiki/:id", showPageHandler(db.PageWiki))
	r.GET("/companies/:id", showPageHandler(db.PageCompany))
	r.GET("/communities/:id", showPageHandler(db.PageCommunity))

	r.POST("/professionals/:id", updatePageHandler)
	r.POST("/wiki/:id", updatePageHandler)
	r.POST("/companies/:id", updatePageHandler)
	r.POST("/communities/:id", updatePageHandler)
	return
}

func main() {
	db.Connect()

	r := gin.Default()
	createServer(r)
	r.Run()

	fmt.Println("hi")
}
