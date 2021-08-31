package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trento-project/trento/web/services"
)

func NewLoginPageHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html.tmpl", gin.H{})
	}
}

func NewLoginHandler(usersService services.IUsersService) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.PostForm("email")
		password := c.PostForm("password")

		if !usersService.AuthenticateByEmailPassword(email, password) {
			c.JSON(http.StatusForbidden, gin.H{"message": "Couldn't log in"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Successfully logged in"})
	}
}
