package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"vigliag/commwiki/db"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func testHTMLResponse(t *testing.T, router *gin.Engine, path string) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", path, nil)
	router.ServeHTTP(w, req)

	log.Println("Testing url ", path)

	headers := w.Header()
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "text/html; charset=utf-8", headers.Get("content-type"))
}

func testResponseCode(t *testing.T, router *gin.Engine, path string, code int) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", path, nil)
	router.ServeHTTP(w, req)

	log.Println("Testing url ", path)
	assert.Equal(t, code, w.Code)
}
func TestPageHandler(t *testing.T) {
	db.ConnectTestDB()

	r := gin.New()
	createServer(r)

	p := db.Page{
		Content: "Ciao",
		Type:    db.PageCompany,
		Title:   "Example company",
	}
	db.Db.Save(&p)
	assert.NotZero(t, p.ID)

	testHTMLResponse(t, r, fmt.Sprintf("/companies/%d", p.ID))
	testResponseCode(t, r, "/companies/notExisting", 404)
}

func TestLogin(t *testing.T) {

	loginRequest := func(username, password string) *http.Request {
		form := url.Values{}
		form.Set("email", username)
		form.Set("password", password)

		req, _ := http.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		return req
	}

	db.ConnectTestDB()

	_, err := db.RegisterEmail("vigliag", "vigliag@gmail.com", "password")
	assert.NoError(t, err)

	r := gin.New()
	createServer(r)

	w := httptest.NewRecorder()
	req := loginRequest("vigliag@gmail.com", "password")
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	req = loginRequest("vigliag@gmail.com", "wrongPassword")
	r.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	db.Db.Close()
}

func TestUserPages(t *testing.T) {
	r := gin.New()
	createServer(r)
	testHTMLResponse(t, r, "/login")
}
