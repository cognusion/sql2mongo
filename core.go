package main

/*
 * sql2mongo - A simple engine to take a SQL table, and dump it into a mongo collection
 *
 */

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// our global debug object. Defaults to discard
var debugOut *log.Logger = log.New(ioutil.Discard, "", log.Lshortfile)

// If you add a new SQL input, make sure to add to this list: sqlList[name] = yourSqlInitFunction
// see mysql.go or pgsql.go
var sqlList map[string]func(*Credential)(*sqlx.DB, *sqlx.Rows)

func init() {
	sqlList = make(map[string]func(*Credential)(*sqlx.DB, *sqlx.Rows))
}

type Config struct {
	Jobs []Job
}

func (c *Config) GetJob(name string) (*Job, bool) {
	for _, job := range c.Jobs {
		if name == job.Name && job.Enabled == true {
			return &job, true
		}
	}
	// sad face
	return nil, false
}

type Job struct {
	Name           string
	Description    string
	Condition      string
	Enabled        bool
	WriteOperation string
	SqlType        string
	Mongo          Credential
	SQL            Credential
}

type Credential struct {
	Name     string
	Host     string
	Username string
	Password string
	Database string
	Table    string
}

func main() {

	var (
		debug      bool
		confFile   string
		configTest bool
		jobName    string
		listJobs   bool
		dontByte   bool
		dont_id    bool
	)

	flag.BoolVar(&debug, "debug", false, "Enable Debug output")
	flag.StringVar(&confFile, "config", "my2mo.json", "Config file to read")
	flag.BoolVar(&configTest, "configtest", false, "Load and parse configs, and exit")
	flag.StringVar(&jobName, "job", "", "Name of the job to run")
	flag.BoolVar(&dontByte, "dontconvertbytes", false, "We automatically convert byte arrays to strings. Set this to prevent it.")
	flag.BoolVar(&dont_id, "dontconvertid", false, "We automatically convert any SQL column 'id' to mongo element '_id'. Set this to prevent it.")
	flag.BoolVar(&listJobs, "list", false, "List the jobs available")
	flag.Parse()

	if debug {
		debugOut = log.New(os.Stdout, "[DEBUG]", log.Lshortfile)
	}

	// Because of my silly logic, the config params are logically 
	// inverse from the function parameters. Dealing with that here.
	var (
		convertBytes bool = true
		convertId    bool = true
	)
	if dontByte {
		convertBytes = false
	}
	if dont_id {
		convertId = false
	}

	// Read config
	conf := loadFile(confFile)

	if configTest {
		// Just kicking the tires...
		fmt.Println("Config loaded and bootstrapped successfully...")
		os.Exit(0)
	}

	if listJobs {
		for _, job := range conf.Jobs {
			if job.Enabled == true {
				fmt.Printf("%s: %s\n", job.Name, job.Description)
			}
		}
		os.Exit(0)
	} else if jobName == "" {
		log.Fatal("--job is required")
	}

	// Grab the job, and let us get this party started
	job, ok := conf.GetJob(jobName)
	if ok == false {
		log.Fatalf("Job name %s is not valid, or disabled. Perhaps --list?\n", jobName)

	} else {

		// ensure writeoperation is one of insert, update, upsert
		if job.WriteOperation == "insert" || job.WriteOperation == "update" || job.WriteOperation == "upsert" {
		} else {
			log.Fatalf("WriteOperation of %s for job %s is invalid!\n", job.WriteOperation, job.Name)
		}

		// Mongo handle, right to the collection
		mongod, mongoc := mongoOpen(&job.Mongo)

		// SQL can vary
		var sqlc *sqlx.Rows
		var sqld *sqlx.DB

		// Test the SqlType
		found := false
		for k, f := range sqlList {
			if k == job.SqlType {
				found = true
				debugOut.Printf("Found SQL type %s\n", k)
				sqld, sqlc = f(&job.SQL)
				break
			}
		}
		if found == false {
			log.Fatalf("SQL type of %s is not currently valid. Pick one of: "+joinKeys(sqlList)+"\n", job.SqlType)
		}

		writeCount := 0
		errorCount := 0
		// Let's do this
		for sqlc.Next() {
			res, err := sqlGetRow(sqlc, convertBytes, convertId)
			// we already outputted any error in sqlGetRow
			// so we'll just skip this write, and move along
			if err == nil {
				ok := mongoWriteRow(mongoc, job.WriteOperation, res)
				if ok {
					writeCount++
				} else {
					errorCount++
				}
			} else {
				errorCount++
			}
		}

		// Clean up, explicitly
		mongod.Close()
		sqld.Close()

		totalCount := writeCount + errorCount
		log.Printf("Load completed with %d/%d rows written", writeCount, totalCount)
	}
}

// joinKeys is a simple function to spit out a comma-separated list of keys from
// specifically the sqlList map
func joinKeys(m map[string]func(*Credential)(*sqlx.DB, *sqlx.Rows)) string {
	var keys []string
	for k, _ := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}

// loadFile loads the specified file, and marshals the JSON into a Config struct
func loadFile(filePath string) Config {
	var conf Config

	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading config file '%s': %s\n", filePath, err)
	}

	err = json.Unmarshal(buf, &conf)
	if err != nil {
		log.Fatalf("Error parsing JSON in config file '%s': %s\n", filePath, err)
	}
	return conf
}
