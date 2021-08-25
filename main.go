package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	now := time.Now()
	yyyy, mm, dd := now.Date()
	tomorrow := time.Date(yyyy, mm, dd+1, 0, 0, 0, 0, now.Location())
	s := tomorrow.Format("2006-01-02")

	url := "https://dashboard.elering.ee/api/nps/price?start=" + s + "T00%3A00%3A00.000Z&end=" + s + "T23%3A59%3A59.999Z"

	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		log.Fatalln(err)
	}

	maps := raw["data"].(map[string]interface{})

	prices := maps["ee"].([]interface{})

	os.Remove("sqlite-database.db")

	log.Println("Creating sqlite-database.db...")
	file, err := os.Create("sqlite-database.db")
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	log.Println("sqlite-database.db created")

	sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite-database.db")
	defer sqliteDatabase.Close()
	createTable(sqliteDatabase)

	for i, s := range prices {
		insert := s.(map[string]interface{})

		price := fmt.Sprintf("%v", insert["price"])
		hour := fmt.Sprintf("%v", i)

		insertPrice(sqliteDatabase, hour, price)
	}

}

func createTable(db *sql.DB) {
	createPriceTableSQL := `CREATE TABLE prices (
		"hour" TEXT,
		"price" TEXT
	  );`

	statement, err := db.Prepare(createPriceTableSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec()
}

func insertPrice(db *sql.DB, hour string, price string) {
	insertPriceSQL := `INSERT INTO prices(hour, price) VALUES (?, ?)`
	statement, err := db.Prepare(insertPriceSQL)
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(hour, price)
	if err != nil {
		log.Fatalln(err.Error())
	}
}
