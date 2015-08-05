package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
)

func init() {
	sqlList["mysql"] = mysqlOpen
}

func (c *Credential) ToMySQL() string {
	//username:password@protocol(address)/dbname
	return c.Username + ":" + c.Password + "@tcp(" + c.Host + ")/" + c.Database
}

func mysqlOpen(c *Credential) (*sqlx.DB, *sqlx.Rows) {
	db, err := sqlx.Open("mysql", c.ToMySQL())
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Queryx("SELECT * FROM " + c.Table)
	if err != nil {
		log.Fatal(err)
	}
	return db, rows
}
