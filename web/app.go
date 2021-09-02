package web

import (
	"embed"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/trento-project/trento/internal/consul"
	"github.com/trento-project/trento/web/services"
)

//go:embed frontend/assets
var assetsFS embed.FS

//go:embed templates
var templatesFS embed.FS

type App struct {
	host string
	port int
	Dependencies
}

type Dependencies struct {
	consul         consul.Client
	engine         *gin.Engine
	sessionsStore  sessions.Store
	authMiddleware gin.HandlerFunc
	usersService   services.UsersService
}

func DefaultDependencies() Dependencies {
	consulClient, _ := consul.DefaultClient()
	engine := gin.Default()
	sessionsStore := sessions.NewCookieStore([]byte("secret"))
	usersService := services.NewUsersService()
	authMiddleware := AuthRequired

	return Dependencies{consulClient, engine, sessionsStore, authMiddleware, usersService}
}

// shortcut to use default dependencies
func NewApp(host string, port int) (*App, error) {
	return NewAppWithDeps(host, port, DefaultDependencies())
}

func NewAppWithDeps(host string, port int, deps Dependencies) (*App, error) {
	app := &App{
		Dependencies: deps,
		host:         host,
		port:         port,
	}

	engine := deps.engine
	engine.Use(sessions.Sessions("trento", deps.sessionsStore))

	engine.HTMLRender = NewLayoutRender(templatesFS, "templates/*.tmpl")
	engine.Use(ErrorHandler)
	engine.StaticFS("/static", http.FS(assetsFS))

	engine.GET("/login", NewLoginPageHandler())
	engine.POST("/login", NewLoginHandler(deps.usersService))
	engine.GET("/logout", NewLogoutHandler())
	private := engine.Group("/")
	private.Use(deps.authMiddleware)
	private.GET("/", HomeHandler)
	private.GET("/hosts", NewHostListHandler(deps.consul))
	private.GET("/hosts/:name", NewHostHandler(deps.consul))
	private.GET("/hosts/:name/ha-checks", NewHAChecksHandler(deps.consul))
	private.GET("/clusters", NewClusterListHandler(deps.consul))
	private.GET("/clusters/:id", NewClusterHandler(deps.consul))
	private.GET("/sapsystems", NewSAPSystemListHandler(deps.consul))
	private.GET("/sapsystems/:sid", NewSAPSystemHandler(deps.consul))

	apiGroup := private.Group("/api")
	{
		apiGroup.GET("/ping", ApiPingHandler)
	}

	return app, nil
}

func (a *App) Start() error {
	s := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", a.host, a.port),
		Handler:        a,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return s.ListenAndServe()
}

func (a *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	a.engine.ServeHTTP(w, req)
}
