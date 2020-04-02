package main

import (
	"flag"
	"fmt"
	"gopoc/lib"
	"net/http"
)

var (
	num     = flag.Int("num", 10, "threads num (todo)")
	timeout = flag.Duration("timeout", 10, "timeout")
	proxy   = flag.String("proxy", "", "http proxy")
	pocName = flag.String("poc", "", "single poc name")
	target  = flag.String("t", "", "target, http://www.test.com")
	//raw     = flag.String("r", "", "request raw file ")
)

func main() {
	flag.Parse()
	if *pocName != "" && *target != "" {
		err := lib.InitHttpClient(*num, *proxy, *timeout)
		if err != nil {
			fmt.Println(err)
			return
		}
		p, err := lib.LoadPocFromFile(*pocName)
		if err != nil {
			fmt.Println(err)
			return
		}
		req, err := http.NewRequest("GET", *target, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		result, err := lib.Check(req, p)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(result)
	}
}
