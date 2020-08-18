package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/decred/dcrd/txscript/v3"
)

func exitUsage() {
	fmt.Printf(("Usage: %s [hex-script]\n"), filepath.Base(os.Args[0]))
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		exitUsage()
	}

	version := uint16(0)
	script, err := hex.DecodeString(os.Args[1])
	if err != nil {
		exitUsage()
	}

	compress := false

	var out strings.Builder
	tkn := txscript.MakeScriptTokenizer(version, script)
	for tkn.Next() {
		err := txscript.DisasmOpcode(&out, tkn.Opcode(), tkn.Data(), compress)
		if err != nil {
			fmt.Printf("Error disasming opcode: %v\n", err)
		}
		out.WriteString(" ")
	}

	if tkn.Err() != nil {
		fmt.Printf("Error parsing script: %v\n", err)
	}

	fmt.Printf("Output:\n%s\n", out.String())
}
