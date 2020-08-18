// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2015-2020 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txscript

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/dcrec"
	"github.com/decred/dcrd/dcrec/secp256k1/v3"
	"github.com/decred/dcrd/dcrutil/v3"
)

// mainNetParams is an instance of the main network parameters and is shared
// throughout the tests.
var mainNetParams = chaincfg.MainNetParams()

// mustParseShortForm parses the passed short form script and returns the
// resulting bytes.  It panics if an error occurs.  This is only used in the
// tests as a helper since the only way it can fail is if there is an error in
// the test source code.
func mustParseShortForm(script string) []byte {
	s, err := parseShortForm(script)
	if err != nil {
		panic("invalid short form script in test source: err " +
			err.Error() + ", script: " + script)
	}

	return s
}

// newAddressPubKey returns a new dcrutil.AddressPubKey from the provided
// serialized public key.  It panics if an error occurs.  This is only used in
// the tests as a helper since the only way it can fail is if there is an error
// in the test source code.
func newAddressPubKey(serializedPubKey []byte) dcrutil.Address {
	pubkey, err := secp256k1.ParsePubKey(serializedPubKey)
	if err != nil {
		panic("invalid public key in test source")
	}
	addr, err := dcrutil.NewAddressSecpPubKeyCompressed(pubkey, mainNetParams)
	if err != nil {
		panic("invalid public key in test source")
	}

	return addr
}

// newAddressPubKeyHash returns a new dcrutil.AddressPubKeyHash from the
// provided hash.  It panics if an error occurs.  This is only used in the tests
// as a helper since the only way it can fail is if there is an error in the
// test source code.
func newAddressPubKeyHash(pkHash []byte) dcrutil.Address {
	addr, err := dcrutil.NewAddressPubKeyHash(pkHash, mainNetParams,
		dcrec.STEcdsaSecp256k1)
	if err != nil {
		panic("invalid public key hash in test source")
	}

	return addr
}

// newAddressScriptHash returns a new dcrutil.AddressScriptHash from the
// provided hash.  It panics if an error occurs.  This is only used in the tests
// as a helper since the only way it can fail is if there is an error in the
// test source code.
func newAddressScriptHash(scriptHash []byte) dcrutil.Address {
	addr, err := dcrutil.NewAddressScriptHashFromHash(scriptHash, mainNetParams)
	if err != nil {
		panic("invalid script hash in test source")
	}

	return addr
}

