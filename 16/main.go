package main

import (
	"flag"
	"fmt"
	"wget/parser"
)

func main() {

	flagM := flag.Int("m", 0, "глубина скачивания ресурса")
	flag.Parse()
	p := parser.NewParserHTML(*flagM, flag.Args()[0])
	err := p.Parse()
	if err != nil {
		fmt.Println(err)
	}

}
