package db2sql

import (
	"fmt"
	"log"

	"github.com/jackc/pgx"
	"github.com/vladyslav2/bitcoin2sql/pkg/address"
	"github.com/vladyslav2/bitcoin2sql/pkg/block"

	"github.com/btcsuite/btcd/blockchain"
)

// AddrStorage address storage
var AddrStorage map[string]*address.Address

// Parse will transfer leveldb blockchain data to SQL
func Parse(bc *blockchain.BlockChain, pg *pgx.ConnPool) {

	rows, err := pg.Query("SELECT id, hash, ballance, income, outcome from address")
	if err != nil {
		log.Fatalf("Cannot read data from addresses, %v", err)
	}

	AddrStorage = make(map[string]*address.Address, 1000000)

	AddPgStorage := address.NewStorage(pg)

	for rows.Next() {
		a := address.New(&AddPgStorage)
		if err := rows.Scan(
			&a.ID,
			&a.Hash,
			&a.Ballance,
			&a.Income,
			&a.Outcome,
		); err != nil {
			log.Fatalf("cannot read address query, %v", err)
		}
		AddrStorage[a.Hash] = a
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
