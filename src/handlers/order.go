package handlers

import (
	"encoding/json"
	"github.com/theplant/luhn"
	"gophermart/src/config"
	"gophermart/src/databases"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
)

func AddOrder(writer http.ResponseWriter, request *http.Request) {

	defer request.Body.Close()
	body, err := io.ReadAll(request.Body)

	number, err := strconv.Atoi(string(body))

	if err != nil {
		http.Error(writer,
			"wrong format of request", http.StatusBadRequest)
		return
	}

	valid := luhn.Valid(number)

	if !valid {
		http.Error(writer, "invalid number of order", http.StatusUnprocessableEntity)
	}
	id, ok := request.Context().Value("userid").(int)

	if !ok {
		http.Error(writer, "invalid number of order", http.StatusInternalServerError)
	}

	status, err := databases.AddOrder(number, id)
	writer.WriteHeader(status)
}

func GetOrders(writer http.ResponseWriter, request *http.Request) {

	id, ok := request.Context().Value("userid").(int)

	if !ok {
		http.Error(writer, "invalid number of order", http.StatusInternalServerError)
		return
	}

	orders, err := databases.GetOrders(id)
	if err != nil {
		http.Error(writer, "error in getting orders", http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		http.Error(writer, "not data for answer", http.StatusNoContent)
		return
	}
	answer, err := json.Marshal(orders)

	writer.Write(answer)
}

func Orders(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		AddOrder(writer, request)
	} else if request.Method == "GET" {
		GetOrders(writer, request)
	}
}

func Order(writer http.ResponseWriter, request *http.Request) {

	limiter := config.GetLimiter()

	if !limiter.Allow() {
		http.Error(writer, "No more than 1 requests per minute allowed", http.StatusTooManyRequests)
		return
	}
	path := request.URL.Path

	_, orderStr := filepath.Split(path)
	orderId, err := strconv.Atoi(orderStr)

	if err != nil {
		http.Error(writer, "Error when convert from string to int", http.StatusInternalServerError)
		return
	}
	order, err := databases.GetOrder(orderId)
	if err != nil {
		http.Error(writer, "Error when get data from database", http.StatusInternalServerError)
		return
	}

	answer, err := json.Marshal(order)

	writer.Write(answer)
}
