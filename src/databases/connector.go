package databases

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"gophermart/src/config"
	"gophermart/src/entities"
	"log"
	"net/http"
	"sync"
)

type Service struct {
	db *sql.DB
}

var lock = &sync.Mutex{}
var dbService *Service

func GetInstance() (*Service, error) {
	if dbService == nil {
		dbService = &Service{}
		lock.Lock()

		sqlInfo := fmt.Sprintf("host=%v port=%v user=%v "+
			"password=%v dbname=%v sslmode=disable",
			config.GetHost(), config.GetPort(), config.GetUser(), config.GetPassword(), config.GetDbName())
		db, err := sql.Open(config.GetDriverName(), sqlInfo)
		if err != nil {
			log.Fatal(err)
		}
		defer lock.Unlock()
		db.SetConnMaxLifetime(0)
		db.SetMaxIdleConns(50)
		db.SetMaxOpenConns(50)
		dbService.db = db

	}
	return dbService, nil
}

func AddUser(name, pwd string) (int64, error) {
	var r sql.Result

	stmt, err := dbService.db.Prepare("INSERT INTO users(name, pwd) VALUES($1, $2) on conflict (name) do nothing;")

	if err != nil {
		log.Printf(err.Error())
		return 0, err
	}
	defer stmt.Close()

	r, err = stmt.Exec(name, pwd)
	if err != nil {
		log.Printf("Error in insert operation, because %v", err)
		return 0, err
	}
	num, err := r.RowsAffected()

	if err != nil {
		log.Printf("Error in getting rowsaffected")
		return 0, err
	}

	if num == 0 {
		log.Printf("User with name: %s extist", name)
		return 0, nil
	}
	return num, nil
}

func GetUserId(name string) (int, error) {

	var id = 0
	err := dbService.db.QueryRow("select id from users where name = $1 limit 1;", name).Scan(&id)
	if err != nil {
		log.Printf(err.Error())
		return 0, err
	}
	return id, nil
}

func GetPassword(name string) (string, error) {
	var pwd = ""
	err := dbService.db.QueryRow("select pwd from users where name = $1 limit 1;", name).Scan(&pwd)
	if err != nil {
		log.Printf(err.Error())
		return "", err
	}
	return pwd, nil

}

func AddOrder(number int, userid int) (int, error) {
	id := 0
	rows, err := dbService.db.Query("select userid from orders where id = $1;", number)

	if err != nil {
		return 500, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return 500, err
		}
		if id == userid {
			return 200, nil
		} else if id != 0 {
			return 409, nil
		}
	}
	if rows.Err() != nil {
		return 500, err
	}
	stmt, err := dbService.db.Prepare("INSERT INTO orders(id, userid, status, accrual, uploaded_at) VALUES($1, $2, '', 0, CURRENT_TIMESTAMP);")
	if err != nil {
		return 500, errors.New("error when adding an order")
	}

	defer stmt.Close()

	_, err = stmt.Exec(number, userid)

	if err != nil {
		return 500, errors.New("error when adding an order")
	}

	return 202, nil
}

func GetOrders(userid int) ([]entities.Order, error) {
	orders := make([]entities.Order, 0)

	rows, err := dbService.db.Query("select id, status, accrual, uploaded_at from orders where userid = $1 and status in ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED', '');", userid)
	if err != nil {
		return orders, errors.New("error when getting orders")
	}
	defer rows.Close()

	for rows.Next() {
		order := entities.Order{}
		err = rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if rows.Err() != nil {
		return orders, err
	}
	return orders, nil

}

func GetNotFinalizedOrders() ([]entities.Order, error) {
	orders := make([]entities.Order, 0)

	rows, err := dbService.db.Query("select id, status, accrual, uploaded_at from orders where status in ('NEW', 'PROCESSING', '', 'REGISTERED') limit 10 for update;")
	if err != nil {
		return orders, errors.New("error when getting orders")
	}
	defer rows.Close()

	for rows.Next() {
		order := entities.Order{}
		err = rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if rows.Err() != nil {
		return orders, err
	}
	return orders, nil

}

func GetOrder(orderId int) (entities.Order, error) {
	order := entities.Order{}
	rows, err := dbService.db.Query("select id, status, accrual from orders where id = $1;", orderId)

	if err != nil {
		return order, errors.New("error when getting an order")
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&order.Number, &order.Status, &order.Accrual)
		if err != nil {
			return order, err
		}
	}
	if rows.Err() != nil {
		return order, err
	}
	return order, nil
}

func GetBalance(userid int) (int, int, error) {
	var balance = 0
	var withdrawn = 0

	rows, err := dbService.db.Query(""+
		"select sum(accrual) current_balance, sum(case when accrual<0 then -accrual end) withdrawn from orders where userid = $1 and accrual <> 0 and status ='PROCESSED';", userid)

	if err != nil {
		return 0, 0, errors.New("error in getting balance")
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&balance, &withdrawn)
		if err != nil {
			return 0, 0, err
		}
	}
	if rows.Err() != nil {
		return 0, 0, err
	}
	return balance, withdrawn, nil
}

func SetWithDraw(r *http.Request, userid, order, sum int) (int, error) {
	balance := 0
	sumbalance := 0

	tx, _ := dbService.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	defer tx.Commit()

	rows, err := dbService.db.Query("select accrual from  orders where userid = $1 and status ='PROCESSED' for update;", userid)
	if err != nil {
		tx.Rollback()
		return 500, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&balance)
		if err != nil {
			tx.Rollback()
			return 500, err
		}
		sumbalance += balance
	}

	if rows.Err() != nil {
		return 500, err
	}

	if sumbalance < sum {
		return 402, errors.New("not enough balance")
	}

	stmt, err := dbService.db.Prepare("INSERT INTO orders(id, userid, status, accrual, uploaded_at) VALUES($1, $2, 'NEW', $3, CURRENT_TIMESTAMP);")
	if err != nil {
		tx.Rollback()
		return 500, errors.New("error when adding an order")
	}

	defer stmt.Close()

	_, err = stmt.Exec(order, userid, -sum)

	if err != nil {
		tx.Rollback()
		return 500, errors.New("error when adding an order")
	}

	return 200, nil

}

func GetWithDrawals(userid int) ([]entities.WithDraw, error) {
	withdrawals := make([]entities.WithDraw, 0)

	rows, err := dbService.db.Query("select id as order, accrual as sum, uploaded_at as processed_at from orders where userid = $1 and accrual<0;", userid)

	if err != nil {
		return nil, errors.New("error when getting withdrawals")
	}
	defer rows.Close()

	for rows.Next() {
		withdraw := entities.WithDraw{}
		err = rows.Scan(&withdraw.Order, &withdraw.Sum, &withdraw.UploadedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, withdraw)
	}
	if rows.Err() != nil {
		return nil, err
	}
	return withdrawals, nil
}

func SetOrderStatus(order entities.Order) error {
	stmt, err := dbService.db.Prepare("UPDATE orders SET status = $1, accrual = $2 WHERE id = $3 and accrual=0;")

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(order.Status, order.Accrual, order.Number)

	if err != nil {
		return err
	}
	return nil
}
