package cmd

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"gopoc/lib"
	"gopoc/utils"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var (
	num        = flag.Int("num", 10, "threads num")
	timeout    = flag.Duration("t", 10*time.Second, "timeout")
	proxy      = flag.String("proxy", "", "http proxy")
	pocName    = flag.String("p", "", "single poc name")
	pocMulti   = flag.String("P", "", "multi poc select, eg: pocs/* or pocs/thinkphp.*")
	target     = flag.String("u", "", "target, http://www.test.com")
	targetFile = flag.String("l", "", "target list file")
	rawFile    = flag.String("r", "", "request rawFile file")
	cookie     = flag.String("c", "", "http cookie header, only for target")
	verbose    = flag.Bool("v", false, "verbose")
	debug      = flag.Bool("vv", false, "debug")
)

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, `gopoc version: 0.0.1
Usage: gopoc -t http://www.test.com -p test.yaml

Options:
`)
	flag.PrintDefaults()
}

func Execute() {
	flag.Parse()
	err := lib.InitHttpClient(*num, *proxy, *timeout)
	if err != nil {
		fmt.Println(err)
		return
	}
	utils.InitLog(*debug, *verbose)

	switch {
	case *target != "":
		req, err := http.NewRequest("GET", *target, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		if *cookie != "" {
			req.Header.Set("Cookie", *cookie)
		}
		if *pocName != "" {
			if poc := lib.CheckSinglePoc(req, *pocName); poc != nil {
				utils.Green("%v, %s", *target, poc.Name)
			}
		} else if *pocMulti != "" {
			lib.CheckMultiPoc(req, *pocMulti, *num)
		}
	case *targetFile != "":
		targets := utils.ReadingLines(*targetFile)
		if *pocName != "" {
			lib.BatchCheckSinglePoc(targets, *pocName, *num)
		} else if *pocMulti != "" {
			lib.BatchCheckMultiPoc(targets, *pocMulti, *num)
		}
	case *rawFile != "":
		raw, err := ioutil.ReadFile(*rawFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(raw)))
		if err != nil {
			fmt.Println(err)
			return
		}
		if *pocName != "" {
			if poc := lib.CheckSinglePoc(req, *pocName); poc != nil {
				utils.Green("%v, %s", *target, poc.Name)
			}
		} else if *pocMulti != "" {
			lib.CheckMultiPoc(req, *pocMulti, *num)
		}
	default:
		usage()
	}
}
