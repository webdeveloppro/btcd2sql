package db2sql

import (
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/pkg/errors"
)

// InsertTxIN parse incoming transactions and update database txin and address tables
func (B2SQL *Block2SQL) insertTxIN(txIn *wire.TxIn, tranID int) error {
	var addressID int32
	var ballance uint64
	var address string

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
				time.Sleep(5 * time.Second)

				address, err = FindPrevAddress(B2SQL.pg, txIn.PreviousOutPoint.Hash.String(), txIn.PreviousOutPoint.Index)
				if err != nil {
					return errors.Wrap(err, "transactions: can't find previous output address")
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

		if _, err = B2SQL.pg.Exec(`
						INSERT INTO txin
							(transaction_id, address_id, amount, prev_out, sequence, size, signature_script)
						VALUES
							(
								$1,
								$2,
								$3,
								$4,
								$5,
								$6,
								$7
							) RETURNING id`,
			tranID,
			addressID,
			ballance,
			txIn.PreviousOutPoint.Hash.String(),
			txIn.Sequence,
			txIn.SerializeSize(),
			disbuf,
		); err != nil {
			return errors.Wrap(err, "transaction: insert txin failed")
		}

		if _, err = B2SQL.pg.Exec(`
						UPDATE address 
						SET ballance = 0, outcome = outcome - $1 
						WHERE id = $2`,
			ballance,
			addressID,
		); err != nil {
			return errors.Wrap(err, fmt.Sprintf("cant update address %s", disbufArr[1]))
		}
	}

	return nil
}

// InsertTxOUT parse outcoming transactions and update database txout and address tables
func (B2SQL *Block2SQL) insertTxOUT(txOut *wire.TxOut, tranID int) error {
	pkScriptHex := hex.EncodeToString(txOut.PkScript)
	_, addresses, _, err := txscript.ExtractPkScriptAddrs(
		txOut.PkScript, &chaincfg.MainNetParams)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Cannot extract pkScript %s", pkScriptHex))
	}

	addrs := ""
	for _, a := range addresses {
		addrs = fmt.Sprint(addrs, a.EncodeAddress(), ",")

		addressID := 0
		if err = B2SQL.pg.QueryRow(`
						INSERT INTO address
							(hash, income, ballance)
						VALUES
							(
								$1,
								$2,
								$3
							) ON CONFLICT (hash) DO UPDATE SET 
							  income = address.income + $2,
							  ballance = address.ballance + $3
							  RETURNING ID`,
			a.EncodeAddress(),
			txOut.Value,
			txOut.Value,
		).Scan(&addressID); err != nil {
			return errors.Wrap(err, "insert or update address failed")
		}

		if _, err = B2SQL.pg.Exec(`
						INSERT INTO txout
							(transaction_id, address_id, val, pk_script)
						VALUES
							(
								$1,
								$2,
								$3,
								$4
							)`,
			tranID,
			addressID,
			txOut.Value,
			pkScriptHex,
		); err != nil {
			return errors.Wrap(err, fmt.Sprintf("insert txout failed: %d, %d, %d, %s", tranID, addressID, txOut.Value, pkScriptHex))
		}
	}
	// fmt.Printf("\t class: %s, to: %s amount: %f \n", cls, addrs, float64(txOut.Value)/100000000)
	return nil
}
