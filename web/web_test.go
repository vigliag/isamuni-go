package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/vigliag/isamuni-go/model"
)

func panicIfNotNull(e error) {
	if e != nil {
		panic(e)
	}
}

func formRequest(addr string, values url.Values) *http.Request {
	req := httptest.NewRequest("POST", addr, strings.NewReader(values.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func assertHTMLReturned(t *testing.T, res *http.Response) {
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", strings.ToLower(res.Header.Get("content-type")))
}

func registerTestAdmin() *model.User {
	u, err := model.RegisterEmail("vigliag", "vigliag@gmail.com", "password", "admin")
	if err != nil {
		panic(err)
	}
	return u
}

func registerTestUser() *model.User {
	u, err := model.RegisterEmail("otheruser", "other@example.com", "password", "")
	if err != nil {
		panic(err)
	}
	return u
}

func TestShowPageHandler(t *testing.T) {
	model.ConnectTestDB()

	r := CreateServer(echo.New())

	p := model.Page{
		Content: "Ciao",
		Type:    model.PageCompany,
		Title:   "Example company",
	}
	model.Db.Save(&p)
	assert.NotZero(t, p.ID)

	client := NewTestClient(r)
	res := client.Get(fmt.Sprintf("/companies/%d", p.ID))
	assertHTMLReturned(t, res)

	res = client.Get("/companies/notExisting")
	assert.Equal(t, 404, res.StatusCode)
}

func TestMeHandler(t *testing.T) {
	model.ConnectTestDB()
	defer model.Close()

	u := registerTestAdmin()

	r := CreateServer(echo.New())

	// Before login
	client := NewTestClient(r)
	res := client.Get("/me")
	assert.Equal(t, http.StatusFound, res.StatusCode)

	redirectURL, _ := res.Location()
	assert.Equal(t, "/login", redirectURL.Path)

	// After login
	client.MustLogin(*u.Email, "password")
	res = client.Get("/me")
	assert.Equal(t, 200, res.StatusCode)
}

func TestLogin(t *testing.T) {
	model.ConnectTestDB()
	defer model.Db.Close()

	u := registerTestAdmin()

	r := CreateServer(echo.New())
	client := NewTestClient(r)

	res := client.Login(*u.Email, "password")
	if res.StatusCode >= 400 {
		assert.Fail(t, "Login returned error status code")
	}
	assert.NotEmpty(t, res.Cookies)

	client = NewTestClient(r)
	res = client.Login(*u.Email, "wrongPassword")
	assert.Equal(t, 404, res.StatusCode)
}

func TestInsertPage(t *testing.T) {
	model.ConnectTestDB()
	defer model.Db.Close()

	u := registerTestAdmin()

	r := CreateServer(echo.New())
	client := NewTestClient(r)
	client.MustLogin(*u.Email, "password")

	form := url.Values{}
	form.Set("title", "Example company")
	form.Set("type", fmt.Sprintf("%d", int(model.PageCompany)))
	form.Set("content", "Example contents")

	req := formRequest("/pages", form)
	res := client.Run(req)

	assert.Equal(t, http.StatusSeeOther, res.StatusCode)

	createdPageURL, err := res.Location()
	assert.NoError(t, err)

	// verify new page loads correctly
	res = client.Get(createdPageURL.String())
	assertHTMLReturned(t, res)
}

func TestEditPage(t *testing.T) {
	model.ConnectTestDB()
	defer model.Close()

	u := registerTestAdmin()
	u2 := registerTestUser()

	p := &model.Page{
		Content: "Ciao",
		Type:    model.PageCompany,
		Title:   "Example company",
	}

	err := model.SavePage(p, u)
	assert.NoError(t, err)

	r := CreateServer(echo.New())

	// Test with admin
	client := NewTestClient(r)
	client.MustLogin(*u.Email, "password")
	res := client.Get(fmt.Sprintf("/companies/%d/edit", p.ID))
	assertHTMLReturned(t, res)

	// Test with non-admin
	client2 := NewTestClient(r)
	client2.MustLogin(*u2.Email, "password")
	res = client2.Get(fmt.Sprintf("/companies/%d/edit", p.ID))
	assertHTMLReturned(t, res)
}

func TestUserPages(t *testing.T) {
	r := CreateServer(echo.New())
	client := NewTestClient(r)

	assertHTMLReturned(t, client.Get("/login"))
	//assertHTMLReturned(t, client.Get("/companies/new"))
}
