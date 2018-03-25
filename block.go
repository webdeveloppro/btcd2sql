package db2sql

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/webdeveloppro/cryptopiggy/pkg/block"
	"github.com/webdeveloppro/cryptopiggy/pkg/transaction"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"

	"github.com/pkg/errors"
)

// Block2SQL structure for holding Block information
type Block2SQL struct {
	ID      int
	Block   *block.Block
	storage Storage
}

// NewBlock Block2SQL struct constructor
func NewBlock(storage Storage, block *block.Block) *Block2SQL {
	blk := Block2SQL{
		Block:   block,
		storage: storage,
	}
	return &blk
}

// ConvertBTCD2SQL Will transfer blockchain block to SQL block
func (blk *Block2SQL) ConvertBTCD2SQL(BTCBlock *btcutil.Block) ([]string, error) {

	if BTCBlock.Hash().String() == "" {
		return []string{}, fmt.Errorf("Btcd2sql have empty hash, nothing to insert")
	}

	blk.Block.Hash = BTCBlock.Hash().String()
	blk.Block.Height = BTCBlock.Height()
	blk.Block.Bits = BTCBlock.MsgBlock().Header.Bits
	blk.Block.Nonce = BTCBlock.MsgBlock().Header.Nonce
	blk.Block.Version = BTCBlock.MsgBlock().Header.Version
	blk.Block.HashMerkleRoot = BTCBlock.MsgBlock().Header.MerkleRoot.String()
	blk.Block.HashPrevBlock = BTCBlock.MsgBlock().Header.PrevBlock.String()
	blk.Block.CreatedAt = BTCBlock.MsgBlock().Header.Timestamp
	return blk.ConvertTransactions(BTCBlock.Transactions())
}

// ConvertTransactions transfer btcd transactions to SQL ones
func (blk *Block2SQL) ConvertTransactions(transactions []*btcutil.Tx) ([]string, error) {

	addressesHash := make([]string, 0)
	for _, tran := range transactions {
		t := blk.storage.getTransaction()
		t.Hash = tran.Hash().String()
		t.HasWitness = tran.HasWitness()

		ins := tran.MsgTx().TxIn
		outs := tran.MsgTx().TxOut
		addressesIds := make([]uint, 0)

		// Gather all TxIn Information
		for _, txIn := range ins {
			txInStruct, err := blk.txin2struct(txIn)
			if err != nil {
				return nil, err
			}
			if txInStruct != nil {
				t.TxIns = append(t.TxIns, *txInStruct)
				addressesIds = append(addressesIds, txInStruct.AddressID)
				addressesHash = append(addressesHash, txInStruct.Address)

				// Log address changes
				// addressesQ = append("insert into address_log(AddressID, Amount, TransactionID, Block.TimeStamp)")
			}
		}

		// Gather all TxOut Information
		for _, txOut := range outs {
			txOutStruct, err := blk.txout2struct(txOut)
			if err != nil {
				return addressesHash, err
			}
			t.TxOuts = append(t.TxOuts, *txOutStruct)
			for _, a := range txOutStruct.Addresses {
				addressesIds = append(addressesIds, AddrStorage[a].ID)
				addressesHash = append(addressesHash, a)
				// Log address changes
				// addressesQ = append("insert into address_log(AddressID, Amount, TransactionID, Block.TimeStamp)")
			}
		}

		t.Addresses = addressesIds
		blk.Block.Transactions = append(blk.Block.Transactions, *t)

		/*
			err = t.Insert()
		*/
	}
	return addressesHash, nil
}

func (blk *Block2SQL) txin2struct(txIn *wire.TxIn) (*transaction.TxIn, error) {

	// Get previous address hash if that exists
	if txIn.PreviousOutPoint.Hash.String() != "0000000000000000000000000000000000000000000000000000000000000000" {

		txStruct := &transaction.TxIn{}
		addressHash := ""

		disbuf, err := txscript.DisasmString(txIn.SignatureScript)
		if err != nil {
			return txStruct, errors.Wrap(err, "txin2json: Cannot get txin address")
		}

		// maybe we can get value without looking for prev output?
		txout, err := blk.FindPrevTxOut(txIn.PreviousOutPoint.Hash.String(), txIn.PreviousOutPoint.Index)
		if err != nil {
			// return txStruct, errors.Wrap(err, "txin2struct: Cannot find previous hash")
			// dirty hack for weired transactions
			// like 9969603dca74d14d29d1d5f56b94c7872551607f8c2d6837ab9715c60721b50e

			if err == sql.ErrNoRows {
				badaddress := disbuf
				if len(badaddress) > 22 {
					badaddress = badaddress[0:22]
				}
				txout.Addresses = make([]string, 1)
				txout.Addresses[0] = fmt.Sprintf("nonstandard-%s", badaddress)
				addressHash = txout.Addresses[0]
			} else {
				return txStruct, errors.Wrap(err, "txin2struct: Cannot find previous hash")
			}
		}

		disbufArr := strings.Split(disbuf, " ")
		txStruct.SignatureScript = disbuf

		if len(disbufArr) > 1 {
			addressHash, err = GetInputAddress(disbufArr[1])
			if err != nil {
				return txStruct, errors.Wrap(err, "txin2json: Cannot get txin address")
			}
		}

		if err := blk.checkAddressHash(addressHash); err != nil {
			return txStruct, errors.Wrap(err, fmt.Sprintf("cant find address %s", addressHash))
		}

		txStruct.Address = addressHash
		txStruct.AddressID = AddrStorage[addressHash].ID
		txStruct.Amount = AddrStorage[addressHash].Ballance
		txStruct.PrevOut = txIn.PreviousOutPoint.Hash.String()
		txStruct.SignatureScript = disbuf
		txStruct.Sequence = txIn.Sequence
		txStruct.Size = txIn.SerializeSize()
		// ToDo
		// Add witness
		// txStruct.Witness = txIn.Witness

		// ToDo
		// Thats not right - you can send half of your money
		// have to get value from transaction
		AddrStorage[addressHash].Income -= AddrStorage[addressHash].Ballance
		AddrStorage[addressHash].Outcome += AddrStorage[addressHash].Ballance
		AddrStorage[addressHash].Ballance = 0

		return txStruct, nil
	}
	return nil, nil
}

