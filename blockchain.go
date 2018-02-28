package main

import (
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

func (BC2SQL *Blockchain2SQL) init() error {
	var n int32
	BC2SQL.Block2sql = make([]Block2SQL, 0, 0)
	// beststate := BC2SQL.Blockchain.BestSnapshot()
	for n = 0; n < 56000; n++ {
		blk, err := BC2SQL.Blockchain.BlockByHeight(n)
		if err != nil {
			return errors.Wrapf(err, "Cannot create block by height %d", n)
		}
		b2sql := Block2SQL{
			Block: blk,
			pg:    BC2SQL.pg,
		}
		b2sql.Transactions = make(map[string]int)

		BC2SQL.Block2sql = append(BC2SQL.Block2sql, b2sql)
	}
	return nil
}

// Parse will transfer leveldb blockchain data to SQL
func (BC2SQL *Blockchain2SQL) Parse() error {

	err := BC2SQL.init()
	if err != nil {
		return err
	}

	for _, b2sql := range BC2SQL.Block2sql {
		err := b2sql.Insert()
		if err != nil {
			return err
		}
		err = b2sql.InsertTransactions()
		if err != nil {
			return err
		}
	}
	return nil
}
