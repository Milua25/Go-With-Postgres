package middleware

import (
	"database/sql"
	"encoding/json" // package to encode and decode the json into struct and vice versa
	"fmt"
	"go-postgres-package/models"
	"log"
	"net/http" // used to access the request and response object of the api
	"os"       // used to read the environment variable
	"strconv"  // package used to covert string into int type

	"github.com/gorilla/mux"   // used to get the params from the route
	"github.com/joho/godotenv" // package used to read the .env file
	_ "github.com/lib/pq"      // postgres golang driver
)

type response struct {
	ID      int64  `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

func createConnection() *sql.DB {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(err)
	}

	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected to postgres")
	return db
}

func CreateStock(w http.ResponseWriter, req *http.Request) {
	var stock models.Stock

	err := json.NewDecoder(req.Body).Decode(&stock)
	if err != nil {
		log.Fatal("Unable to decode the request body", err)
	}

	insertID := insertStock(stock)

	res := response{
		ID:      insertID,
		Message: "stock created successfully",
	}

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Fatal("Could not encode the response", err)
	}
}

func GetStock(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Fatalf("Unable to convert id string to int. %v", err)
	}

	stock, err := getStock(int64(id))

	if err != nil {
		log.Fatalf("Unable to get stock. %v", err)
	}

	// send the response
	json.NewEncoder(w).Encode(stock)
	// err = json.NewEncoder(w).Encode(stock)
	// 	if err != nil {
	// 	log.Fatal("Could not encode the response", err)
	// }
}

func GetAllStock(w http.ResponseWriter, req *http.Request) {
	stocks, err := getAllStocks()

	if err != nil {
		log.Fatalf("Unable to get all the stocks. %v", err)
	}
	json.NewEncoder(w).Encode(stocks)

}

func UpdateStock(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Fatalf("Unable to convert id string to int. %v", err)
	}

	var stock models.Stock
	err = json.NewDecoder(req.Body).Decode(&stock)
	if err != nil {
		log.Fatalf("Unable to decode the request body . %v", err)
	}

	updateRows := updateStock(int64(id), stock)
	msg := fmt.Sprintf("Stock updated successfully. Total rows/records affected %v", updateRows)

	res := response{
		ID:      int64(id),
		Message: msg,
	}

	json.NewEncoder(w).Encode(res)
}

func DeleteStock(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Fatalf("Unable to conver string to int. %v", err)
	}

	deletedRows := deleteStock(int64(id))

	msg := fmt.Sprintf("Stock deleted successfully. Total rows/records %v", deletedRows)

	res := response{
		ID:      int64(id),
		Message: msg,
	}

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Fatal("Unable to encode")
	}
}

func insertStock(stock models.Stock) int64 {
	db := createConnection()

	defer db.Close()

	sqlStatement := `INSERT INTO stocks(name, price, company) VALUES($1, $2, $3) RETURNING stockid`

	var id int64
	err := db.QueryRow(sqlStatement, stock.Name, stock.Price, stock.Company).Scan(&id)
	if err != nil {
		log.Fatalf("Unable to excute the query %v", err)
	}
	fmt.Printf("Inserted a single record %v", id)
	return id

}

func getStock(id int64) (models.Stock, error) {
	db := createConnection()

	defer db.Close()

	var stock models.Stock
	sqlStatement := `SELECT  * FROM stocks WHERE stockid=$1`

	row := db.QueryRow(sqlStatement, id)

	err := row.Scan(&stock.StockID, &stock.Name, &stock.Price, &stock.Company)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned")
		return stock, nil
	case nil:
		return stock, nil
	default:
		log.Fatalf("Unable to scan the row %v", err)
	}

	return stock, err
}

func getAllStocks() ([]models.Stock, error) {
	db := createConnection()

	defer db.Close()

	var stocks []models.Stock
	sqlStatement := `SELECT  * FROM stocks`
	rows, err := db.Query(sqlStatement)
	if err != nil {
		log.Fatalf("unable to execute the query. %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var stock models.Stock

		err = rows.Scan(&stock.StockID, &stock.Name, &stock.Price, &stock.Company)

		if err != nil {
			log.Fatalf("Unable to scan the row %v", err)
		}
		stocks = append(stocks, stock)
	}
	return stocks, err

}

func updateStock(id int64, stock models.Stock) int64 {
	db := createConnection()

	defer db.Close()

	sqlStatement := `UPDATE stocks SET name=$2, price=$3, company=$4 WHERE stockid=$1`

	res, err := db.Exec(sqlStatement, id, stock.Name, stock.Price, stock.Company)
	if err != nil {
		log.Fatalf("unable to execute the query. %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Fatalf("Error while checking the affected rows. %v", err)
	}
	fmt.Sprintf("Total rows affected %v", rowsAffected)
	return rowsAffected
}

func deleteStock(id int64) int64 {
	db := createConnection()

	defer db.Close()

	sqlStatement := `DELETE from stocks WHERE stockid=$1`
	res, err := db.Exec(sqlStatement, id)
	if err != nil {
		log.Fatalf("unable to execute the query. %v", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Fatalf("Error while checking the affected rows. %v", err)
	}
	fmt.Sprintf("Total rows affected %v", rowsAffected)
	return rowsAffected

}
