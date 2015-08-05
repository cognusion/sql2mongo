package main

import (
	"gopkg.in/mgo.v2"
	"log"
)

func (c *Credential) ToMongo() string {
	//mongodb://myuser:mypass@localhost:40001,otherhost:40001/mydb
	return "mongodb://" + c.Username + ":" + c.Password + "@" + c.Host + "/" + c.Database
}

func mongoOpen(c *Credential) (*mgo.Session, *mgo.Collection) {
	session, err := mgo.Dial(c.ToMongo())
	if err != nil {
		log.Fatal(err)
	}
	//defer session.Close()

	coll := session.DB(c.Database).C(c.Table)

	return session, coll
}

func mongoWriteRow(coll *mgo.Collection, op string, row *map[string]interface{}) bool {
	var err error

	// Sanity
	if op == "" {
		op = "upsert"
	}

	// Some stupidity to deref the row and grab the id
	// since neither row["_id"] nor *row["_id"] are valid
	// TIMTOWTDI
	xr := *row
	rid := xr["_id"]

	// Switch on the writeoperation
	if op == "upsert" {
		_, err = coll.UpsertId(rid, row)
	} else if op == "insert" {
		err = coll.Insert(row)
	} else if op == "update" {
		err = coll.UpdateId(rid, row)
	}

	if err != nil {
		log.Println(err)
		return false
	}
	return true
}
