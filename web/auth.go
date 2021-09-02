package web

import (
	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/trento-project/trento/web/services"
)

const (
	userkey = "user"
)

func NewLoginPageHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html.tmpl", gin.H{})
	}
}

func NewLoginHandler(usersService services.UsersService) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.PostForm("email")
		password := c.PostForm("password")

		if !usersService.AuthenticateByEmailPassword(email, password) {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		// Save the username in the session

		session := sessions.Default(c)
		session.Set(userkey, email) // In real world usage you'd set this to the users ID
		err := session.Save()
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		c.Redirect(http.StatusFound, "/")
	}
}

func NewLogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		c.Redirect(http.StatusFound, "/login")
	}
}

func AuthRequired(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	if user == nil {
		// Abort the request with the appropriate error code
		//c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
		return
	}
	// Continue down the chain to handler etc
	c.Next()
}
