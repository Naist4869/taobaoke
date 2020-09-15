package tools

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"reflect"
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
func NotNil(data interface{}, include []string, exclude []string) error {

	value := reflect.ValueOf(data)

	t := value.Type()
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
		value = value.Elem()
	}
	switch t.Kind() {
	case reflect.Struct, reflect.Slice, reflect.Array, reflect.Map, reflect.Chan, reflect.Func:
	default:
		return errors.New("不支持的类型")
	}
	if len(include) == 0 && len(exclude) == 0 {
		for i := 0; i < value.NumField(); i++ {
			field := value.Field(i)
			switch field.Type().Kind() {
			case reflect.Ptr, reflect.Slice, reflect.Array, reflect.Map, reflect.Chan, reflect.Func:
				if value.Field(i).IsNil() {
					return fmt.Errorf("字段[%s]为nil", t.Field(i).Name)
				}
			default:
				continue

			}

		}
	}
	return nil
}