// TestExtractPkScriptAddrs ensures that extracting the type, addresses, and
// number of required signatures from PkScripts works as intended.
func TestExtractPkScriptAddrs(t *testing.T) {
	t.Parallel()

	const scriptVersion = 0
	tests := []struct {
		name    string
		script  []byte
		addrs   []dcrutil.Address
		reqSigs int
		class   ScriptClass
		noparse bool
	}{
		{
			name: "standard p2pk with compressed pubkey (0x02)",
			script: hexToBytes("2102192d74d0cb94344c9569c2e779015" +
				"73d8d7903c3ebec3a957724895dca52c6b4ac"),
			addrs: []dcrutil.Address{
				newAddressPubKey(hexToBytes("02192d74d0cb9434" +
					"4c9569c2e77901573d8d7903c3ebec3a9577" +
					"24895dca52c6b4")),
			},
			reqSigs: 1,
			class:   PubKeyTy,
		},
		{
			name: "standard p2pk with uncompressed pubkey (0x04)",
			script: hexToBytes("410411db93e1dcdb8a016b49840f8c53b" +
				"c1eb68a382e97b1482ecad7b148a6909a5cb2e0eaddf" +
				"b84ccf9744464f82e160bfa9b8b64f9d4c03f999b864" +
				"3f656b412a3ac"),
			addrs: []dcrutil.Address{
				newAddressPubKey(hexToBytes("0411db93e1dcdb8a" +
					"016b49840f8c53bc1eb68a382e97b1482eca" +
					"d7b148a6909a5cb2e0eaddfb84ccf9744464" +
					"f82e160bfa9b8b64f9d4c03f999b8643f656" +
					"b412a3")),
			},
			reqSigs: 1,
			class:   PubKeyTy,
		},
		{
			name: "standard p2pk with compressed pubkey (0x03)",
			script: hexToBytes("2103b0bd634234abbb1ba1e986e884185" +
				"c61cf43e001f9137f23c2c409273eb16e65ac"),
			addrs: []dcrutil.Address{
				newAddressPubKey(hexToBytes("03b0bd634234abbb" +
					"1ba1e986e884185c61cf43e001f9137f23c2" +
					"c409273eb16e65")),
			},
			reqSigs: 1,
			class:   PubKeyTy,
		},
		{
			name: "2nd standard p2pk with uncompressed pubkey (0x04)",
			script: hexToBytes("4104b0bd634234abbb1ba1e986e884185" +
				"c61cf43e001f9137f23c2c409273eb16e6537a576782" +
				"eba668a7ef8bd3b3cfb1edb7117ab65129b8a2e681f3" +
				"c1e0908ef7bac"),
			addrs: []dcrutil.Address{
				newAddressPubKey(hexToBytes("04b0bd634234abbb" +
					"1ba1e986e884185c61cf43e001f9137f23c2" +
					"c409273eb16e6537a576782eba668a7ef8bd" +
					"3b3cfb1edb7117ab65129b8a2e681f3c1e09" +
					"08ef7b")),
			},
			reqSigs: 1,
			class:   PubKeyTy,
		},
		{
			name: "standard p2pkh",
			script: hexToBytes("76a914ad06dd6ddee55cbca9a9e3713bd" +
				"7587509a3056488ac"),
			addrs: []dcrutil.Address{
				newAddressPubKeyHash(hexToBytes("ad06dd6ddee5" +
					"5cbca9a9e3713bd7587509a30564")),
			},
			reqSigs: 1,
			class:   PubKeyHashTy,
		},
		{
			name: "standard p2sh",
			script: hexToBytes("a91463bcc565f9e68ee0189dd5cc67f1b" +
				"0e5f02f45cb87"),
			addrs: []dcrutil.Address{
				newAddressScriptHash(hexToBytes("63bcc565f9e6" +
					"8ee0189dd5cc67f1b0e5f02f45cb")),
			},
			reqSigs: 1,
			class:   ScriptHashTy,
		},
		// from real tx 60a20bd93aa49ab4b28d514ec10b06e1829ce6818ec06cd3aabd013ebcdc4bb1, vout 0
		{
			name: "standard 1 of 2 multisig",
			script: hexToBytes("514104cc71eb30d653c0c3163990c47b9" +
				"76f3fb3f37cccdcbedb169a1dfef58bbfbfaff7d8a47" +
				"3e7e2e6d317b87bafe8bde97e3cf8f065dec022b51d1" +
				"1fcdd0d348ac4410461cbdcc5409fb4b4d42b51d3338" +
				"1354d80e550078cb532a34bfa2fcfdeb7d76519aecc6" +
				"2770f5b0e4ef8551946d8a540911abe3e7854a26f39f" +
				"58b25c15342af52ae"),
			addrs: []dcrutil.Address{
				newAddressPubKey(hexToBytes("04cc71eb30d653c0" +
					"c3163990c47b976f3fb3f37cccdcbedb169a" +
					"1dfef58bbfbfaff7d8a473e7e2e6d317b87b" +
					"afe8bde97e3cf8f065dec022b51d11fcdd0d" +
					"348ac4")),
				newAddressPubKey(hexToBytes("0461cbdcc5409fb4" +
					"b4d42b51d33381354d80e550078cb532a34b" +
					"fa2fcfdeb7d76519aecc62770f5b0e4ef855" +
					"1946d8a540911abe3e7854a26f39f58b25c1" +
					"5342af")),
			},
			reqSigs: 1,
			class:   MultiSigTy,
		},
		// from real tx d646f82bd5fbdb94a36872ce460f97662b80c3050ad3209bef9d1e398ea277ab, vin 1
		{
			name: "standard 2 of 3 multisig",
			script: hexToBytes("524104cb9c3c222c5f7a7d3b9bd152f36" +
				"3a0b6d54c9eb312c4d4f9af1e8551b6c421a6a4ab0e2" +
				"9105f24de20ff463c1c91fcf3bf662cdde4783d4799f" +
				"787cb7c08869b4104ccc588420deeebea22a7e900cc8" +
				"b68620d2212c374604e3487ca08f1ff3ae12bdc63951" +
				"4d0ec8612a2d3c519f084d9a00cbbe3b53d071e9b09e" +
				"71e610b036aa24104ab47ad1939edcb3db65f7fedea6" +
				"2bbf781c5410d3f22a7a3a56ffefb2238af8627363bd" +
				"f2ed97c1f89784a1aecdb43384f11d2acc64443c7fc2" +
				"99cef0400421a53ae"),
			addrs: []dcrutil.Address{
				newAddressPubKey(hexToBytes("04cb9c3c222c5f7a" +
					"7d3b9bd152f363a0b6d54c9eb312c4d4f9af" +
					"1e8551b6c421a6a4ab0e29105f24de20ff46" +
					"3c1c91fcf3bf662cdde4783d4799f787cb7c" +
					"08869b")),
				newAddressPubKey(hexToBytes("04ccc588420deeeb" +
					"ea22a7e900cc8b68620d2212c374604e3487" +
					"ca08f1ff3ae12bdc639514d0ec8612a2d3c5" +
					"19f084d9a00cbbe3b53d071e9b09e71e610b" +
					"036aa2")),
				newAddressPubKey(hexToBytes("04ab47ad1939edcb" +
					"3db65f7fedea62bbf781c5410d3f22a7a3a5" +
					"6ffefb2238af8627363bdf2ed97c1f89784a" +
					"1aecdb43384f11d2acc64443c7fc299cef04" +
					"00421a")),
			},
			reqSigs: 2,
			class:   MultiSigTy,
		},

		// The below are nonstandard script due to things such as
		// invalid pubkeys, failure to parse, and not being of a
		// standard form.

		{
			name: "p2pk with uncompressed pk missing OP_CHECKSIG",
			script: hexToBytes("410411db93e1dcdb8a016b49840f8c53b" +
				"c1eb68a382e97b1482ecad7b148a6909a5cb2e0eaddf" +
				"b84ccf9744464f82e160bfa9b8b64f9d4c03f999b864" +
				"3f656b412a3"),
			addrs:   nil,
			reqSigs: 0,
			class:   NonStandardTy,
		},
		{
			name: "valid signature from a sigscript - no addresses",
			script: hexToBytes("47304402204e45e16932b8af514961a1d" +
				"3a1a25fdf3f4f7732e9d624c6c61548ab5fb8cd41022" +
				"0181522ec8eca07de4860a4acdd12909d831cc56cbba" +
				"c4622082221a8768d1d0901"),
			addrs:   nil,
			reqSigs: 0,
			class:   NonStandardTy,
		},
		// Note the technically the pubkey is the second item on the
		// stack, but since the address extraction intentionally only
		// works with standard PkScripts, this should not return any
		// addresses.
		{
			name: "valid sigscript to redeem p2pk - no addresses",
			script: hexToBytes("493046022100ddc69738bf2336318e4e0" +
				"41a5a77f305da87428ab1606f023260017854350ddc0" +
				"22100817af09d2eec36862d16009852b7e3a0f6dd765" +
				"98290b7834e1453660367e07a014104cd4240c198e12" +
				"523b6f9cb9f5bed06de1ba37e96a1bbd13745fcf9d11" +
				"c25b1dff9a519675d198804ba9962d3eca2d5937d58e" +
				"5a75a71042d40388a4d307f887d"),
			addrs:   nil,
			reqSigs: 0,
			class:   NonStandardTy,
		},
		// adapted from btc:
		// tx 691dd277dc0e90a462a3d652a1171686de49cf19067cd33c7df0392833fb986a, vout 0
		// invalid public keys
		{
			name: "1 of 3 multisig with invalid pubkeys",
			script: hexToBytes("5141042200007353455857696b696c656" +
				"16b73204361626c6567617465204261636b75700a0a6" +
				"361626c65676174652d3230313031323034313831312" +
				"e377a0a0a446f41046e6c6f61642074686520666f6c6" +
				"c6f77696e67207472616e73616374696f6e732077697" +
				"468205361746f736869204e616b616d6f746f2773206" +
				"46f776e6c6f61410420746f6f6c2077686963680a636" +
				"16e20626520666f756e6420696e207472616e7361637" +
				"4696f6e2036633533636439383731313965663739376" +
				"435616463636453ae"),
			addrs:   []dcrutil.Address{},
			reqSigs: 1,
			class:   MultiSigTy,
		},
		// adapted from btc:
		// tx 691dd277dc0e90a462a3d652a1171686de49cf19067cd33c7df0392833fb986a, vout 44
		// invalid public keys
		{
			name: "1 of 3 multisig with invalid pubkeys 2",
			script: hexToBytes("514104633365633235396337346461636" +
				"536666430383862343463656638630a6336366263313" +
				"93936633862393461333831316233363536313866653" +
				"16539623162354104636163636539393361333938386" +
				"134363966636336643664616266640a3236363363666" +
				"13963663463303363363039633539336333653931666" +
				"56465373032392102323364643432643235363339643" +
				"338613663663530616234636434340a00000053ae"),
			addrs:   []dcrutil.Address{},
			reqSigs: 1,
			class:   MultiSigTy,
		},
		{
			name:    "empty script",
			script:  []byte{},
			addrs:   nil,
			reqSigs: 0,
			class:   NonStandardTy,
		},
		{
			name:    "script that does not parse",
			script:  []byte{OP_DATA_45},
			addrs:   nil,
			reqSigs: 0,
			class:   NonStandardTy,
			noparse: true,
		},
	}

	// Run tests with treasury disabled.
	t.Logf("Running %d tests without treasury agenda.", len(tests))
	for i, test := range tests {
		class, addrs, reqSigs, err := ExtractPkScriptAddrs(scriptVersion,
			test.script, mainNetParams, noTreasury)
		if err != nil && !test.noparse {
			t.Errorf("ExtractPkScriptAddrs #%d (%s): %v", i,
				test.name, err)
		}

		if !reflect.DeepEqual(addrs, test.addrs) {
			t.Errorf("ExtractPkScriptAddrs #%d (%s) unexpected "+
				"addresses\ngot  %v\nwant %v", i, test.name,
				addrs, test.addrs)
			continue
		}

		if reqSigs != test.reqSigs {
			t.Errorf("ExtractPkScriptAddrs #%d (%s) unexpected "+
				"number of required signatures - got %d, "+
				"want %d", i, test.name, reqSigs, test.reqSigs)
			continue
		}

		if class != test.class {
			t.Errorf("ExtractPkScriptAddrs #%d (%s) unexpected "+
				"script type - got %s, want %s", i, test.name,
				class, test.class)
			continue
		}
	}

	// Run same tests with treasury agenda active.
	t.Logf("Running %d tests with treasury agenda.", len(tests))
	for i, test := range tests {
		class, addrs, reqSigs, err := ExtractPkScriptAddrs(scriptVersion,
			test.script, mainNetParams, withTreasury)
		if err != nil && !test.noparse {
			t.Errorf("ExtractPkScriptAddrs #%d (%s): %v", i,
				test.name, err)
		}

		if !reflect.DeepEqual(addrs, test.addrs) {
			t.Errorf("ExtractPkScriptAddrs #%d (%s) unexpected "+
				"addresses\ngot  %v\nwant %v", i, test.name,
				addrs, test.addrs)
			continue
		}

		if reqSigs != test.reqSigs {
			t.Errorf("ExtractPkScriptAddrs #%d (%s) unexpected "+
				"number of required signatures - got %d, "+
				"want %d", i, test.name, reqSigs, test.reqSigs)
			continue
		}

		if class != test.class {
			t.Errorf("ExtractPkScriptAddrs #%d (%s) unexpected "+
				"script type - got %s, want %s", i, test.name,
				class, test.class)
			continue
		}
	}
}

