package db2sql

/*
import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ripemd160"
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
	typ, addresses, _, err := txscript.ExtractPkScriptAddrs(
		txOut.PkScript, &chaincfg.MainNetParams)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Cannot extract pkScript %s", pkScriptHex))
	}

	// dirty hack for weired transactions
	// like 9969603dca74d14d29d1d5f56b94c7872551607f8c2d6837ab9715c60721b50e
	if typ.String() == "nonstandard" {
		rp := ripemd160.New()
		_, err = rp.Write([]byte(pkScriptHex))
		if err != nil {
			return err
		}
		bcipher := rp.Sum(nil)
		log.Printf("Nonstandard txout, transaction_id %d", tranID)
		B2SQL.insertAddressTxOUT(txOut, tranID, pkScriptHex, string(bcipher))
	}

	for _, a := range addresses {
		err = B2SQL.insertAddressTxOUT(txOut, tranID, pkScriptHex, a.EncodeAddress())
		if err != nil {
			return err
		}
	}
	return nil
}

func (B2SQL *Block2SQL) insertAddressTxOUT(txOut *wire.TxOut, tranID int, pkScriptHex, address string) error {
	addressID := 0
	if err := B2SQL.pg.QueryRow(`
		INSERT INTO address (hash, income, ballance)
		VALUES ($1, $2, $3)
		ON CONFLICT (hash) DO UPDATE SET
		income = address.income + $2,
		ballance = address.ballance + $3
		RETURNING ID`,
		address,
		txOut.Value,
		txOut.Value,
	).Scan(&addressID); err != nil {
		return errors.Wrap(err, "insert or update address failed")
	}

	if _, err := B2SQL.pg.Exec(`
		INSERT INTO txout (transaction_id, address_id, val, pk_script)
		VALUES ($1, $2,	$3, $4)`,
		tranID,
		addressID,
		txOut.Value,
		pkScriptHex,
	); err != nil {
		return errors.Wrap(err, fmt.Sprintf("insert txout failed: %d, %d, %d, %s", tranID, addressID, txOut.Value, pkScriptHex))
	}
	return nil
}
*/
