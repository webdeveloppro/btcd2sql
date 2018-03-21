package db2sql

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil/base58"
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