// bogusAddress implements the dcrutil.Address interface so the tests can ensure
// unsupported address types are handled properly.
type bogusAddress struct{}

// Address simply returns an empty string.  It exists to satisfy the
// dcrutil.Address interface.
func (b *bogusAddress) Address() string {
	return ""
}

// ScriptAddress simply returns an empty byte slice.  It exists to satisfy the
// dcrutil.Address interface.
func (b *bogusAddress) ScriptAddress() []byte {
	return nil
}

// Hash160 simply returns an empty byte slice.  It exists to satisfy the
// dcrutil.Address interface.
func (b *bogusAddress) Hash160() *[20]byte {
	return nil
}

// String simply returns an empty string.  It exists to satisfy the
// dcrutil.Address interface.
func (b *bogusAddress) String() string {
	return ""
}

// TestPayToAddrScript ensures the PayToAddrScript function generates the
// correct scripts for the various types of addresses.
func TestPayToAddrScript(t *testing.T) {
	t.Parallel()

	// 1MirQ9bwyQcGVJPwKUgapu5ouK2E2Ey4gX
	p2pkhMain, err := dcrutil.NewAddressPubKeyHash(hexToBytes("e34cce70c86"+
		"373273efcc54ce7d2a491bb4a0e84"), mainNetParams, dcrec.STEcdsaSecp256k1)
	if err != nil {
		t.Fatalf("Unable to create public key hash address: %v", err)
	}

	// Taken from transaction:
	// b0539a45de13b3e0403909b8bd1a555b8cbe45fd4e3f3fda76f3a5f52835c29d
	p2shMain, _ := dcrutil.NewAddressScriptHashFromHash(hexToBytes("e8c30"+
		"0c87986efa84c37c0519929019ef86eb5b4"), mainNetParams)
	if err != nil {
		t.Fatalf("Unable to create script hash address: %v", err)
	}

	//  mainnet p2pk 13CG6SJ3yHUXo4Cr2RY4THLLJrNFuG3gUg
	p2pkCompressedMain, err := dcrutil.NewAddressSecpPubKey(hexToBytes("02192d7"+
		"4d0cb94344c9569c2e77901573d8d7903c3ebec3a957724895dca52c6b4"),
		mainNetParams)
	if err != nil {
		t.Fatalf("Unable to create pubkey address (compressed): %v",
			err)
	}
	p2pkCompressed2Main, err := dcrutil.NewAddressSecpPubKey(hexToBytes("03b0b"+
		"d634234abbb1ba1e986e884185c61cf43e001f9137f23c2c409273eb16e65"),
		mainNetParams)
	if err != nil {
		t.Fatalf("Unable to create pubkey address (compressed 2): %v",
			err)
	}

	p2pkUncompressedMain := newAddressPubKey(hexToBytes("0411db" +
		"93e1dcdb8a016b49840f8c53bc1eb68a382e97b1482ecad7b148a6909a5cb2" +
		"e0eaddfb84ccf9744464f82e160bfa9b8b64f9d4c03f999b8643f656b412a3"))

	tests := []struct {
		in       dcrutil.Address
		expected string
		err      error
	}{
		// pay-to-pubkey-hash address on mainnet 0
		{
			p2pkhMain,
			"DUP HASH160 DATA_20 0xe34cce70c86373273efcc54ce7d2a4" +
				"91bb4a0e8488 CHECKSIG",
			nil,
		},
		// pay-to-script-hash address on mainnet 1
		{
			p2shMain,
			"HASH160 DATA_20 0xe8c300c87986efa84c37c0519929019ef8" +
				"6eb5b4 EQUAL",
			nil,
		},
		// pay-to-pubkey address on mainnet. compressed key. 2
		{
			p2pkCompressedMain,
			"DATA_33 0x02192d74d0cb94344c9569c2e77901573d8d7903c3" +
				"ebec3a957724895dca52c6b4 CHECKSIG",
			nil,
		},
		// pay-to-pubkey address on mainnet. compressed key (other way). 3
		{
			p2pkCompressed2Main,
			"DATA_33 0x03b0bd634234abbb1ba1e986e884185c61cf43e001" +
				"f9137f23c2c409273eb16e65 CHECKSIG",
			nil,
		},
		// pay-to-pubkey address on mainnet. for Decred this would
		// be uncompressed, but standard for Decred is 33 byte
		// compressed public keys.
		{
			p2pkUncompressedMain,
			"DATA_33 0x0311db93e1dcdb8a016b49840f8c53bc1eb68a382e97b" +
				"1482ecad7b148a6909a5cac",
			nil,
		},

		// Supported address types with nil pointers.
		{(*dcrutil.AddressPubKeyHash)(nil), "", ErrUnsupportedAddress},
		{(*dcrutil.AddressScriptHash)(nil), "", ErrUnsupportedAddress},
		{(*dcrutil.AddressSecpPubKey)(nil), "", ErrUnsupportedAddress},
		{(*dcrutil.AddressEdwardsPubKey)(nil), "", ErrUnsupportedAddress},
		{(*dcrutil.AddressSecSchnorrPubKey)(nil), "", ErrUnsupportedAddress},

		// Unsupported address type.
		{&bogusAddress{}, "", ErrUnsupportedAddress},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		pkScript, err := PayToAddrScript(test.in)
		if !errors.Is(err, test.err) {
			t.Errorf("PayToAddrScript #%d unexpected error - got %v, want %v",
				i, err, test.err)
			continue
		}

		expected := mustParseShortForm(test.expected)
		if !bytes.Equal(pkScript, expected) {
			t.Errorf("PayToAddrScript #%d got: %x\nwant: %x",
				i, pkScript, expected)
			continue
		}
	}
}

