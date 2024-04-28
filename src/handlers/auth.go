package handlers

import (
	"context"
	"encoding/json"
	"gophermart/internal"
	"gophermart/src/databases"
	"net/http"
)

type User struct {
	Name string `json:"login"`
	Pwd  string `json:"password"`
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

	num, err := databases.AddUser(user.Name, user.Pwd)
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
	var authHeader = request.Header.Get(`Authorization`)
	user_password, err := internal.GetUserPassword(authHeader)
	if err != nil {
		http.Error(writer,
			"error in authorization error", http.StatusBadRequest)
		return
	}

	userid, err := databases.CheckUser(user_password[0], user_password[1])

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	if userid == 0 {
		http.Error(writer, "wrong user/password", http.StatusUnauthorized)
		return
	}
}

func Auth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var authHeader = r.Header.Get(`Authorization`)
		userPassword, err := internal.GetUserPassword(authHeader)
		if err != nil {
			http.Error(w, "error in authorization error", http.StatusBadRequest)
			return
		}
		status, err := databases.CheckUser(userPassword[0], userPassword[1])

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "userid", status))

		handlerFunc(w, r)
	}

}
