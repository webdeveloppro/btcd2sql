package db2sql

import (
	"database/sql"
	"fmt"

	"github.com/btcsuite/btcutil"
	"github.com/pkg/errors"
)

type Block interface {
	Insert() error
	InsertTransactions() error
}

// Block2SQL structure for holding Block information
type Block2SQL struct {
	ID           int
	Block        *btcutil.Block
	Transactions map[string]int
	pg           *sql.DB
}

// NewBlock Block2SQL struct constructor
func NewBlock(block *btcutil.Block, pg *sql.DB) *Block2SQL {
	blk := Block2SQL{
		Block: block,
		pg:    pg,
	}
	blk.Transactions = make(map[string]int)
	return &blk
}

// Insert Will transfer blockchain block to SQL block table. Will not touch blockchain transactions data
func (B2SQL *Block2SQL) Insert() error {

	if B2SQL.Block.Hash().String() == "" {
		return fmt.Errorf("Btcd2sql have empty hash, nothing to insert")
	}

	if err := B2SQL.pg.QueryRow(`
		INSERT INTO block
			(bits, height, nonce, version, hash_prev_block, hash_merkle_root, created_at, hash)
		VALUES
			(
				$1,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8
			)
			RETURNING id`,
		B2SQL.Block.MsgBlock().Header.Bits,
		B2SQL.Block.Height(),
		B2SQL.Block.MsgBlock().Header.Nonce,
		B2SQL.Block.MsgBlock().Header.Version,
		B2SQL.Block.MsgBlock().Header.PrevBlock.String(),
		B2SQL.Block.MsgBlock().Header.MerkleRoot.String(),
		B2SQL.Block.MsgBlock().Header.Timestamp,
		B2SQL.Block.Hash().String(),
	).Scan(&B2SQL.ID); err != nil {
		err = errors.Wrap(err, "insert block failed")
		return err
	}
	return nil
}

// InsertTransactions looping block transactions and invoke parsing of in and out data
func (B2SQL *Block2SQL) InsertTransactions() error {
	transactions := B2SQL.Block.Transactions()

	for _, tran := range transactions {
		ins := tran.MsgTx().TxIn
		outs := tran.MsgTx().TxOut
		tranHash := tran.Hash().String()
		tranID := 0

		sql := `
		INSERT INTO transaction
			(hash, block_id, has_witness)
		VALUES
			(
				$1,
				$2,
				$3
			)
			RETURNING id`

		if err := B2SQL.pg.QueryRow(sql,
			tranHash,
			B2SQL.ID,
			tran.HasWitness(),
		).Scan(&tranID); err != nil {
			if err.Error() == "pq: duplicate key value violates unique constraint \"transaction_hash_key\"" {
				// https://bitcoin.stackexchange.com/questions/71918/why-does-transaction-d5d27987d2a3dfc724e359870c6644b40e497bdc0589a033220fe15429d
				continue
			}

			fmt.Println(sql, tranHash, B2SQL.ID, tran.HasWitness())
			return errors.Wrap(err, "insert transaction failed")
		}
		B2SQL.Transactions[tranHash] = tranID

		// fmt.Printf("Transaction: %s\n Input:", tranHash)

		for _, txIn := range ins {
			err := B2SQL.insertTxIN(txIn, B2SQL.Transactions[tranHash])
			if err != nil {
				return err
			}
		}

		for _, txOut := range outs {
			err := B2SQL.insertTxOUT(txOut, B2SQL.Transactions[tranHash])
			if err != nil {
				return err
			}
		}
	}
	return nil
}