// FindPrevTxOut looks for a blockchain address from previous hash function
func (blk *Block2SQL) FindPrevTxOut(hash string, index uint32) (*transaction.TxOut, error) {

	// ToDo
	// Addresses[0] - seems like a wrong assumption

	// Check if transaction is in current block
	for _, t := range blk.Block.Transactions {
		if t.Hash == hash {
			return &t.TxOuts[index], nil
		}
	}

	txout, err := blk.storage.GetTxOutByIndex(hash, index)
	if err != nil {
		return &transaction.TxOut{}, errors.Wrapf(err, "Utils: Cannot find txout.pk_script for trasn: %s, index: %d", hash, index)
	}

	dst := make([]byte, hex.DecodedLen(len(txout.PkScript)))
	if _, err := hex.Decode(dst, []byte(txout.PkScript)); err != nil {
		return txout, errors.Wrap(err, "Utils: Cannot convert hex string to bytes")
	}

	typ, addresses, _, err := txscript.ExtractPkScriptAddrs(dst, &chaincfg.MainNetParams)
	if err != nil {
		return txout, errors.Wrap(err, fmt.Sprintf("Cannot extract pkScript %s", txout.PkScript))
	}

	if typ == txscript.NonStandardTy {
		badaddress := string(txout.PkScript)

		if len(badaddress) > 22 {
			badaddress = badaddress[0:22]
		}
		txout.Addresses = make([]string, 1)
		txout.Addresses[0] = fmt.Sprintf("nonstandard-%s", badaddress)

		return txout, nil
	}

	txout.Addresses = make([]string, len(addresses))

	for i, a := range addresses {
		txout.Addresses[i] = a.EncodeAddress()
	}

	return txout, nil
}

func (blk *Block2SQL) txout2struct(txOut *wire.TxOut) (*transaction.TxOut, error) {
	txStruct := &transaction.TxOut{
		PkScript: hex.EncodeToString(txOut.PkScript),
		Value:    txOut.Value,
	}

	_, err := txStruct.GetAddresses()
	if err != nil {
		return txStruct, errors.Wrapf(err, "block: cannot convert txout to struct")
	}

	for _, addressHash := range txStruct.Addresses {

		if err := blk.checkAddressHash(addressHash); err != nil {
			return txStruct, errors.Wrap(err, fmt.Sprintf("block: cant find address %s", addressHash))
		}

		// ToDo:
		// https://blockchain.info/block/00000000689051c09ff2cd091cc4c22c10b965eb8db3ad5f032621cc36626175
		// if person move bitcoin to him self - do not update outcome
		AddrStorage[addressHash].Income += txStruct.Value
		AddrStorage[addressHash].Ballance = txStruct.Value
	}
	return txStruct, nil
}

func (blk *Block2SQL) checkAddressHash(hash string) error {
	if _, ok := AddrStorage[hash]; ok == false {
		addr, err := blk.storage.GetAddressByHash(hash)
		switch err {
		case sql.ErrNoRows:
			if er := addr.Save(); er != nil {
				return errors.Wrap(err, fmt.Sprintf("checkaddress: cant create new address %s", hash))
			}
		case nil:
		default:
			return errors.Wrap(err, fmt.Sprintf("checkaddress: cant find address %s", hash))
		}
		AddrStorage[hash] = addr
	}
	return nil
}

// Insert Block data to database
func (blk *Block2SQL) Insert(affectedAddresses []string) error {
	err := blk.Block.Insert()
	if err != nil {
		return err
	}

	for _, hash := range affectedAddresses {
		err := AddrStorage[hash].Save()
		if err != nil {
			return err
		}
	}
	return nil
}