// TestMultiSigScript ensures the MultiSigScript function returns the expected
// scripts and errors.
func TestMultiSigScript(t *testing.T) {
	t.Parallel()

	//  mainnet p2pk 13CG6SJ3yHUXo4Cr2RY4THLLJrNFuG3gUg
	p2pkCompressedMain, err := dcrutil.NewAddressSecpPubKey(hexToBytes("02192d"+
		"74d0cb94344c9569c2e77901573d8d7903c3ebec3a957724895dca52c6b4"),
		mainNetParams)
	if err != nil {
		t.Fatalf("Unable to create pubkey address (compressed): %v",
			err)
	}
	p2pkCompressed2Main, err := dcrutil.NewAddressSecpPubKey(hexToBytes("03b0b"+
		"d634234abbb1ba1e986e884185c61cf43e001f9137f23c2c409273eb16e65"),
		mainNetParams)
	if err != nil {
		t.Fatalf("Unable to create pubkey address (compressed 2): %v",
			err)
	}

	p2pkUncompressedMain := newAddressPubKey(hexToBytes("0411d" +
		"b93e1dcdb8a016b49840f8c53bc1eb68a382e97b1482ecad7b148a6909a5c" +
		"b2e0eaddfb84ccf9744464f82e160bfa9b8b64f9d4c03f999b8643f656b41" +
		"2a3"))

	tests := []struct {
		keys      []*dcrutil.AddressSecpPubKey
		nrequired int
		expected  string
		err       error
	}{
		{
			[]*dcrutil.AddressSecpPubKey{
				p2pkCompressedMain,
				p2pkCompressed2Main,
			},
			1,
			"1 DATA_33 0x02192d74d0cb94344c9569c2e77901573d8d7903c" +
				"3ebec3a957724895dca52c6b4 DATA_33 0x03b0bd634" +
				"234abbb1ba1e986e884185c61cf43e001f9137f23c2c4" +
				"09273eb16e65 2 CHECKMULTISIG",
			nil,
		},
		{
			[]*dcrutil.AddressSecpPubKey{
				p2pkCompressedMain,
				p2pkCompressed2Main,
			},
			2,
			"2 DATA_33 0x02192d74d0cb94344c9569c2e77901573d8d7903c" +
				"3ebec3a957724895dca52c6b4 DATA_33 0x03b0bd634" +
				"234abbb1ba1e986e884185c61cf43e001f9137f23c2c4" +
				"09273eb16e65 2 CHECKMULTISIG",
			nil,
		},
		{
			[]*dcrutil.AddressSecpPubKey{
				p2pkCompressedMain,
				p2pkCompressed2Main,
			},
			3,
			"",
			ErrTooManyRequiredSigs,
		},
		{
			// By default compressed pubkeys are used in Decred.
			[]*dcrutil.AddressSecpPubKey{
				p2pkUncompressedMain.(*dcrutil.AddressSecpPubKey),
			},
			1,
			"1 DATA_33 0x0311db93e1dcdb8a016b49840f8c53bc1eb68a3" +
				"82e97b1482ecad7b148a6909a5c 1 CHECKMULTISIG",
			nil,
		},
		{
			[]*dcrutil.AddressSecpPubKey{
				p2pkUncompressedMain.(*dcrutil.AddressSecpPubKey),
			},
			2,
			"",
			ErrTooManyRequiredSigs,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		script, err := MultiSigScript(test.keys, test.nrequired)
		if !errors.Is(err, test.err) {
			t.Errorf("MultiSigScript #%d: unexpected error - got %v, want %v",
				i, err, test.err)
			continue
		}

		expected := mustParseShortForm(test.expected)
		if !bytes.Equal(script, expected) {
			t.Errorf("MultiSigScript #%d got: %x\nwant: %x",
				i, script, expected)
			continue
		}
	}
}

// TestCalcMultiSigStats ensures the CalcMutliSigStats function returns the
// expected errors.
func TestCalcMultiSigStats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		script string
		err    error
	}{
		{
			name: "short script",
			script: "0x046708afdb0fe5548271967f1a67130b7105cd6a828" +
				"e03909a67962e0ea1f61d",
			err: ErrNotMultisigScript,
		},
		{
			name: "stack underflow",
			script: "RETURN DATA_41 0x046708afdb0fe5548271967f1a" +
				"67130b7105cd6a828e03909a67962e0ea1f61deb649f6" +
				"bc3f4cef308",
			err: ErrNotMultisigScript,
		},
		{
			name: "multisig script",
			script: "1 DATA_33 0x0232abdc893e7f0631364d7fd01cb33d24da45329a0" +
				"0357b3a7886211ab414d55a 1 CHECKMULTISIG",
			err: nil,
		},
	}

	for _, test := range tests {
		script := mustParseShortForm(test.script)
		_, _, err := CalcMultiSigStats(script)
		if !errors.Is(err, test.err) {
			t.Errorf("%s: unexpected error - got %v, want %v", test.name, err,
				test.err)
			continue
		}
	}
}

