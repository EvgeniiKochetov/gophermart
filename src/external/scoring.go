package external

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gophermart/src/config"
	"gophermart/src/databases"
	"gophermart/src/entities"
)

func poll() {

	orders, err := databases.GetNotFinalizedOrders()

	if err != nil {
		log.Print(err)
		return
	}
	if orders == nil || len(orders) == 0 {
		log.Print(err)
		return
	}

	sbBuilder := strings.Builder{}

	config.GetAccrualSystemAddress()

	for _, order := range orders {
		fmt.Println(order.Number)
		sbBuilder.WriteString(config.GetAccrualSystemAddress())
		sbBuilder.WriteString("/api/orders/")
		sbBuilder.WriteString(strconv.Itoa(order.Number))

		r, err := http.Get(sbBuilder.String())

		if err != nil {
			log.Print(err)
			sbBuilder.Reset()
			continue
		}

		switch r.StatusCode {
		case 429:
			time.Sleep(10 * time.Second)
		case 200:

			decoder := json.NewDecoder(r.Body)

			order := entities.Order{}

			err := decoder.Decode(&order)
			if err != nil {
				log.Print(err)
				continue
			}
			databases.SetOrderStatus(order)
			sbBuilder.Reset()
		default:
			continue
		}

	}
}

func Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			poll()
		}
	}
}
