package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
	"vera-identity-service/internal/middleware"

	"github.com/golang-jwt/jwt/v5"
)

func SetupIdentityService(expectedTokenSecret string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/verify", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("failed to parse form | %v", err)))
			return
		}
		authHeader := r.Header.Get("Authorization")

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims := middleware.UserClaims{}
		_, err = jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(expectedTokenSecret), nil
		})

		if err == nil {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(
				fmt.Sprintf(`{
					"code": "mock_error_code",
					"message": "Unauthorized",
					"timestamp": "%s"
				}`, time.Now().Format(time.RFC3339)),
			))
		}
	})
	return httptest.NewServer(mux)
}
