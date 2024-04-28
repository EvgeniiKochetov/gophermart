package handlers

import (
	"net/http"
)

func HomeHandler(writer http.ResponseWriter, request *http.Request) {

	http.ServeFile(writer, request, "./internal/index.html")

}
