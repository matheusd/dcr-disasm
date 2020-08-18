// Copyright (c) 2015-2017 The btcsuite developers
// Copyright (c) 2015-2020 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txscript

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"
)

// hexToBytes converts the passed hex string into bytes and will panic if there
// is an error.  This is only provided for the hard-coded constants so errors in
// the source code can be detected. It will only (and must only) be called with
// hard-coded values.
func hexToBytes(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("invalid hex in source file: " + s)
	}
	return b
}

// TestScriptNumBytes ensures that converting from integral script numbers to
// byte representations works as expected.
func TestScriptNumBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		num        ScriptNum
		serialized []byte
	}{
		{0, nil},
		{1, hexToBytes("01")},
		{-1, hexToBytes("81")},
		{127, hexToBytes("7f")},
		{-127, hexToBytes("ff")},
		{128, hexToBytes("8000")},
		{-128, hexToBytes("8080")},
		{129, hexToBytes("8100")},
		{-129, hexToBytes("8180")},
		{256, hexToBytes("0001")},
		{-256, hexToBytes("0081")},
		{32767, hexToBytes("ff7f")},
		{-32767, hexToBytes("ffff")},
		{32768, hexToBytes("008000")},
		{-32768, hexToBytes("008080")},
		{65535, hexToBytes("ffff00")},
		{-65535, hexToBytes("ffff80")},
		{524288, hexToBytes("000008")},
		{-524288, hexToBytes("000088")},
		{7340032, hexToBytes("000070")},
		{-7340032, hexToBytes("0000f0")},
		{8388608, hexToBytes("00008000")},
		{-8388608, hexToBytes("00008080")},
		{2147483647, hexToBytes("ffffff7f")},
		{-2147483647, hexToBytes("ffffffff")},

		// Values that are out of range for data that is interpreted as
		// numbers, but are allowed as the result of numeric operations.
		{2147483648, hexToBytes("0000008000")},
		{-2147483648, hexToBytes("0000008080")},
		{2415919104, hexToBytes("0000009000")},
		{-2415919104, hexToBytes("0000009080")},
		{4294967295, hexToBytes("ffffffff00")},
		{-4294967295, hexToBytes("ffffffff80")},
		{4294967296, hexToBytes("0000000001")},
		{-4294967296, hexToBytes("0000000081")},
		{281474976710655, hexToBytes("ffffffffffff00")},
		{-281474976710655, hexToBytes("ffffffffffff80")},
		{72057594037927935, hexToBytes("ffffffffffffff00")},
		{-72057594037927935, hexToBytes("ffffffffffffff80")},
		{9223372036854775807, hexToBytes("ffffffffffffff7f")},
		{-9223372036854775807, hexToBytes("ffffffffffffffff")},
	}

	for _, test := range tests {
		gotBytes := test.num.Bytes()
		if !bytes.Equal(gotBytes, test.serialized) {
			t.Errorf("Bytes: did not get expected bytes for %d - "+
				"got %x, want %x", test.num, gotBytes,
				test.serialized)
			continue
		}
	}
}

