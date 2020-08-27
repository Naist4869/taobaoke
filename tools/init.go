package tools

import (
	"flag"
	"time"
	_ "time/tzdata"
)

func init() {
	var err error
	if location, err = time.LoadLocation("Asia/Shanghai"); err != nil {
		panic(err.Error())
	}
}
func IsDebug() bool {
	return flag.Lookup("test.v") != nil
}
