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

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func testHTMLResponse(t *testing.T, router *echo.Echo, path string) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", path, nil)
	router.ServeHTTP(w, req)

	log.Println("Testing url ", path)

	headers := w.Header()
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "text/html; charset=utf-8", strings.ToLower(headers.Get("content-type")))
}

func testResponseCode(t *testing.T, router *echo.Echo, path string, code int) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", path, nil)
	router.ServeHTTP(w, req)

	log.Println("Testing url ", path)
	assert.Equal(t, code, w.Code)
}
func TestPageHandler(t *testing.T) {
	db.ConnectTestDB()

	r := echo.New()
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

func loginRequest(username, password string) *http.Request {
	form := url.Values{}
	form.Set("email", username)
	form.Set("password", password)

	req, _ := http.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func TestLogin(t *testing.T) {
	db.ConnectTestDB()

	_, err := db.RegisterEmail("vigliag", "vigliag@gmail.com", "password")
	assert.NoError(t, err)

	r := echo.New()
	createServer(r)

	w := httptest.NewRecorder()
	req := loginRequest("vigliag@gmail.com", "password")
	r.ServeHTTP(w, req)
	res := w.Result()
	assert.Equal(t, 200, res.StatusCode)
	assert.NotEmpty(t, res.Cookies)

	w = httptest.NewRecorder()
	req = loginRequest("vigliag@gmail.com", "wrongPassword")
	r.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	db.Db.Close()
}

func login(h http.Handler, email string, password string) *http.Response {
	w := httptest.NewRecorder()
	req := loginRequest(email, password)

	h.ServeHTTP(w, req)
	res := w.Result()
	if res.StatusCode != 200 {
		panic("Login failed")
	}
	return res
}

func copyCookies(res *http.Response, req *http.Request) {
	for _, cookie := range res.Cookies() {
		req.AddCookie(cookie)
	}
}

func formRequest(addr string, values url.Values) *http.Request {
	req, err := http.NewRequest("POST", addr, strings.NewReader(values.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		panic(err)
	}
	return req
}

func TestInsertPage(t *testing.T) {
	db.ConnectTestDB()
	db.RegisterEmail("vigliag", "vigliag@gmail.com", "password")

	r := echo.New()
	createServer(r)

	lres := login(r, "vigliag@gmail.com", "password")
	form := url.Values{}

	form.Set("title", "Example company")
	form.Set("type", fmt.Sprintf("%d", int(db.PageCompany)))
	form.Set("content", "Example contents")

	req := formRequest("/companies", form)
	copyCookies(lres, req)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	res := w.Result()

	assert.Equal(t, 307, res.StatusCode)

	createdPageUrl, err := res.Location()
	assert.NoError(t, err)

	// verify new page loads correctly
	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", createdPageUrl.String(), nil)
	r.ServeHTTP(w, req)
	res = w.Result()
	assert.Equal(t, 200, res.StatusCode)
}
func TestUserPages(t *testing.T) {
	r := echo.New()
	createServer(r)

	testHTMLResponse(t, r, "/login")
}
