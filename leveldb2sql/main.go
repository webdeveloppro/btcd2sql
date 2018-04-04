package main

import (
	"log"
	"os"
	"strconv"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/jackc/pgx"
	"github.com/webdeveloppro/btcd2sql"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/database"
	_ "github.com/btcsuite/btcd/database/ffldb"
	"github.com/btcsuite/btcd/wire"

	_ "github.com/lib/pq"
)

var chainParams = &chaincfg.MainNetParams

func main() {
	var startBlock int32
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USERNAME")
	dbpassword := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	start := os.Getenv("START_BLOCK")

	if host == "" {
		log.Print("Empty host string, setup DB_HOST env")
		host = "localhost"
	}

	if user == "" {
		log.Fatal("Empty user string, setup DB_USER env")
		return
	}

	if dbname == "" {
		log.Fatal("Empty dbname string, setup DB_DBNAME env")
		return
	}

	if start == "" {
		startBlock = 0
	} else {
		i, err := strconv.ParseInt(start, 10, 32)
		if err != nil {
			log.Fatalf("Cannot parse %s to int", start)
		}
		startBlock = int32(i)
	}

	connPoolConfig := pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     host,
			User:     user,
			Password: dbpassword,
			Database: dbname,
		},
		MaxConnections: 100,
	}

	pg, err := pgx.NewConnPool(connPoolConfig)
	if err != nil {
		log.Fatalf("Unable to create connection pool %v", err)
	}
	defer pg.Close()

	db, err := database.Open("ffldb", os.Getenv("BTCD_DATADIR"), wire.MainNet)

	if err != nil {
		log.Fatalf("fatal error happenned: %v", err)
		// Handle error
	}
	defer db.Close()

	cfg := blockchain.Config{
		DB:          db,
		ChainParams: &chaincfg.MainNetParams,
		TimeSource:  blockchain.NewMedianTime(),
	}

	bc, err := blockchain.New(&cfg)
	if err != nil {
		log.Fatalf("cannot create blockchain %v", err)
	}

	db2sql.Parse(bc, pg, startBlock)
}
