package internal

import (
	"encoding/base64"
	"errors"
	"strings"
)

func GetUserPassword(authHeader string) ([]string, error) {
	userAuth := make([]string, 2)

	if authHeader == "" {
		return userAuth, errors.New("don't find users auth")
	}
	auth := strings.SplitN(authHeader, " ", 2)
	if len(auth) != 2 {

		return userAuth, errors.New("don't find users auth")
	}

	authUserPwd, err := base64.StdEncoding.DecodeString(auth[1])

	if err != nil {

		return userAuth, errors.New("error when decode user auth")
	}

	userAuth = strings.SplitN(string(authUserPwd), ":", 2)
	if len(userAuth) != 2 {

		return userAuth, errors.New("error when decode user auth")
	}
	return userAuth, nil
}
