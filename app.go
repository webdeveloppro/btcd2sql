package db2sql

import (
	"fmt"
	"log"

	"github.com/vladyslav2/bitcoin2sql/pkg/block"

	"github.com/vladyslav2/bitcoin2sql/pkg/address"

	"github.com/btcsuite/btcd/blockchain"

	"database/sql"
)

// App structure will be responsible for step by step process of migration
type App struct {
	addrStorage address.Storage
}

// New create App object
func New(storage address.Storage) *App {
	return &App{
		addrStorage: storage,
	}
}

// Parse will transfer leveldb blockchain data to SQL
func (a *App) Parse(bc *blockchain.BlockChain, pg *sql.DB) {

	rows, err := pg.Query("SELECT id, hash, ballance, income, outcome from address")
	if err != nil {
		log.Fatalf("Cannot read data from addresses, %v", err)
	}

	Addresses = make(map[string]*address.Address, 0)
	for rows.Next() {
		a := address.Address{}
		if err := rows.Scan(
			&a.ID,
			&a.Hash,
			&a.Ballance,
			&a.Income,
			&a.Outcome,
		); err != nil {
			log.Fatalf("cannot read address query, %v", err)
		}
		Addresses[a.Hash] = &a
	}

	b := bc.BestSnapshot()
	fmt.Printf("best height: %d\n", b.Height)

	// beststate := BC2SQL.Blockchain.BestSnapshot()
	var n int32
	for n = 0; n < b.Height; n++ {
		blk, err := bc.BlockByHeight(n)

		if err != nil {
			// return errors.Wrapf(err, "Cannot get block by height %d", n)
			log.Fatalf("Cannot get block by height %d", n)
		}

		sqlStorage := block.NewStorage(pg)
		sqlBlock := block.New(&sqlStorage)
		storage := NewStorage(pg)

		b2sql := NewBlock(storage, sqlBlock)
		fmt.Printf("Started block %d\n", blk.Height())

		affectedAddresses, err := b2sql.ConvertBTCD2SQL(blk)
		if err != nil {
			log.Fatalf("Cannot convert btcd data to SQL: %v", err)
		}

		if err := b2sql.Insert(affectedAddresses); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Ended block %d\n", blk.Height())
	}
}
