package tools

import (
	"flag"
	"io/ioutil"
	"time"
)

var (
	debug = IsDebug()
)

func Init() {
	if !debug {
		if data, err := ioutil.ReadFile(locationFilePath); err != nil {
			panic("时间文件不存在")
		} else {
			if location, err = time.LoadLocationFromTZData("Asia/shanghai", data); err != nil {
				panic(err.Error())
			}
		}
	} else {
		var err error
		if location, err = time.LoadLocation("Asia/Shanghai"); err != nil {
			panic(err.Error())
		}
	}
}

func IsDebug() bool {
	return flag.Lookup("test.v") != nil
}
