module github.com/matheusd/dcr-disasm

go 1.14

replace github.com/decred/dcrd/txscript/v3 => ./txscript_vendored

require (
	github.com/decred/dcrd/dcrec/secp256k1/v3 v3.0.0-20200818052744-5bb9f3e87ff3 // indirect
	github.com/decred/dcrd/txscript/v3 v3.0.0-00010101000000-000000000000
)