// scriptClassTests houses several test scripts used to ensure various class
// determination is working as expected.  It's defined as a test global versus
// inside a function scope since this spans both the standard tests and the
// consensus tests (pay-to-script-hash is part of consensus).
var scriptClassTests = []struct {
	name     string
	script   string
	class    ScriptClass
	subClass ScriptClass
}{
	{
		name: "Pay Pubkey",
		script: "DATA_65 0x0411db93e1dcdb8a016b49840f8c53bc1eb68a382e" +
			"97b1482ecad7b148a6909a5cb2e0eaddfb84ccf9744464f82e16" +
			"0bfa9b8b64f9d4c03f999b8643f656b412a3 CHECKSIG",
		class: PubKeyTy,
	},
	// tx 599e47a8114fe098103663029548811d2651991b62397e057f0c863c2bc9f9ea
	{
		name: "Pay PubkeyHash",
		script: "DUP HASH160 DATA_20 0x660d4ef3a743e3e696ad990364e555" +
			"c271ad504b EQUALVERIFY CHECKSIG",
		class: PubKeyHashTy,
	},
	// part of tx 6d36bc17e947ce00bb6f12f8e7a56a1585c5a36188ffa2b05e10b4743273a74b
	// codeseparator parts have been elided. (bitcoin core's checks for
	// multisig type doesn't have codesep either).
	{
		name: "multisig",
		script: "1 DATA_33 0x0232abdc893e7f0631364d7fd01cb33d24da4" +
			"5329a00357b3a7886211ab414d55a 1 CHECKMULTISIG",
		class: MultiSigTy,
	},
	// tx e5779b9e78f9650debc2893fd9636d827b26b4ddfa6a8172fe8708c924f5c39d
	{
		name: "P2SH",
		script: "HASH160 DATA_20 0x433ec2ac1ffa1b7b7d027f564529c57197f" +
			"9ae88 EQUAL",
		class: ScriptHashTy,
	},
	{
		name: "Stake Submission P2SH",
		script: "SSTX HASH160 DATA_20 0x433ec2ac1ffa1b7b7d027f564529" +
			"c57197f9ae88 EQUAL",
		class:    StakeSubmissionTy,
		subClass: ScriptHashTy,
	},
	{
		name: "Stake Submission Generation P2SH",
		script: "SSGEN HASH160 DATA_20 0x433ec2ac1ffa1b7b7d027f564529" +
			"c57197f9ae88 EQUAL",
		class:    StakeGenTy,
		subClass: ScriptHashTy,
	},
	{
		name: "Stake Submission Revocation P2SH",
		script: "SSRTX HASH160 DATA_20 0x433ec2ac1ffa1b7b7d027f564529" +
			"c57197f9ae88 EQUAL",
		class:    StakeRevocationTy,
		subClass: ScriptHashTy,
	},
	{
		name: "Stake Submission Change P2SH",
		script: "SSTXCHANGE HASH160 DATA_20 0x433ec2ac1ffa1b7b7d027f5" +
			"64529c57197f9ae88 EQUAL",
		class:    StakeSubChangeTy,
		subClass: ScriptHashTy,
	},

	{
		// Nulldata with no data at all.
		name:   "nulldata no data",
		script: "RETURN",
		class:  NullDataTy,
	},
	{
		// Nulldata with single zero push.
		name:   "nulldata zero",
		script: "RETURN 0",
		class:  NullDataTy,
	},
	{
		// Nulldata with small integer push.
		name:   "nulldata small int",
		script: "RETURN 1",
		class:  NullDataTy,
	},
	{
		// Nulldata with max small integer push.
		name:   "nulldata max small int",
		script: "RETURN 16",
		class:  NullDataTy,
	},
	{
		// Nulldata with small data push.
		name:   "nulldata small data",
		script: "RETURN DATA_8 0x046708afdb0fe554",
		class:  NullDataTy,
	},
	{
		// Canonical nulldata with 60-byte data push.
		name: "canonical nulldata 60-byte push",
		script: "RETURN 0x3c 0x046708afdb0fe5548271967f1a67130b7105cd" +
			"6a828e03909a67962e0ea1f61deb649f6bc3f4cef3046708afdb" +
			"0fe5548271967f1a67130b7105cd6a",
		class: NullDataTy,
	},
	{
		// Non-canonical nulldata with 60-byte data push.
		name: "non-canonical nulldata 60-byte push",
		script: "RETURN PUSHDATA1 0x3c 0x046708afdb0fe5548271967f1a67" +
			"130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef3" +
			"046708afdb0fe5548271967f1a67130b7105cd6a",
		class: NullDataTy,
	},
	{
		// Nulldata with max allowed data to be considered standard.
		name: "nulldata max standard push",
		script: "RETURN PUSHDATA1 0x50 0x046708afdb0fe5548271967f1a67" +
			"130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef3" +
			"046708afdb0fe5548271967f1a67130b7105cd6a828e03909a67" +
			"962e0ea1f61deb649f6bc3f4cef3",
		class: NullDataTy,
	},
	{
		// Nulldata with more than max allowed data to be considered
		// standard (so therefore nonstandard)
		name: "nulldata exceed max standard push",
		script: "RETURN PUSHDATA2 0x1801 0x046708afdb0fe5548271967f1a670" +
			"46708afdb0fe5548271967f1a67046708afdb0fe5548271967f1a670467" +
			"08afdb0fe5548271967f1a67046708afdb0fe5548271967f1a67046708a" +
			"fdb0fe5548271967f1a67046708afdb0fe5548271967f1a67046708afdb" +
			"0fe5548271967f1a67046708afdb0fe5548271967f1a67046708afdb0fe" +
			"5548271967f1a67",
		class: NonStandardTy,
	},
	{
		// Almost nulldata, but add an additional opcode after the data
		// to make it nonstandard.
		name:   "almost nulldata",
		script: "RETURN 4 TRUE",
		class:  NonStandardTy,
	},

	// The next few are almost multisig (it is the more complex script type)
	// but with various changes to make it fail.
	{
		// Multisig but invalid nsigs.
		name: "strange 1",
		script: "DUP DATA_33 0x0232abdc893e7f0631364d7fd01cb33d24da45" +
			"329a00357b3a7886211ab414d55a 1 CHECKMULTISIG",
		class: NonStandardTy,
	},
	{
		// Multisig but invalid pubkey.
		name:   "strange 2",
		script: "1 1 1 CHECKMULTISIG",
		class:  NonStandardTy,
	},
	{
		// Multisig but no matching npubkeys opcode.
		name: "strange 3",
		script: "1 DATA_33 0x0232abdc893e7f0631364d7fd01cb33d24da4532" +
			"9a00357b3a7886211ab414d55a DATA_33 0x0232abdc893e7f0" +
			"631364d7fd01cb33d24da45329a00357b3a7886211ab414d55a " +
			"CHECKMULTISIG",
		class: NonStandardTy,
	},
	{
		// Multisig but with multisigverify.
		name: "strange 4",
		script: "1 DATA_33 0x0232abdc893e7f0631364d7fd01cb33d24da4532" +
			"9a00357b3a7886211ab414d55a 1 CHECKMULTISIGVERIFY",
		class: NonStandardTy,
	},
	{
		// Multisig but wrong length.
		name:   "strange 5",
		script: "1 CHECKMULTISIG",
		class:  NonStandardTy,
	},
	{
		name:   "doesn't parse",
		script: "DATA_5 0x01020304",
		class:  NonStandardTy,
	},
	{
		name: "multisig script with wrong number of pubkeys",
		script: "2 " +
			"DATA_33 " +
			"0x027adf5df7c965a2d46203c781bd4dd8" +
			"21f11844136f6673af7cc5a4a05cd29380 " +
			"DATA_33 " +
			"0x02c08f3de8ee2de9be7bd770f4c10eb0" +
			"d6ff1dd81ee96eedd3a9d4aeaf86695e80 " +
			"3 CHECKMULTISIG",
		class: NonStandardTy,
	},
}

