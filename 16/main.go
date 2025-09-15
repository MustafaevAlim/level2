package main

import (
	"flag"
	"fmt"
	"wget/parser"
)

// "https://pkg.go.dev/net/http@go1.25.1#example-Get"
func main() {

	flagM := flag.Int("m", 0, "глубина скачивания ресурса")
	flag.Parse()
	p := parser.NewParserHTML(*flagM, flag.Args()[0])
	err := p.Parse()
	if err != nil {
		fmt.Println(err)
	}

	// u, _ := url.Parse("/assets/built/screen.css?v=c98b486ace")
	// fmt.Println(u.User.Username())
	// err = d.DownloadFile(d.GetPath("https://mainfo.ru/"), res.Body)
	// fmt.Println(err)

}
