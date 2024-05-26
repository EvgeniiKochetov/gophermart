package handlers

import (
	"context"
	"strings"
	"time"

	//	"database/sql/driver"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gophermart/src/config"
	"gophermart/src/databases"
	"net/http"
)

type User struct {
	Name string `json:"login"`
	Pwd  string `json:"password"`
}

type Claim struct {
	Name string `json:"name"`
	jwt.StandardClaims
}

func RegisterUser(writer http.ResponseWriter, request *http.Request) {
	user := User{}
	err := json.NewDecoder(request.Body).Decode(&user)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if len(user.Name) == 0 || len(user.Pwd) == 0 {
		http.Error(writer, "Empty name or password. Check request", http.StatusBadRequest)
		return
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(user.Pwd), bcrypt.DefaultCost)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	num, err := databases.AddUser(user.Name, string(hashedPwd))
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	if num == 0 {
		http.Error(writer, "This login has used yet", http.StatusConflict)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func LoginUser(writer http.ResponseWriter, request *http.Request) {

	user := User{}
	err := json.NewDecoder(request.Body).Decode(&user)
	if err != nil {
		http.Error(writer, "error when encode", http.StatusBadRequest)
		return
	}

	storedPwd, err := databases.GetPassword(user.Name)
	if err != nil {
		http.Error(writer, "error in authorization", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPwd), []byte(user.Pwd))

	if err != nil {
		http.Error(writer, "invaliv password", http.StatusUnauthorized)
		return
	}

	expiredAt := time.Now().Add(72 * time.Hour)
	claim := &Claim{Name: user.Name, StandardClaims: jwt.StandardClaims{ExpiresAt: expiredAt.Unix()}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	tokenString, err := token.SignedString([]byte(config.GetSecretKey()))

	if err != nil {
		http.Error(writer, "error when generate token", http.StatusInternalServerError)
		return
	}

	writer.Write([]byte(tokenString))
}

func Auth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		tokenRequest := r.Header.Get("Authorization")
		if tokenRequest == "" {
			http.Error(w, "don't find authorization data", http.StatusUnauthorized)
			return
		}
		authData := strings.Split(tokenRequest, " ")
		if len(authData) != 2 {
			http.Error(w, "don't find authorization data", http.StatusUnauthorized)
			return
		}
		tokenString := authData[1]
		token, err := jwt.Parse(tokenString,
			func(token *jwt.Token) (interface{}, error) { return []byte(config.GetSecretKey()), nil })

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		claim := token.Claims.(jwt.MapClaims)

		userId, err := databases.GetUserId(claim["name"].(string))
		if err != nil {
			http.Error(w, "don't find username", http.StatusUnauthorized)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), "userid", userId))

		handlerFunc(w, r)
	}

}
