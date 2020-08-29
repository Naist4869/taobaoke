package tools

import (
	"crypto/rand"
	"math/big"
)

func MakeNonce() (nonce string) {
	limit := big.NewInt(1)
	limit = limit.Lsh(limit, 64)
	n, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return
	}
	nonce = n.String()
	return
}

func Unique(ids []string, AllowEmpty bool) (res []string) {
	res = make([]string, 0, len(ids))
	mm := map[string]bool{}
	for _, id := range ids {
		_, exist := mm[id]
		if exist || (!AllowEmpty && id == "") {
			continue
		}
		res = append(res, id)
		mm[id] = true
	}
	return
}

func Separate(number string) string {
	integerPart, decimalPart := separate(number)
	return integerPart + "." + decimalPart
}
func separate(number string) (integerPart string, decimalPart string) {
	switch len(number) {
	case 0:
		decimalPart = "00"
		integerPart = "0"
	case 1:
		decimalPart = "0" + number
		integerPart = "0"
	case 2:
		decimalPart = number
		integerPart = "0"
	default:
		integerPart = number[:len(number)-2]
		decimalPart = number[len(number)-2:]
	}
	return
}
