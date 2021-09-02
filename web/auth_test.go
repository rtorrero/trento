package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trento-project/trento/web/services/mocks"
)

func TestLoginHandler(t *testing.T) {
	usersServiceMock := new(mocks.UsersService)
	usersServiceMock.On("AuthenticateByEmailPassword", mock.Anything, mock.Anything).Return(true)

	deps := defaultTestDependencies()
	deps.authMiddleware = AuthRequired
	deps.usersService = usersServiceMock

	app, err := NewAppWithDeps("", 80, deps)
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/login", strings.NewReader(getLoginFormPayload()))

	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "text/html")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(getLoginFormPayload())))

	app.ServeHTTP(resp, req)
	assert.Equal(t, 302, resp.Code)
	assert.Equal(t, "/", resp.Header().Get("Location"))
}

func TestLogoutHandler(t *testing.T) {
	deps := defaultTestDependencies()
	deps.authMiddleware = AuthRequired

	app, err := NewAppWithDeps("", 80, deps)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := http.NewRequest("POST", "/login", strings.NewReader(getLoginFormPayload())); err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/logout", nil)

	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "text/html")

	app.ServeHTTP(resp, req)
	assert.Equal(t, 302, resp.Code)
	assert.Equal(t, "/login", resp.Header().Get("Location"))
}

func TestAuthRequired(t *testing.T) {
	deps := defaultTestDependencies()
	deps.authMiddleware = AuthRequired

	app, err := NewAppWithDeps("", 80, deps)
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "text/html")

	app.ServeHTTP(resp, req)
	assert.Equal(t, 302, resp.Code)
	assert.Equal(t, "/login", resp.Header().Get("Location"))
}

func getLoginFormPayload() string {
	params := url.Values{}
	params.Add("email", "banana@suse.com")
	params.Add("password", "potato")

	return params.Encode()
}
