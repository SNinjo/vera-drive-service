package middleware

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"vera-identity-service/internal/apperror"
	"vera-identity-service/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	jwt.RegisteredClaims
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

type AuthHandler gin.HandlerFunc

func NewAuthHandler(config *config.Config) AuthHandler {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		req, err := http.NewRequest("POST", config.IdentityServiceURL+"/auth/verify", nil)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}
		req.Header.Set("Authorization", authHeader)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.Error(apperror.New(apperror.CodeIdentityServiceUnavailable, "failed to call identity service | "+err.Error()))
			c.Abort()
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				c.Error(err)
				c.Abort()
				return
			}

			c.Data(resp.StatusCode, "application/json", body)
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		userClaims := &UserClaims{}
		_, _, err = new(jwt.Parser).ParseUnverified(token, userClaims)
		if err != nil {
			c.Error(apperror.New(apperror.CodeInvalidClaimsInUserToken, "invalid claims in user token | "+err.Error()))
			c.Abort()
			return
		}

		userID, err := strconv.Atoi(userClaims.Subject)
		if err != nil {
			c.Error(apperror.New(apperror.CodeInvalidClaimsInUserToken, "invalid user ID in token | "+err.Error()))
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