// TestScriptClass ensures all the scripts in scriptClassTests have the expected
// class.
func TestScriptClass(t *testing.T) {
	t.Parallel()

	const scriptVersion = 0
	for _, test := range scriptClassTests {
		script := mustParseShortForm(test.script)
		class := GetScriptClass(scriptVersion, script, noTreasury)
		if class != test.class {
			t.Errorf("%s: expected %s got %s (script %x)", test.name,
				test.class, class, script)
			continue
		}
	}

	// Repeat tests with treasury.
	for _, test := range scriptClassTests {
		script := mustParseShortForm(test.script)
		class := GetScriptClass(scriptVersion, script, withTreasury)
		if class != test.class {
			t.Errorf("%s: expected %s got %s (script %x)", test.name,
				test.class, class, script)
			continue
		}
	}
}

// TestStringifyClass ensures the script class string returns the expected
// string for each script class.
func TestStringifyClass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		class    ScriptClass
		stringed string
	}{
		{
			name:     "nonstandardty",
			class:    NonStandardTy,
			stringed: "nonstandard",
		},
		{
			name:     "pubkey",
			class:    PubKeyTy,
			stringed: "pubkey",
		},
		{
			name:     "pubkeyhash",
			class:    PubKeyHashTy,
			stringed: "pubkeyhash",
		},
		{
			name:     "scripthash",
			class:    ScriptHashTy,
			stringed: "scripthash",
		},
		{
			name:     "multisigty",
			class:    MultiSigTy,
			stringed: "multisig",
		},
		{
			name:     "nulldataty",
			class:    NullDataTy,
			stringed: "nulldata",
		},
		{
			name:     "treasuryadd",
			class:    TreasuryAddTy,
			stringed: "treasuryadd",
		},
		{
			name:     "treasuryspend",
			class:    TreasurySpendTy,
			stringed: "treasuryspend",
		},
		{
			name:     "broken",
			class:    ScriptClass(255),
			stringed: "Invalid",
		},
	}

	for _, test := range tests {
		typeString := test.class.String()
		if typeString != test.stringed {
			t.Errorf("%s: got %#q, want %#q", test.name,
				typeString, test.stringed)
		}
	}
}

// TestGenerateProvablyPruneableOut tests whether GenerateProvablyPruneableOut returns a valid script.
func TestGenerateProvablyPruneableOut(t *testing.T) {
	const scriptVersion = 0
	tests := []struct {
		name     string
		data     []byte
		expected []byte
		err      error
		class    ScriptClass
	}{
		{
			name:     "small int",
			data:     hexToBytes("01"),
			expected: mustParseShortForm("RETURN 1"),
			err:      nil,
			class:    NullDataTy,
		},
		{
			name:     "max small int",
			data:     hexToBytes("10"),
			expected: mustParseShortForm("RETURN 16"),
			err:      nil,
			class:    NullDataTy,
		},
		{
			name: "data of size before OP_PUSHDATA1 is needed",
			data: hexToBytes("0102030405060708090a0b0c0d0e0f10111" +
				"2131415161718"),
			expected: mustParseShortForm("RETURN 0x18 0x01020304" +
				"05060708090a0b0c0d0e0f101112131415161718"),
			err:   nil,
			class: NullDataTy,
		},
		{
			name: "just right",
			data: hexToBytes("000102030405060708090a0b0c0d0e0f1011121" +
				"31415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f3" +
				"03132333435363738393a3b3c3d3e3f404142434445464748494a4b4c4" +
				"d4e4f202122232425262728292a2b2c2d2e2f303132333435363738393" +
				"a3b3c3d3e3f404142434445464748494a4b4c4d4e4f202122232425262" +
				"728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f404142434" +
				"445464748494a4b4c4d4e4f202122232425262728292a2b2c2d2e2f303" +
				"132333435363738393a3b3c3d3e3f404142434445464748494a4b4c4d4" +
				"e4f202122232425262728292a2b2c2d2e2f303132333435363738393a3" +
				"b3c3d3e"),
			expected: mustParseShortForm("RETURN PUSHDATA1 0xFF " +
				"0x000102030405060708090a0b0c0d0e0f101112131415161" +
				"718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f" +
				"303132333435363738393a3b3c3d3e3f40414243444546474" +
				"8494a4b4c4d4e4f202122232425262728292a2b2c2d2e2f30" +
				"3132333435363738393a3b3c3d3e3f4041424344454647484" +
				"94a4b4c4d4e4f202122232425262728292a2b2c2d2e2f3031" +
				"32333435363738393a3b3c3d3e3f404142434445464748494" +
				"a4b4c4d4e4f202122232425262728292a2b2c2d2e2f303132" +
				"333435363738393a3b3c3d3e3f404142434445464748494a4" +
				"b4c4d4e4f202122232425262728292a2b2c2d2e2f30313233" +
				"3435363738393a3b3c3d3e"),
			err:   nil,
			class: NullDataTy,
		},
		{
			name: "too big",
			data: hexToBytes("000102030405060708090a0b0c0d0e0f10111213141516" +
				"1718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f303132333435363" +
				"738393a3b3c3d3e3f404142434445464748494a4b4c4d4e4f2021222324252627" +
				"28292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f40414243444546474" +
				"8494a4b4c4d4e4f202122232425262728292a2b2c2d2e2f303132333435363738" +
				"393a3b3c3d3e3f404142434445464748494a4b4c4d4e4f2021222324252627282" +
				"92a2b2c2d2e2f303132333435363738393a3b3c3d3e3f40414243444546474849" +
				"4a4b4c4d4e4f202122232425262728292a2b2c2d2e2f303132333435363738393" +
				"a3b3c3d3e3f3f"),
			expected: nil,
			err:      ErrTooMuchNullData,
			class:    NonStandardTy,
		},
	}

	for i, test := range tests {
		script, err := GenerateProvablyPruneableOut(test.data)
		if !errors.Is(err, test.err) {
			t.Errorf("%s: unexpected error - got %v, want %v", test.name, err,
				test.err)
			continue
		}

		// Check that the expected result was returned.
		if !bytes.Equal(script, test.expected) {
			t.Errorf("GenerateProvablyPruneableOut: #%d (%s) wrong result\n"+
				"got: %x\nwant: %x", i, test.name, script,
				test.expected)
			continue
		}

		// Check that the script has the correct type.
		scriptType := GetScriptClass(scriptVersion, script, noTreasury)
		if scriptType != test.class {
			t.Errorf("GetScriptClass: #%d (%s) wrong result -- "+
				"got: %v, want: %v", i, test.name, scriptType,
				test.class)
			continue
		}

		// Check that the script has the correct type with treasury
		// agenda enabled.
		scriptType = GetScriptClass(scriptVersion, script, withTreasury)
		if scriptType != test.class {
			t.Errorf("GetScriptClass: #%d (%s) wrong result -- "+
				"got: %v, want: %v", i, test.name, scriptType,
				test.class)
			continue
		}
	}
}