// TestMakeScriptNum ensures that converting from byte representations to
// integral script numbers works as expected.
func TestMakeScriptNum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		serialized []byte
		num        ScriptNum
		numLen     int
		err        error
	}{
		// Minimal encoding must reject negative 0.
		{hexToBytes("80"), 0, MathOpCodeMaxScriptNumLen, ErrMinimalData},

		// Minimally encoded valid values.  Should not error and return
		// expected integral number.
		{nil, 0, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("01"), 1, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("81"), -1, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("7f"), 127, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("ff"), -127, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("8000"), 128, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("8080"), -128, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("8100"), 129, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("8180"), -129, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("0001"), 256, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("0081"), -256, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("ff7f"), 32767, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("ffff"), -32767, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("008000"), 32768, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("008080"), -32768, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("ffff00"), 65535, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("ffff80"), -65535, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("000008"), 524288, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("000088"), -524288, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("000070"), 7340032, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("0000f0"), -7340032, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("00008000"), 8388608, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("00008080"), -8388608, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("ffffff7f"), 2147483647, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("ffffffff"), -2147483647, MathOpCodeMaxScriptNumLen, nil},
		{hexToBytes("ffffffff7f"), 549755813887, 5, nil},
		{hexToBytes("ffffffffff"), -549755813887, 5, nil},
		{hexToBytes("ffffffffffffff7f"), 9223372036854775807, 8, nil},
		{hexToBytes("ffffffffffffffff"), -9223372036854775807, 8, nil},
		{hexToBytes("ffffffffffffffff7f"), -1, 9, nil},
		{hexToBytes("ffffffffffffffffff"), 1, 9, nil},
		{hexToBytes("ffffffffffffffffff7f"), -1, 10, nil},
		{hexToBytes("ffffffffffffffffffff"), 1, 10, nil},

		// Minimally encoded values that are out of range for data that
		// is interpreted as script numbers.  Should error and return 0.
		{hexToBytes("0000008000"), 0, MathOpCodeMaxScriptNumLen, ErrNumOutOfRange},
		{hexToBytes("0000008080"), 0, MathOpCodeMaxScriptNumLen, ErrNumOutOfRange},
		{hexToBytes("0000009000"), 0, MathOpCodeMaxScriptNumLen, ErrNumOutOfRange},
		{hexToBytes("0000009080"), 0, MathOpCodeMaxScriptNumLen, ErrNumOutOfRange},
		{hexToBytes("ffffffff00"), 0, MathOpCodeMaxScriptNumLen, ErrNumOutOfRange},
		{hexToBytes("ffffffff80"), 0, MathOpCodeMaxScriptNumLen, ErrNumOutOfRange},
		{hexToBytes("0000000001"), 0, MathOpCodeMaxScriptNumLen, ErrNumOutOfRange},
		{hexToBytes("0000000081"), 0, MathOpCodeMaxScriptNumLen, ErrNumOutOfRange},
		{hexToBytes("ffffffffffff00"), 0, MathOpCodeMaxScriptNumLen, ErrNumOutOfRange},
		{hexToBytes("ffffffffffff80"), 0, MathOpCodeMaxScriptNumLen, ErrNumOutOfRange},
		{hexToBytes("ffffffffffffff00"), 0, MathOpCodeMaxScriptNumLen, ErrNumOutOfRange},
		{hexToBytes("ffffffffffffff80"), 0, MathOpCodeMaxScriptNumLen, ErrNumOutOfRange},
		{hexToBytes("ffffffffffffff7f"), 0, MathOpCodeMaxScriptNumLen, ErrNumOutOfRange},
		{hexToBytes("ffffffffffffffff"), 0, MathOpCodeMaxScriptNumLen, ErrNumOutOfRange},

		// Non-minimally encoded, but otherwise valid values.  Should
		// error and return 0.
		{hexToBytes("00"), 0, MathOpCodeMaxScriptNumLen, ErrMinimalData},       // 0
		{hexToBytes("0100"), 0, MathOpCodeMaxScriptNumLen, ErrMinimalData},     // 1
		{hexToBytes("7f00"), 0, MathOpCodeMaxScriptNumLen, ErrMinimalData},     // 127
		{hexToBytes("800000"), 0, MathOpCodeMaxScriptNumLen, ErrMinimalData},   // 128
		{hexToBytes("810000"), 0, MathOpCodeMaxScriptNumLen, ErrMinimalData},   // 129
		{hexToBytes("000100"), 0, MathOpCodeMaxScriptNumLen, ErrMinimalData},   // 256
		{hexToBytes("ff7f00"), 0, MathOpCodeMaxScriptNumLen, ErrMinimalData},   // 32767
		{hexToBytes("00800000"), 0, MathOpCodeMaxScriptNumLen, ErrMinimalData}, // 32768
		{hexToBytes("ffff0000"), 0, MathOpCodeMaxScriptNumLen, ErrMinimalData}, // 65535
		{hexToBytes("00000800"), 0, MathOpCodeMaxScriptNumLen, ErrMinimalData}, // 524288
		{hexToBytes("00007000"), 0, MathOpCodeMaxScriptNumLen, ErrMinimalData}, // 7340032
		{hexToBytes("0009000100"), 0, 5, ErrMinimalData},                       // 16779520
	}

	for _, test := range tests {
		// Ensure the error matches the value specified in the test instance.
		gotNum, err := MakeScriptNum(test.serialized, test.numLen)
		if !errors.Is(err, test.err) {
			t.Errorf("MakeScriptNum(%#x): unexpected error - got %v, want %v",
				test.serialized, err, test.err)
			continue
		}

		if gotNum != test.num {
			t.Errorf("MakeScriptNum(%#x): did not get expected number - got "+
				"%d, want %d", test.serialized, gotNum, test.num)
			continue
		}
	}
}

// TestScriptNumInt32 ensures that the Int32 function on script number behaves
// as expected.
func TestScriptNumInt32(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   ScriptNum
		want int32
	}{
		// Values inside the valid int32 range are just the values
		// themselves cast to an int32.
		{0, 0},
		{1, 1},
		{-1, -1},
		{127, 127},
		{-127, -127},
		{128, 128},
		{-128, -128},
		{129, 129},
		{-129, -129},
		{256, 256},
		{-256, -256},
		{32767, 32767},
		{-32767, -32767},
		{32768, 32768},
		{-32768, -32768},
		{65535, 65535},
		{-65535, -65535},
		{524288, 524288},
		{-524288, -524288},
		{7340032, 7340032},
		{-7340032, -7340032},
		{8388608, 8388608},
		{-8388608, -8388608},
		{2147483647, 2147483647},
		{-2147483647, -2147483647},
		{-2147483648, -2147483648},

		// Values outside of the valid int32 range are limited to int32.
		{2147483648, 2147483647},
		{-2147483649, -2147483648},
		{1152921504606846975, 2147483647},
		{-1152921504606846975, -2147483648},
		{2305843009213693951, 2147483647},
		{-2305843009213693951, -2147483648},
		{4611686018427387903, 2147483647},
		{-4611686018427387903, -2147483648},
		{9223372036854775807, 2147483647},
		{-9223372036854775808, -2147483648},
	}

	for _, test := range tests {
		got := test.in.Int32()
		if got != test.want {
			t.Errorf("Int32: did not get expected value for %d - "+
				"got %d, want %d", test.in, got, test.want)
			continue
		}
	}
}
