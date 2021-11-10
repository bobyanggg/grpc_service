package sql

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gitlab.com/leopardx602/grpc_service/model"
)

type Product struct {
	Word       string
	ProductID  string
	Name       string
	Price      int
	ImageURL   string
	ProductURL string
}

var Conn *sql.DB

func init() {
	// Read the sql config.
	sqlConfig, err := model.OpenJson("../config/sql.json")
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the sql.
	connInfo := fmt.Sprintf(
		"%s:%s@tcp(%s:%v)/%s?parseTime=%v",
		sqlConfig["username"],
		sqlConfig["password"],
		sqlConfig["addr"],
		sqlConfig["port"],
		sqlConfig["database"],
		sqlConfig["parseTime"])
	Conn, err = sql.Open("mysql", connInfo)
	if err != nil {
		log.Fatal(err)
	}

	// Clear old data
	go func() {
		for {
			if err := Delete(); err != nil {
				log.Println(err)
			}
			now := time.Now()
			next := now.Add(time.Hour * 24)
			next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
			t := time.NewTimer(next.Sub(now))
			<-t.C
		}
	}()
}

// Create the table products and keyword.
func Create() error {
	sql := `CREATE TABlE IF NOT EXISTS products(
		productID VARCHAR(63) NOT NULL,
		name VARCHAR(255) NOT NULL DEFAULT "",
		price INT NOT NULL DEFAULT 0,
		imageURL VARCHAR(255) NOT NULL DEFAULT "",
		productURL VARCHAR(255) NOT NULL DEFAULT "",
		updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (productID)
		);`
	_, err := Conn.Query(sql)
	if err != nil {
		return err
	}

	sql = `CREATE TABlE IF NOT EXISTS keyword(
			word VARCHAR(255) NOT NULL,
			productID VARCHAR(63) NOT NULL,
			updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (productID) REFERENCES products (productID)
		);`
	_, err = Conn.Query(sql)
	if err != nil {
		return err
	}
	return nil
}

func Insert(product Product) error {
	_, err := Conn.Exec("INSERT INTO products (productID, name, price, imageURL, productURL) VALUES (?, ?, ?, ?, ?)", product.ProductID, product.Name, product.Price, product.ImageURL, product.ProductURL)
	if err != nil {
		if strings.Contains(err.Error(), "1062") {
			fmt.Println("already exists")
		} else {
			return err
		}
	}

	_, err = Conn.Exec("INSERT INTO keyword (word, productID) VALUES (?, ?)", product.Word, product.ProductID)
	if err != nil {
		return err
	}
	return nil
}

func Select(keyword string) ([]Product, error) {
	sql := fmt.Sprintf("SELECT name, price, imageURL, productURL FROM products WHERE productID IN (SELECT productID FROM keyword WHERE word = '%s')", keyword)
	res, err := Conn.Query(sql)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var products []Product
	for res.Next() {
		var product Product
		if err := res.Scan(&product.Name, &product.Price, &product.ImageURL, &product.ProductURL); err != nil {
			return nil, err
		}
		//fmt.Println(product)
		products = append(products, product)
	}
	//fmt.Println(products)
	return products, nil
}

func Delete() error {
	_, err := Conn.Exec("DELETE FROM keyword WHERE updated_at < (NOW() - INTERVAL 1 DAY)")
	if err != nil {
		return err
	}
	_, err = Conn.Exec("DELETE FROM products WHERE updated_at < (NOW() - INTERVAL 1 DAY)")
	if err != nil {
		return err
	}
	return nil
}
