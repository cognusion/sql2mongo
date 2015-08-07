# sql2mongo
A simple engine to take a SQL (MySQL or PostgreSQL, currently, but trivial to add more (see "Adding new SQL sources" below)) table, and dump it into a MongoDB collection.

I'm under no delusion that others have [re]invented this wheel previously. I had specific needs and didn't find anything that did what I wanted without lots of overhead I didn't want.

## Basics

### Build

There are zipped binaries for OSX (64bit) and Linux (i386 and x86_64) in the _binaries/_ folder

Assuming you have Go and Git:
```bash
git clone https://github.com/cognusion/sql2mongo.git

go get github.com/jmoiron/sqlx
go get gopkg.in/mgo.v2
go get github.com/lib/pq  # PostgreSQL driver
go get github.com/go-sql-driver/mysql  # MySQL driver

cd sql2mongo
go build
```

If you're not going to use a particular SQL server type, you may omit the driver from the _go get_ list above, **and** delete its _whatever.go_ file from the sql2mongo folder before running ```go build```

### Usage

```bash
./sql2mongo --help
Usage of ./sql2mongo:
  -config="my2mo.json": Config file to read
  -configtest=false: Load and parse configs, and exit
  -debug=false: Enable Debug output
  -dontconvertbytes=false: We automatically convert byte arrays to strings. Set this to prevent it.
  -dontconvertid=false: We automatically convert any SQL column 'id' to mongo element '_id'. Set this to prevent it.
  -job="": Name of the job to run
  -list=false: List the jobs available
```

## Config

A config file defining jobs is necessary to accomplish work. Below is a simple one-job config that would be executed like ```./sql2mongo --job dupethattable``` 

```json
{
	"jobs": [
		{
			"name": "dupethattable",
			"description": "Copy the that table that collections",
			"writeoperation": "upsert",
			"enabled": true,
			"sqltype": "mysql",
			"sqlcolumns": "id, afield, `key`, bfield",
			"sql": {
				"host": "mysqlserver:3306",
				"username": "myuser",
				"password": "mypass",
				"database": "mydb",
				"table": "mytable"
			},
			"mongo": {
				"host": "mongo1:27017,mongo2:27017",
				"username": "mouser",
				"password": "mopass",
				"database": "modb",
				"table": "mocollection"
			}
		}	
	]
}
```

###Job Definition

Jobs have metadata:
* Name - Whatever you want to call this job. Needs uniqueness
* Description - Some words about what this job is doing
* Enabled - Boolean true/false as to whether this job is available for running
* WriteOperation - Which Mongo operation should be used: 
  * insert (will fail if exists)
  * update (will fail if doesn't exist)
  * upsert (will always succeed)
* SqlType - Which SQL type to use:
  * mysql - MySQL
  * pgsql - PostgreSQL
* SqlColumns - A list of columns, properly delimited and escaped, to be SELECTed
* Mongo - A Credential definition for the target MongoDB server
* SQL - A Credential defintion for the source SQL server

###Credential Definition

"Credentials" are JSON subdocs that are structured and passed around inside the application. They are defined separately for the SQL source server, and the Mongo destination server.
* Name - Optional name of the server. Never consulted.
* Host - Host connect string. e.g. "myserver:3306" or whatever
* Username - User to connect as
* Password - Password to authenticate with
* Database - Database to connect to
* Table - Table (or Collection) in the above Database to read from (or write to)

## Tomorrow, Tomorrow ...

What may happen, in no particular order:
* More SQL sources
* Source WHERE clause spec
* Source joins

What won't happen, in no particular order:
* Any non-SQL sources
* Any non-MongoDB destinations


## Adding new SQL sources

Any _database/sql_ -compatible DB can be added trivially, even if you aren't particularly handy with Go. I would recommend cloning mysql.go and salting that as needed. Copied below with extra comments to draw out the 5 changes needed:

```go
package main

import (
	_ "github.com/go-sql-driver/mysql" // 1. import your driver
	"github.com/jmoiron/sqlx"
	"log"
)

func init() {
	sqlList["mysql"] = mysqlOpen // 2. shove your function's address into sqlList[]. 
								 // The keyname will be the "sqltype" referenced in configs
}

func (c *Credential) ToMySQL() string {	// 3. rename and salt this function to return 
										// a valid "Open" string from a Credential
	
	//username:password@protocol(address)/dbname
	return c.Username + ":" + c.Password + "@tcp(" + c.Host + ")/" + c.Database
}

func mysqlOpen(c *Credential, columns string) (*sqlx.DB, *sqlx.Rows) { // 4. rename
	db, err := sqlx.Open("mysql", c.ToMySQL())	// 5. rename the driver, 
												//and the c.To*() to #3 above
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
```

