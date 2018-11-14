package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

func loginRequest(username, password string) *http.Request {
	form := url.Values{}
	form.Set("email", username)
	form.Set("password", password)

	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return req
}

type TestClient struct {
	cookies []*http.Cookie
	h       http.Handler
}

func NewTestClient(h http.Handler) *TestClient {
	return &TestClient{h: h}
}

func (c *TestClient) Login(email, password string) *http.Response {
	w := httptest.NewRecorder()
	req := loginRequest(email, password)

	c.h.ServeHTTP(w, req)
	res := w.Result()
	c.cookies = res.Cookies()
	return res
}

func (c *TestClient) MustLogin(email, password string) {
	res := c.Login(email, password)
	if res.StatusCode >= 400 {
		panic("couldNotLogin")
	}
}

func (c *TestClient) Run(req *http.Request) *http.Response {
	w := httptest.NewRecorder()
	for _, cookie := range c.cookies {
		req.AddCookie(cookie)
	}
	c.h.ServeHTTP(w, req)
	return w.Result()
}

func (c *TestClient) Get(url string) *http.Response {
	return c.Run(httptest.NewRequest("GET", url, nil))
}
