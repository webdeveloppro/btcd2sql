package db2sql

import (
	"github.com/btcsuite/btcd/wire"
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

	fmt.Printf("Started: %d\n", B2SQL.Block.Height())
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

		// Gather all TxIn Information
		txSQL := ""
		for _, txIn := range ins {
			txSQL = fmt.Sprintf("%s,%s", txSQL, txIn2JSONB(txIn))
			if err != nil {
				return err
			}
		}

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

		for _, txOut := range outs {
			err := B2SQL.insertTxOUT(txOut, B2SQL.Transactions[tranHash])
			if err != nil {
				return err
			}
		}
	}

	fmt.Printf("Ended: %d\n", B2SQL.Block.Height())
	return nil
}

func txIn2JSONB(txIn *wire.TxIn, address_id int) string {
  
	// Get previous address hash if that exists
	if txIn.PreviousOutPoint.Hash.String() != "0000000000000000000000000000000000000000000000000000000000000000" {
		// spew.Dump(txIn)
		disbuf, err := txscript.DisasmString(txIn.SignatureScript)
		if err != nil {
			log.Fatalln(err)
		}
		disbufArr := strings.Split(disbuf, " ")
		// spew.Dump(disbuf, txIn.PreviousOutPoint)

		if len(disbufArr) > 1 {
			address, err = GetInputAddress(disbufArr[1])
			if err != nil {
				return errors.Wrap(err, "transaction: Cannot get txin address")
			}
		} else {
			address, err = FindPrevAddress(B2SQL.pg, txIn.PreviousOutPoint.Hash.String(), txIn.PreviousOutPoint.Index)

			if err != nil {
				// dirty hack for weired transactions
				// like 9969603dca74d14d29d1d5f56b94c7872551607f8c2d6837ab9715c60721b50e
				if err == sql.ErrNoRows {
					address = fmt.Sprintf("nonstandard-%s", disbuf)
				} else {
					return err
				}
			}
		}

		if err = B2SQL.pg.QueryRow(`
			SELECT id, ballance 
			FROM address 
			WHERE hash = $1`,
			address,
		).Scan(&addressID, &ballance); err != nil {
			return errors.Wrap(err, fmt.Sprintf("cant find address %s", address))
		}
	}

	return fmt.Sprintf(`{
		"address_id": %d, 
		"amount": %d,
		"prev_out": %s,
		"size": %d,
		"signature_script": %s,
		"sequence": %d,
		"witness": %s}`, 
		address_id,
		txIn.amount
	)

}