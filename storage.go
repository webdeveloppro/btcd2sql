package db2sql

import (
	"database/sql"
	"fmt"

	"github.com/vladyslav2/bitcoin2sql/pkg/transaction"

	"github.com/pkg/errors"
	"github.com/vladyslav2/bitcoin2sql/pkg/address"
)

// Storage general interface
type Storage interface {
	GetPKScript(hash string, index uint32) (string, error)
	GetAddressByHash(string) (*address.Address, error)
	getTransaction() *transaction.Transaction
}

// PGStorage postgresql storage
type PGStorage struct {
	con *sql.DB
}

// NewStorage will create new postgresql storage
func NewStorage(con *sql.DB) *PGStorage {
	return &PGStorage{
		con: con,
	}
}

// GetPKScript looks for a pk script in txout jsonb
func (pg *PGStorage) GetPKScript(hash string, index uint32) (string, error) {

	var PkScript string
	sql := fmt.Sprintf(`SELECT 
		txout::json#>>'{%d,pk_script}' FROM transaction 
		WHERE hash = $1`, index)

	if err := pg.con.QueryRow(sql, hash).Scan(&PkScript); err != nil {
		return PkScript, errors.Wrapf(err, "Storage: Cannot query: %v", err)
	}
	return PkScript, nil
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
