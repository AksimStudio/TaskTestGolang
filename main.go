package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "06011501"
	dbname   = "task"
)

func main() {
	flag.Parse()
	orderNumbersStr := flag.Arg(0)

	dbinfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	orderNumbers := strings.Split(orderNumbersStr, ",")
	fmt.Printf("=+=+=+=\nСтраница сборки заказов %s\n\n", orderNumbersStr)

	rows, err := db.Query(`
		SELECT s.id, s.name AS shelf_name, p.name AS product_name, p.id AS product_id, oi.order_id, oi.quantity
		FROM shelf_items si
		INNER JOIN shelves s ON si.shelf_id = s.id
		INNER JOIN order_items oi ON si.order_items_id = oi.id
		INNER JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id IN (` + strings.Join(orderNumbers, ",") + `)
		ORDER BY s.name, s.id
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	//Карта для отслеживания уникальных стеллажей и их ID
	visitedShelves := make(map[string][]int)

	for rows.Next() {
		var shelfID, productID, orderID, quantity int
		var shelfName, productName string

		if err := rows.Scan(&shelfID, &shelfName, &productName, &productID, &orderID, &quantity); err != nil {
			log.Fatal(err)
		}

		//Проверяем,есть ли уже такой стеллаж в карте
		existingIDs, ok := visitedShelves[shelfName]
		if ok {
			//текущий ID не равен предыдущему, считаем его дополнительным стеллажом
			if !contains(existingIDs, shelfID) {
				visitedShelves[shelfName] = append(existingIDs, shelfID)
				fmt.Printf("доп стеллаж: %s (id=%d)===\n", shelfName, shelfID)
			}
		} else {
			//выводим стеллаж как обычно и добавляем его ID в карту
			fmt.Printf("===Стеллаж %s", shelfName)
			if shelfID != 0 {
				fmt.Printf(" (id=%d)", shelfID)
				visitedShelves[shelfName] = []int{shelfID}
			}
			fmt.Println("===")
		}

		fmt.Printf("%s (id=%d)\n", productName, productID)
		fmt.Printf("заказ %d, %d шт\n\n", orderID, quantity)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}

func contains(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
