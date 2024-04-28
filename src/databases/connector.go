package databases

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"gophermart/src/config"
	"gophermart/src/entities"
	"log"
	"sync"
)

type Service struct {
	db *sql.DB
}

var lock = &sync.Mutex{}
var dbService *Service

func GetInstance() *Service {
	if dbService == nil {
		dbService = &Service{}
		lock.Lock()
		defer lock.Unlock()
		sqlInfo := fmt.Sprintf("host=%v port=%v user=%v "+
			"password=%v dbname=%v sslmode=disable",
			config.GetHost(), config.GetPort(), config.GetUser(), config.GetPassword(), config.GetDbName())
		db, err := sql.Open(config.GetDriverName(), sqlInfo)
		if err != nil {
			log.Fatal(err)
		}
		db.SetConnMaxLifetime(0)
		db.SetMaxIdleConns(50)
		db.SetMaxOpenConns(50)
		dbService.db = db

	}
	return dbService
}

func AddUser(name, pwd string) (int64, error) {

	stmt, err := dbService.db.Prepare("INSERT INTO users(name, pwd) VALUES($1, $2) on conflict (name) do nothing;")
	defer stmt.Close()

	if err != nil {
		log.Printf(err.Error())
		return 0, err
	}
	var r sql.Result
	r, err = stmt.Exec(name, pwd)
	if err != nil {
		log.Printf("Error in insert operation, because %v", err)
		return 0, err
	}
	num, err := r.RowsAffected()

	if err != nil {
		log.Printf("Error in getting rowsaffected")
	}

	if num == 0 {
		log.Printf("User with name: %s extist", name)
		return 0, nil
	}
	return num, nil
}

func CheckUser(name, pwd string) (int, error) {

	var id = 0
	err := dbService.db.QueryRow("select id from users where name = $1 and pwd = $2 limit 1;", name, pwd).Scan(&id)
	if err != nil {
		log.Printf(err.Error())
		return 0, err

	}
	return id, nil
}

func AddOrder(number int, userid int) (int, error) {
	id := 0
	rows, err := dbService.db.Query("select userid from orders where id = $1;", number)
	defer rows.Close()
	if err != nil {
		return 500, err
	}
	if rows.Next() {
		rows.Scan(&id)
		if id == userid {
			return 200, nil
		} else if id != 0 {
			return 409, nil
		}
	}
	stmt, err := dbService.db.Prepare("INSERT INTO orders(id, userid, status, accrual, uploaded_at) VALUES($1, $2, 'NEW', 0, CURRENT_TIMESTAMP);")
	defer stmt.Close()

	if err != nil {
		return 500, errors.New("error when adding an order")
	}

	_, err = stmt.Exec(number, userid)

	if err != nil {
		return 500, errors.New("error when adding an order")
	}

	return 202, nil
}

func GetOrders(userid int) ([]entities.Order, error) {

	rows, err := dbService.db.Query("select id, status, accrual, uploaded_at from orders where userid = $1;", userid)
	defer rows.Close()
	orders := make([]entities.Order, 0)
	if err != nil {
		return orders, errors.New("error when getting orders")
	}
	for rows.Next() {
		order := entities.Order{}
		rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		orders = append(orders, order)
	}
	return orders, nil

}

func GetOrder(orderId int) (entities.Order, error) {
	rows, err := dbService.db.Query("select id, status, accrual from orders where id = $1;", orderId)
	defer rows.Close()
	order := entities.Order{}
	if err != nil {
		return order, errors.New("error when getting an order")
	}
	if rows.Next() {
		rows.Scan(&order.Number, &order.Status, &order.Accrual)
	}
	return order, nil
}

func GetBalance(userid int) (int, int, error) {
	rows, err := dbService.db.Query(""+
		"select sum(accrual) current_balance, sum(case when accrual<0 then -accrual end) withdrawn from orders where userid = $1 and accrual <> 0;", userid)
	defer rows.Close()

	var balance = 0
	var withdrawn = 0
	if err != nil {
		return 0, 0, errors.New("error in getting balance")
	}

	if rows.Next() {
		rows.Scan(&balance, &withdrawn)
	}
	return balance, withdrawn, nil
}

func SetWithDraw(userid, order, sum int) (int, error) {
	balance := 0
	rows, err := dbService.db.Query("select sum(accrual) from  orders where userid = $1;", userid)
	defer rows.Close()
	if err != nil {
		return 500, err
	}
	if rows.Next() {
		rows.Scan(&balance)
	}
	if balance < sum {
		return 402, errors.New("not enough balance")
	}

	stmt, err := dbService.db.Prepare("INSERT INTO orders(id, userid, status, accrual, uploaded_at) VALUES($1, $2, 'NEW', $3, CURRENT_TIMESTAMP);")
	defer stmt.Close()

	if err != nil {
		return 500, errors.New("error when adding an order")
	}

	_, err = stmt.Exec(order, userid, -sum)

	if err != nil {
		return 500, errors.New("error when adding an order")
	}

	return 200, nil

}

func GetWithDrawals(userid int) ([]entities.WithDraw, error) {

	rows, err := dbService.db.Query("select id as order, accrual as sum, uploaded_at as processed_at from orders where userid = $1 and accrual<0;", userid)
	defer rows.Close()
	withdrawals := make([]entities.WithDraw, 0)
	if err != nil {
		return withdrawals, errors.New("error when getting withdrawals")
	}

	for rows.Next() {
		withdraw := entities.WithDraw{}
		rows.Scan(&withdraw.Order, &withdraw.Sum, &withdraw.UploadedAt)
		withdrawals = append(withdrawals, withdraw)
	}
	return withdrawals, nil
}
