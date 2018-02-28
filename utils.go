package main

import (
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil/base58"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ripemd160"
)

// GetInputAddress will return sender addres from signature script
func GetInputAddress(pubKeyHex string) (string, error) {
	decoded, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return "", err
	}
	// pack with sha256 once
	sha := chainhash.HashB(decoded)

	// encode with ripemd160 once
	rp := ripemd160.New()
	_, err = rp.Write(sha)
	if err != nil {
		return "", err
	}
	bcipher := rp.Sum(nil)

	// fill first byte with \x0
	one := make([]byte, 1)
	one[0] = 0x00
	bcipher = append(one[:], bcipher[:]...)

	// append data with last 4 bytes of sha256^2(data)
	res := append(bcipher[:], chainhash.DoubleHashB(bcipher)[:4]...)
	return base58.Encode(res), nil
}

func FindPrevAddress(pg *sql.DB, hash string, offset uint32) (string, error) {

	var PkScript string
	sql := `SELECT
		tx.pk_script
		FROM transaction as t JOIN txout as tx ON t.id = tx.transaction_id
		WHERE hash = $1 limit 1 offset $2`

	if err := pg.QueryRow(sql, hash, offset).Scan(&PkScript); err != nil {
		return "", errors.Wrap(err, "utils: Cannot get address from txout")
	}

	dst := make([]byte, hex.DecodedLen(len(PkScript)))
	_, err := hex.Decode(dst, []byte(PkScript))
	if err != nil {
		return "", errors.Wrap(err, "block: Cannot convert hex string to bytes")
	}

	_, addresses, _, err := txscript.ExtractPkScriptAddrs(dst, &chaincfg.MainNetParams)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Cannot extract pkScript %s", PkScript))
	}

	return addresses[0].EncodeAddress(), nil
}
