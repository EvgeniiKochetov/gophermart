package src

import (
	"context"
	"github.com/gorilla/mux"
	"gophermart/src/external"
	"gophermart/src/handlers"
	"log"
	"net/http"
	"time"
)

func Run() {

	r := mux.NewRouter()

	r.HandleFunc("/api/user/register", handlers.RegisterUser)
	r.HandleFunc("/api/user/login", handlers.LoginUser)
	r.HandleFunc("/api/user/orders", handlers.Auth(handlers.Orders))

	r.HandleFunc("/api/user/balance", handlers.Auth(handlers.UserBalance))
	r.HandleFunc("/api/user/balance/withdraw", handlers.Auth(handlers.WithDraw))
	r.HandleFunc("/api/user/withdrawals", handlers.Auth(handlers.WithDrawals))

	r.HandleFunc("/api/orders/{number}", handlers.Order)

	ctx, _ := context.WithCancel(context.Background())
	go external.Run(ctx)

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
