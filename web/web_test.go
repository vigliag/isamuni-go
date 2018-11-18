package web

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/vigliag/isamuni-go/mail"

	"github.com/vigliag/isamuni-go/index"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/vigliag/isamuni-go/model"
)

type TestEnvironment struct {
	model   *model.Model
	index   *index.Index
	ctl     *Controller
	app     *echo.Echo
	mailer  *mail.TestMailer
	tempdir string
}

func (t *TestEnvironment) Close() {
	t.model.Close()
	t.index.Close()
}

func GetTestEnv() *TestEnvironment {
	appURL := "http://localhost:8080"
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		log.Fatal(err)
	}
	mailer := mail.TestMailer{}
	m := model.New(model.ConnectTestDB())
	bleveidx, err := index.NewBleve(dir)
	panicIfNotNull(err)
	idx := index.New(bleveidx, m)
	ctl := NewController(appURL, m, idx, &mailer)
	return &TestEnvironment{
		m, idx, ctl, CreateServer(echo.New(), ctl), &mailer, dir,
	}
}

func (t *TestEnvironment) TestClient() *TestClient {
	return NewTestClient(t.app)
}

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

func (t *TestEnvironment) registerTestAdmin() *model.User {
	u, err := t.model.RegisterEmail("vigliag", "vigliag@gmail.com", "password", "admin")
	if err != nil {
		panic(err)
	}
	return u
}

func (t *TestEnvironment) registerTestUser() *model.User {
	u, err := t.model.RegisterEmail("otheruser", "other@example.com", "password", "")
	if err != nil {
		panic(err)
	}
	return u
}

func TestShowPageHandler(t *testing.T) {
	env := GetTestEnv()
	defer env.Close()

	p := model.Page{
		Content: "Ciao",
		Type:    model.PageCompany,
		Title:   "Example company",
	}
	env.model.Db.Save(&p)
	assert.NotZero(t, p.ID)

	client := env.TestClient()
	res := client.Get(fmt.Sprintf("/companies/%d", p.ID))
	assertHTMLReturned(t, res)

	res = client.Get("/companies/notExisting")
	assert.Equal(t, 404, res.StatusCode)
}

func TestMeHandler(t *testing.T) {
	env := GetTestEnv()
	defer env.Close()

	u := env.registerTestAdmin()

	// Before login
	client := env.TestClient()
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
	env := GetTestEnv()
	defer env.Close()
	client := env.TestClient()

	u := env.registerTestAdmin()
	res := client.Login(*u.Email, "password")
	if res.StatusCode >= 400 {
		assert.Fail(t, "Login returned error status code")
	}
	assert.NotEmpty(t, res.Cookies)

	client = env.TestClient()
	res = client.Login(*u.Email, "wrongPassword")
	assert.Equal(t, 404, res.StatusCode)
}

func TestInsertPage(t *testing.T) {
	env := GetTestEnv()
	defer env.Close()

	u := env.registerTestAdmin()
	client := env.TestClient()
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
	env := GetTestEnv()
	defer env.Close()

	u := env.registerTestAdmin()
	u2 := env.registerTestUser()

	p := &model.Page{
		Content: "Ciao",
		Type:    model.PageCompany,
		Title:   "Example company",
	}

	err := env.model.SavePage(p, u)
	assert.NoError(t, err)

	// Test with admin
	client := env.TestClient()
	client.MustLogin(*u.Email, "password")
	res := client.Get(fmt.Sprintf("/companies/%d/edit", p.ID))
	assertHTMLReturned(t, res)

	// Test with non-admin
	client2 := env.TestClient()
	client2.MustLogin(*u2.Email, "password")
	res = client2.Get(fmt.Sprintf("/companies/%d/edit", p.ID))
	assertHTMLReturned(t, res)
}

func TestUserPages(t *testing.T) {
	env := GetTestEnv()
	defer env.Close()
	client := env.TestClient()
	assertHTMLReturned(t, client.Get("/login"))
}

func TestSearch(t *testing.T) {
	env := GetTestEnv()
	defer env.Close()
	client := env.TestClient()
	assertHTMLReturned(t, client.Get("/search"))
	assertHTMLReturned(t, client.Get("/search?query=promuove"))
}

func TestSendMailVerification(t *testing.T) {
	env := GetTestEnv()
	defer env.Close()

	// Send the verification mail
	u := env.registerTestUser()
	u.EmailVerified = false
	err := env.model.Db.Save(&u).Error
	panicIfNotNull(err)

	err = env.ctl.SendEmailVerification(u)
	assert.NoError(t, err)
	assert.NotEmpty(t, env.mailer.Mails)

	// Get the verification url from the mail
	mail := env.mailer.Mails[len(env.mailer.Mails)-1]
	urlRegex := regexp.MustCompile(`http[s]?:\/\/\S+`)
	fmt.Println(mail.Body)
	confirmationAddr := urlRegex.FindString(mail.Body)
	assert.NotEmpty(t, confirmationAddr)

	// Visit the verification url
	client := env.TestClient()
	res := client.Get(confirmationAddr)
	assert.Equal(t, http.StatusFound, res.StatusCode)

	// Check the user email was verified
	u = env.model.RetrieveUser(u.ID)
	assert.Equal(t, true, u.EmailVerified)
}
