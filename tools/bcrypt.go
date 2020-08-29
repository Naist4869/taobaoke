package tools

import (
	"crypto/md5"
	"fmt"
)

var salt = "Salt123"

func Md5(str string) string {
	strData := str + salt
	data := []byte(strData)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has) // 将[]byte转成16进制
	return md5str
}
