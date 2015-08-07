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

func mysqlOpen(c *Credential, columns string) (*sqlx.DB, *sqlx.Rows) {
	db, err := sqlx.Open("mysql", c.ToMySQL())
	if err != nil {
		log.Fatal(err)
	}
	
	if columns == "" {
		columns = "*"
	}
	
	rows, err := db.Queryx("SELECT " + columns + " FROM " + c.Table)
	if err != nil {
		log.Fatal(err)
	}
	return db, rows
}
