package web

import (
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func defaultTestDependencies() Dependencies {
	engine := gin.Default()
	authMiddleware := func(c *gin.Context) {}
	sessionsStore := sessions.NewCookieStore([]byte("secret"))

	return Dependencies{
		engine:         engine,
		authMiddleware: authMiddleware,
		sessionsStore:  sessionsStore,
	}
}
