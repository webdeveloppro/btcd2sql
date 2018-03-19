package db2sql

import (
	"fmt"
	"log"

	"github.com/btcsuite/btcd/blockchain"

	"database/sql"

	"github.com/pkg/errors"
)

// Blockchain2SQL layer that connects blockchain and sql data
type Blockchain2SQL struct {
	Blockchain *blockchain.BlockChain
	Block2sql  []Block2SQL
	pg         *sql.DB
}

// New returns a BlockChain instance using the provided configuration details.
func New(cfg *blockchain.Config, pg *sql.DB) (Blockchain2SQL, error) {

	bc, err := blockchain.New(cfg)
	if err != nil {
		return Blockchain2SQL{}, errors.Wrap(err, "cannot create new blockchain")
	}

	b2q := Blockchain2SQL{
		Blockchain: bc,
		pg:         pg,
	}
	return b2q, nil
}

// Parse will transfer leveldb blockchain data to SQL
func (BC2SQL *Blockchain2SQL) Parse() {

	var n int32
	BC2SQL.Block2sql = make([]Block2SQL, 0, 0)

	b := BC2SQL.Blockchain.BestSnapshot()
	fmt.Printf("best height: %d", b.Height)

	// beststate := BC2SQL.Blockchain.BestSnapshot()
	for n = 154012; n < 1068641; n++ {
		blk, err := BC2SQL.Blockchain.BlockByHeight(n)

		if err != nil {
			// return errors.Wrapf(err, "Cannot get block by height %d", n)
			log.Fatalf("Cannot get block by height %d", n)
		}

		b2sql := NewBlock(blk, BC2SQL.pg)

		if err := b2sql.Insert(); err != nil {
			log.Fatal(err)
		}

		if err := b2sql.InsertTransactions(); err != nil {
			log.Fatal(err)
		}
	}
}
