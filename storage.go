package db2sql

import (
	"fmt"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/webdeveloppro/cryptopiggy/pkg/address"
	"github.com/webdeveloppro/cryptopiggy/pkg/transaction"
)

// Storage general interface
type Storage interface {
	GetTxOutByIndex(hash string, index uint32) (*transaction.TxOut, error)
	GetAddressByHash(string) (*address.Address, error)
	getTransaction() *transaction.Transaction
}

// PGStorage postgresql storage
type PGStorage struct {
	con *pgx.ConnPool
}

// NewStorage will create new postgresql storage
func NewStorage(con *pgx.ConnPool) *PGStorage {
	return &PGStorage{
		con: con,
	}
}

// GetTxOutByIndex looks for a txout by transaction hash and index position
func (pg *PGStorage) GetTxOutByIndex(hash string, index uint32) (*transaction.TxOut, error) {

	txout := &transaction.TxOut{}
	sql := fmt.Sprintf(`SELECT 
		txout::json#>>'{%d,pk_script}', txout::json#>>'{%d,val}' FROM transaction 
		WHERE hash = $1`, index, index)

	if err := pg.con.QueryRow(sql, hash).Scan(&txout.PkScript, &txout.Value); err != nil {
		return txout, errors.Wrapf(err, "Storage: Cannot query: %v", err)
	}
	return txout, nil
}

// GetAddressByHash looks for an address by hash
func (pg *PGStorage) GetAddressByHash(hash string) (*address.Address, error) {
	addrStorage := address.NewStorage(pg.con)
	addr := address.New(&addrStorage)
	err := addr.GetByHash(hash)
	return addr, err
}

// return transaction with pg storage
func (pg *PGStorage) getTransaction() *transaction.Transaction {
	tranStorage := transaction.NewStorage(pg.con)
	return transaction.New(tranStorage)
}
