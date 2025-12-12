package main

import (
	"github.com/Mahaveer86619/bookture/server/pkg/config"
	"github.com/Mahaveer86619/bookture/server/pkg/db"
)

func main() {
	config.LoadConfig()

	db.InitBookture(true)

	booktureDB := db.GetBooktureDB()
	booktureDB.MigrateTables()
}