// TestGenerateSStxAddrPush ensures an expected OP_RETURN push is generated.
func TestGenerateSStxAddrPush(t *testing.T) {
	testNetParams := chaincfg.TestNet3Params()
	var tests = []struct {
		addrStr  string
		net      dcrutil.AddressParams
		amount   dcrutil.Amount
		limits   uint16
		expected []byte
	}{
		{
			"Dcur2mcGjmENx4DhNqDctW5wJCVyT3Qeqkx",
			mainNetParams,
			1000,
			10,
			hexToBytes("6a1ef5916158e3e2c4551c1796708db8367207ed1" +
				"3bbe8030000000000800a00"),
		},
		{
			"TscB7V5RuR1oXpA364DFEsNDuAs8Rk6BHJE",
			testNetParams,
			543543,
			256,
			hexToBytes("6a1e7a5c4cca76f2e0b36db4763daacbd6cbb6ee6" +
				"e7b374b0800000000000001"),
		},
	}
	for _, test := range tests {
		addr, err := dcrutil.DecodeAddress(test.addrStr, test.net)
		if err != nil {
			t.Errorf("DecodeAddress failed: %v", err)
			continue
		}
		s, err := GenerateSStxAddrPush(addr, test.amount, test.limits)
		if err != nil {
			t.Errorf("GenerateSStxAddrPush failed: %v", err)
			continue
		}
		if !bytes.Equal(s, test.expected) {
			t.Errorf("GenerateSStxAddrPush: unexpected script:\n "+
				"got %x\nwant %x", s, test.expected)
		}
	}
}

// TestGenerateSSGenBlockRef ensures an expected OP_RETURN push is generated.
func TestGenerateSSGenBlockRef(t *testing.T) {
	var tests = []struct {
		blockHash string
		height    uint32
		expected  []byte
	}{
		{
			"0000000000004740ad140c86753f9295e09f9cc81b1bb75d7f5552aeeedb7012",
			1000,
			hexToBytes("6a241270dbeeae52557f5db71b1bc89c9fe095923" +
				"f75860c14ad4047000000000000e8030000"),
		},
		{
			"000000000000000033eafc268a67c8d1f02343d7a96cf3fe2a4915ef779b52f9",
			290000,
			hexToBytes("6a24f9529b77ef15492afef36ca9d74323f0d1c86" +
				"78a26fcea330000000000000000d06c0400"),
		},
	}
	for _, test := range tests {
		h, err := chainhash.NewHashFromStr(test.blockHash)
		if err != nil {
			t.Errorf("NewHashFromStr failed: %v", err)
			continue
		}
		s, err := GenerateSSGenBlockRef(*h, test.height)
		if err != nil {
			t.Errorf("GenerateSSGenBlockRef failed: %v", err)
			continue
		}
		if !bytes.Equal(s, test.expected) {
			t.Errorf("GenerateSSGenBlockRef: unexpected script:\n"+
				" got %x\nwant %x", s, test.expected)
		}
	}
}

// TestGenerateSSGenVotes ensures an expected OP_RETURN push is generated.
func TestGenerateSSGenVotes(t *testing.T) {
	var tests = []struct {
		votebits uint16
		expected []byte
	}{
		{65535, hexToBytes("6a02ffff")},
		{256, hexToBytes("6a020001")},
		{127, hexToBytes("6a027f00")},
		{0, hexToBytes("6a020000")},
	}
	for _, test := range tests {
		s, err := GenerateSSGenVotes(test.votebits)
		if err != nil {
			t.Errorf("GenerateSSGenVotes failed: %v", err)
			continue
		}
		if !bytes.Equal(s, test.expected) {
			t.Errorf("GenerateSSGenVotes: unexpected script:\n "+
				"got %x\nwant %x", s, test.expected)
		}
	}
}

// mustExpectedAtomicSwapData is a convenience function that converts the passed
// parameters into an expected atomic swap data pushes structure and will panic
// if there is an error.  This is only provided for the hard-coded constants so
// errors in the source code can be detected. It will only (and must only) be
// called with hard-coded values.
func mustExpectedAtomicSwapData(recipientHash, refundHash, secretHash string, secretSize, lockTime int64) *AtomicSwapDataPushes {
	result := &AtomicSwapDataPushes{
		SecretSize: secretSize,
		LockTime:   lockTime,
	}
	copy(result.RecipientHash160[:], hexToBytes(recipientHash))
	copy(result.RefundHash160[:], hexToBytes(refundHash))
	copy(result.SecretHash[:], hexToBytes(secretHash))
	return result
}

