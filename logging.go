package main

import (
	"fmt"
	"os"
	"time"

	"github.com/go-kit/kit/log"
)

////////////////////////////////////////////////////////////////////////////
// Constant and data type/structure definitions

// Escape sequence
const esc = "\x1b"
const escSeq = esc + "\x5b"

////////////////////////////////////////////////////////////////////////////
// Global variables definitions

var (
	logger log.Logger
	debug  = 0
)

////////////////////////////////////////////////////////////////////////////
// Function definitions

//==========================================================================
// init

func init() {
	// https://godoc.org/github.com/go-kit/kit/log#TimestampFormat
	timestampFormat := log.TimestampFormat(time.Now, "0102T15:04:05") // 2006-01-02
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", timestampFormat)
	fmt.Println()
}

//==========================================================================
// support functions

func logIf(level int, message string, args ...interface{}) {
	if debug < level {
		return
	}
	p := make([]interface{}, 0)
	p = append(p, "msg")
	// ansi-bold + message + unbold
	//p = append(p, escSeq+"1m"+message+esc+"(B"+escSeq+"m")
	p = append(p, message)
	p = append(p, args...)
	//fmt.Printf("%#v\n", p)
	logger.Log(p...)
}
