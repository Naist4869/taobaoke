//+build sdkcodegen

package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	log.Println("开始生成代码")
	// TODO: error handling
	filename := os.Args[1]
	var destFilename string
	var pkg string
	if len(os.Args) == 3 {
		destFilename = os.Args[2]
	} else {
		// blindly append `.go` so the result looks like `foo.md.go`
		destFilename = filename + ".go"
	}
	log.Println("目标文件名", destFilename)
	pkg, _ = filepath.Split(destFilename)
	log.Println("目标文件所在目录", pkg)
	split := strings.Split(pkg, `/`)
	if len(split)-2 >= 0 {
		pkg = split[len(split)-2]
	}

	emitToStdout := destFilename == "-"

	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open input file failed: %+v\n", err)
		os.Exit(1)
		return // unreachable
	}
	log.Println("markdown文件名", file.Name())

	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read input failed: %+v\n", err)
		os.Exit(1)
		return // unreachable
	}
	log.Println("读取成功")

	mdRoot := parseDocument(content)
	hir, err := analyzeDocument(mdRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "syntax error in spec: %+v\n", err)
		os.Exit(1)
		return // unreachable
	}
	log.Println("分析文件成功")

	var sink io.Writer
	if emitToStdout {
		sink = os.Stdout
		log.Println("写入标准输出")

	} else {
		file, err := os.Create(destFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "open '%s' for writing failed: %+v\n", destFilename, err)
			os.Exit(1)
			return // unreachable
		}
		log.Println("创建文件成功")

		bufWriter := bufio.NewWriter(file)
		sink = bufWriter
		defer func() {
			bufWriter.Flush()
			file.Close()
		}()
	}
	em := &goEmitter{
		Sink: sink,
	}

	err = em.EmitCode(&hir, pkg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "code emission failed: %+v\n", err)
		os.Exit(1)
		return // unreachable
	}
	log.Println("填充文件成功")

	err = em.Finalize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "finalization failed: %+v\n", err)
		os.Exit(1)
		return // unreachable
	}
	log.Println("完成")

}