// TestExtractAtomicSwapDataPushes ensures atomic swap scripts are recognized
// properly and the correct information is extracted from them.
func TestExtractAtomicSwapDataPushes(t *testing.T) {
	// Define some values shared in the tests for convenience.
	secret := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	recipient := "0000000000000000000000000000000000000001"
	refund := "0000000000000000000000000000000000000002"

	tests := []struct {
		name          string                // test description
		scriptVersion uint16                // version of script to analyze
		script        string                // script to analyze
		data          *AtomicSwapDataPushes // expected data pushes
		err           error                 // expected error
	}{{
		name: "normal valid atomic swap",
		script: fmt.Sprintf("IF SIZE 32 EQUALVERIFY SHA256 DATA_32 "+
			"0x%s EQUALVERIFY DUP HASH160 DATA_20 0x%s ELSE 300000 "+
			"CHECKLOCKTIMEVERIFY DROP DUP HASH160 DATA_20 0x%s ENDIF "+
			"EQUALVERIFY CHECKSIG", secret, recipient, refund),
		scriptVersion: 0,
		data: mustExpectedAtomicSwapData(recipient, refund, secret, 32,
			300000),
		err: nil,
	}, {
		name: "atomic swap with mismatched smallint secret size",
		script: fmt.Sprintf("IF SIZE 16 EQUALVERIFY SHA256 DATA_32 "+
			"0x%s EQUALVERIFY DUP HASH160 DATA_20 0x%s ELSE 300000 "+
			"CHECKLOCKTIMEVERIFY DROP DUP HASH160 DATA_20 0x%s ENDIF "+
			"EQUALVERIFY CHECKSIG", secret, recipient, refund),
		scriptVersion: 0,
		data: mustExpectedAtomicSwapData(recipient, refund, secret, 16,
			300000),
		err: nil,
	}, {
		name: "atomic swap with smallint locktime",
		script: fmt.Sprintf("IF SIZE 32 EQUALVERIFY SHA256 DATA_32 "+
			"0x%s EQUALVERIFY DUP HASH160 DATA_20 0x%s ELSE 10 "+
			"CHECKLOCKTIMEVERIFY DROP DUP HASH160 DATA_20 0x%s ENDIF "+
			"EQUALVERIFY CHECKSIG", secret, recipient, refund),
		scriptVersion: 0,
		data: mustExpectedAtomicSwapData(recipient, refund, secret, 32,
			10),
		err: nil,
	}, {
		name: "almost valid, but NOP for secret size",
		script: fmt.Sprintf("IF SIZE NOP EQUALVERIFY SHA256 DATA_32 "+
			"0x%s EQUALVERIFY DUP HASH160 DATA_20 0x%s ELSE 300000 "+
			"CHECKLOCKTIMEVERIFY DROP DUP HASH160 DATA_20 0x%s ENDIF "+
			"EQUALVERIFY CHECKSIG", secret, recipient, refund),
		scriptVersion: 0,
		data:          nil,
		err:           nil,
	}, {
		name: "almost valid, but NOP for locktime",
		script: fmt.Sprintf("IF SIZE 32 EQUALVERIFY SHA256 DATA_32 "+
			"0x%s EQUALVERIFY DUP HASH160 DATA_20 0x%s ELSE NOP "+
			"CHECKLOCKTIMEVERIFY DROP DUP HASH160 DATA_20 0x%s ENDIF "+
			"EQUALVERIFY CHECKSIG", secret, recipient, refund),
		scriptVersion: 0,
		data:          nil,
		err:           nil,
	}, {
		name: "almost valid, but wrong sha256 secret size",
		script: fmt.Sprintf("IF SIZE 32 EQUALVERIFY SHA256 DATA_31 "+
			"0x%s EQUALVERIFY DUP HASH160 DATA_20 0x%s ELSE 300000 "+
			"CHECKLOCKTIMEVERIFY DROP DUP HASH160 DATA_20 0x%s ENDIF "+
			"EQUALVERIFY CHECKSIG", secret[:len(secret)-2], recipient, refund),
		scriptVersion: 0,
		data:          nil,
		err:           nil,
	}, {
		name: "almost valid, but wrong recipient hash size",
		script: fmt.Sprintf("IF SIZE 32 EQUALVERIFY SHA256 DATA_32 "+
			"0x%s EQUALVERIFY DUP HASH160 DATA_19 0x%s ELSE 300000 "+
			"CHECKLOCKTIMEVERIFY DROP DUP HASH160 DATA_20 0x%s ENDIF "+
			"EQUALVERIFY CHECKSIG", secret, recipient[:len(recipient)-2],
			refund),
		scriptVersion: 0,
		data:          nil,
		err:           nil,
	}, {
		name: "almost valid, but wrong refund hash size",
		script: fmt.Sprintf("IF SIZE 32 EQUALVERIFY SHA256 DATA_32 "+
			"0x%s EQUALVERIFY DUP HASH160 DATA_20 0x%s ELSE 300000 "+
			"CHECKLOCKTIMEVERIFY DROP DUP HASH160 DATA_19 0x%s ENDIF "+
			"EQUALVERIFY CHECKSIG", secret, recipient, refund[:len(refund)-2]),
		scriptVersion: 0,
		data:          nil,
		err:           nil,
	}, {
		name: "almost valid, but missing final CHECKSIG",
		script: fmt.Sprintf("IF SIZE 32 EQUALVERIFY SHA256 DATA_32 "+
			"0x%s EQUALVERIFY DUP HASH160 DATA_20 0x%s ELSE 300000 "+
			"CHECKLOCKTIMEVERIFY DROP DUP HASH160 DATA_20 0x%s ENDIF "+
			"EQUALVERIFY", secret, recipient, refund),
		scriptVersion: 0,
		data:          nil,
		err:           nil,
	}, {
		name: "almost valid, but additional opcode at end",
		script: fmt.Sprintf("IF SIZE 32 EQUALVERIFY SHA256 DATA_32 "+
			"0x%s EQUALVERIFY DUP HASH160 DATA_20 0x%s ELSE 300000 "+
			"CHECKLOCKTIMEVERIFY DROP DUP HASH160 DATA_20 0x%s ENDIF "+
			"EQUALVERIFY CHECKSIG NOP", secret, recipient, refund),
		scriptVersion: 0,
		data:          nil,
		err:           nil,
	}, {
		name: "valid atomic swap for v0 script, but unsupported version",
		script: fmt.Sprintf("IF SIZE 32 EQUALVERIFY SHA256 DATA_32 "+
			"0x%s EQUALVERIFY DUP HASH160 DATA_20 0x%s ELSE 300000 "+
			"CHECKLOCKTIMEVERIFY DROP DUP HASH160 DATA_20 0x%s ENDIF "+
			"EQUALVERIFY CHECKSIG", secret, recipient, refund),
		scriptVersion: 65535,
		data:          nil,
		err:           ErrUnsupportedScriptVersion,
	}}

	for _, test := range tests {
		script := mustParseShortForm(test.script)

		// Attempt to extract the atomic swap data from the script and ensure
		// the error is as expected.
		data, err := ExtractAtomicSwapDataPushes(test.scriptVersion, script)
		if !errors.Is(err, test.err) {
			t.Fatalf("%q: unexpected err -- got %v, want nil", test.name, err)
		}
		if test.err != nil {
			continue
		}

		// Ensure there is either extract data or not as expected.
		switch {
		case test.data == nil && data != nil:
			t.Fatalf("%q: unexpected extracted data", test.name)

		case test.data != nil && data == nil:
			t.Fatalf("%q: failed to extract expected data", test.name)

		case data == nil:
			continue
		}

		// Ensure the individual fields of the extracted data is accurate.  The
		// two structs could be directly compared, but testing them individually
		// allows nicer error reporting in the case of failure.
		if data.RecipientHash160 != test.data.RecipientHash160 {
			t.Fatalf("%q: unexpected recipient hash -- got %x, want %x",
				test.name, data.RecipientHash160, test.data.RecipientHash160)
		}
		if data.RefundHash160 != test.data.RefundHash160 {
			t.Fatalf("%q: unexpected refund hash -- got %x, want %x", test.name,
				data.RefundHash160, test.data.RefundHash160)
		}
		if data.SecretHash != test.data.SecretHash {
			t.Fatalf("%q: unexpected secret hash -- got %x, want %x", test.name,
				data.SecretHash, test.data.SecretHash)
		}
		if data.SecretSize != test.data.SecretSize {
			t.Fatalf("%q: unexpected secret size -- got %d, want %d", test.name,
				data.SecretSize, test.data.SecretSize)
		}
		if data.LockTime != test.data.LockTime {
			t.Fatalf("%q: unexpected locktime -- got %d, want %d", test.name,
				data.LockTime, test.data.LockTime)
		}
	}
}
