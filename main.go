package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/alecthomas/colour"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hgfischer/mysqlsuperdump/dumper"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("klepto", "Steals data from production to put on staging")

	steal = app.Command("steal", "Steal a live database")
	stealingDSN = steal.Flag("inputdsn", "DSN for the input database").Default("root:root@localhost/example").String()
	swagDSN = steal.Flag("outputdsn", "DSN for the output database (or just 'STDOUT' for a dump)").Default("root:root@localhost/example").String()
)

func ensureConnectionIsGood(db *sql.DB) error {
	tables := make([]string, 0)
	var rows *sql.Rows
	var err = error(nil)
	if rows, err = db.Query("SHOW FULL TABLES"); err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var tableName, tableType string
		if err := rows.Scan(&tableName, &tableType); err != nil {
			return err
		}
		if tableType == "BASE TABLE" {
			tables = append(tables, tableName)
		}
	}
	return nil
}

func main() {
	command, err := app.Parse(os.Args[1:])

	if err != nil {
		colour.Stderr.Printf("^1%s^R, try --help\n", err)
		os.Exit(2)
	}

	db, err := sql.Open("mysql", *stealingDSN)
	logger := log.New(os.Stdout, "klepto: ", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)
	if err != nil {
		log.Fatalf("MySQL connection failed: %s \n", err)
	}

	err = ensureConnectionIsGood(db)
	if err != nil {
		log.Fatalf("Error in MySQL connection: %s \n", err)
	}

	switch command {
	case steal.FullCommand():
		d := dumper.NewMySQLDumper(db, logger)
		// TODO: Define masks from config file
		if *swagDSN == "STDOUT" {
			d.Dump(os.Stdout) // TODO: Define out as another mysql db (config from CLI)
		}
		break
	}
}