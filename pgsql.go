package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
)

func init() {
	sqlList["pgsql"] = pgsqlOpen
}

func (c *Credential) ToPgSQL() string {
	//postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full
	return "postgres://" + c.Username + ":" + c.Password + "@" + c.Host + "/" + c.Database
}

func pgsqlOpen(c *Credential, columns string) (*sqlx.DB, *sqlx.Rows) {
	db, err := sqlx.Open("postgres", c.ToPgSQL())
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
