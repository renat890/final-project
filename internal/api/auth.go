package api

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"tracker/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type pass struct {
	Password string `json:"password"`
}

type TokenBody struct {
	Token string `json:"token"`
}

type Claims struct {
	Data string
	jwt.RegisteredClaims
}

var (
	ErrWrongPassword = errors.New("неверный пароль")
	ErrFailSignToken = errors.New("не смог подписать токен")
	secretKey        = []byte("vouis7y6gtf3897o4hfbewro87fvgoe*&()GET)C(V(GSs96r87v))") // так лучше не делать, лучше хранить в Vault
)

func makeHandleAuth(conf *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body pass
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			writeResponse(w, "error", err.Error(), http.StatusBadRequest)
			return 
		}

		if body.Password != conf.TodoPass {
			writeResponse(w, "error", ErrWrongPassword.Error(), http.StatusUnauthorized)
			return 
		}

		// в качестве payload формируем контрольную сумму из пароля
		sum := sha256.Sum256([]byte(body.Password))
		claims := Claims{Data: fmt.Sprintf("%x", sum)}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signingToken, err := token.SignedString(secretKey)

		if err != nil {
			writeResponse(w, "error", "внутренняя ошибка", http.StatusInternalServerError)
			return
		}

		type response struct {
			Token string `json:"token"`
		}

		writeJson(w, response{Token: signingToken})
	}
}

func authMiddleware(next http.HandlerFunc, conf *config.Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(conf.TodoPass) > 0 {
			var jwtValue string

			cookie, err := r.Cookie("token")
			if err == nil {
				jwtValue = cookie.Value
			}

			token, err := jwt.ParseWithClaims(jwtValue, &Claims{}, func(t *jwt.Token) (any, error) {return secretKey, nil})
			if err != nil {
				writeResponse(w, "error", "внутренняя ошибка", http.StatusUnauthorized)
				return 
			}

			claims, ok := token.Claims.(*Claims)
			if !ok {
				writeResponse(w, "error", "невалидный токен", http.StatusUnauthorized)
				return
			}

			sum := sha256.Sum256([]byte(conf.TodoPass))
			if !token.Valid || claims.Data != fmt.Sprintf("%x", sum) {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
                return
			}
		}
		next.ServeHTTP(w, r)
	})
}