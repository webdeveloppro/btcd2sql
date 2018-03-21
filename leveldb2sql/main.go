package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/vladyslav2/db2sql"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/database"
	_ "github.com/btcsuite/btcd/database/ffldb"
	"github.com/btcsuite/btcd/wire"

	_ "github.com/lib/pq"
)

var chainParams = &chaincfg.MainNetParams

func main() {
	dbinfo := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_NAME"))
	db, err := database.Open("ffldb", os.Getenv("BTCD_DATADIR"), wire.MainNet)

	if err != nil {
		log.Fatalf("fatal error happenned: %v", err)
		// Handle error
	}
	defer db.Close()

	pg, err := sql.Open("postgres", dbinfo)
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer pg.Close()
	pg.SetMaxOpenConns(80)

	cfg := blockchain.Config{
		DB:          db,
		ChainParams: &chaincfg.MainNetParams,
		TimeSource:  blockchain.NewMedianTime(),
	}

	bc, err := blockchain.New(&cfg)
	if err != nil {
		log.Fatalf("cannot create blockchain %v", err)
	}

	db2sql.Parse(bc, pg)
}
