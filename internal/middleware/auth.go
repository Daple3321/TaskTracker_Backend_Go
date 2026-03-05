package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

type AuthHandler struct {
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) RegisterRoutes() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("POST /", LoginHandler)

	return r
}

func CreateToken(username string, password string) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"password": password,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) error {
	token, err := jwt.Parse(tokenString,
		func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})
	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("Invalid token")
	}

	return nil
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var u User
	json.NewDecoder(r.Body).Decode(&u)
	//fmt.Printf("The user request value %v", u)

	if u.Username == "" || u.Password == "" {
		http.Error(w, "Invalid credentials", http.StatusBadRequest)
		return
	}

	tokenString, err := CreateToken(u.Username, u.Password)
	if err != nil {
		http.Error(w, "Error creating token", http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, tokenString)
}

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Token not provided.", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		err := VerifyToken(tokenString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
