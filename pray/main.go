package main

import (
	"flag"
	"fmt"
	statsgod "github.com/laslowh/statsgod/client"
	"os"
)

var fAddr = flag.String("a", "", "server address")
var fBufferSize = flag.Int64("b", 1, "buffer size")
var fKey = flag.String("k", "", "key")
var fValue = flag.Float64("v", 0.0, "value")

func main() {
	flag.Parse()

	if *fAddr == "" {
		fmt.Fprint(os.Stderr, "'a' flag required\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *fKey == "" {
		fmt.Fprint(os.Stderr, "'k' flag required\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	c, err := statsgod.Dial(*fAddr, int(*fBufferSize))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: could not connect to server", err)
		os.Exit(1)
	}
	c.Update(*fKey, *fValue)
	c.Close()
}
