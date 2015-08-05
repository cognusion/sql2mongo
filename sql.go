package main

import (
	"github.com/jmoiron/sqlx"
	"log"
)

func sqlGetRow(rows *sqlx.Rows, bytesToStrings, idTo_id bool) (*map[string]interface{}, error) {
	results := make(map[string]interface{})
	err := rows.MapScan(results)
	if err != nil {
		log.Println(err)
	} else {

		if idTo_id {
			// Convert "id" column to "_id" for mongo
			if _, ok := results["id"]; ok {
				// Mangle id, because mongo is persnickity
				results["_id"] = results["id"]
				delete(results, "id")
			}
		}

		if bytesToStrings {
			// Convert byte arrays to strings
			for k, v := range results {
				if _, ok := v.([]byte); ok {
					// Damn. Byte. Arrays. Sqlx.
					results[k] = string(v.([]byte))
				}
			}
		}
	}
	return &results, err
}
