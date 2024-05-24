package handlers

import (
	"encoding/json"
	"gophermart/src/databases"
	"gophermart/src/entities"
	"net/http"
)

func UserBalance(writer http.ResponseWriter, request *http.Request) {
	id, ok := request.Context().Value("userid").(int)

	if !ok {
		http.Error(writer, "invalid number of order", http.StatusInternalServerError)
		return
	}

	current, withdrawn, err := databases.GetBalance(id)
	if err != nil {
		http.Error(writer, "error in getting balance", http.StatusInternalServerError)
		return

	}

	balance := entities.Balance{}
	balance.Current = current
	balance.WithDrawn = withdrawn
	answer, err := json.Marshal(balance)
	if err != nil {
		http.Error(writer, "error in getting balance", http.StatusInternalServerError)
		return

	}
	writer.Write(answer)

}

func WithDraw(writer http.ResponseWriter, request *http.Request) {
	id, ok := request.Context().Value("userid").(int)

	if !ok {
		http.Error(writer, "error in withdraw", http.StatusInternalServerError)
		return
	}
	withdraw := entities.WithDraw{}

	err := json.NewDecoder(request.Body).Decode(&withdraw)

	if err != nil {
		http.Error(writer, "error in withdraw", http.StatusInternalServerError)
		return
	}
	res, err := databases.SetWithDraw(request, id, withdraw.Order, withdraw.Sum)

	writer.WriteHeader(res)
}

func WithDrawals(writer http.ResponseWriter, request *http.Request) {
	id, ok := request.Context().Value("userid").(int)

	if !ok {
		http.Error(writer, "invalid number of order", http.StatusInternalServerError)
		return
	}

	withdrawals, err := databases.GetWithDrawals(id)
	if err != nil {
		http.Error(writer, "error in getting orders", http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		http.Error(writer, "not data for answer", http.StatusNoContent)
		return
	}
	answer, err := json.Marshal(withdrawals)

	writer.Write(answer)
}
