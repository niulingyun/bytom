package pseudohsm

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/bytom/crypto/sm2"
	"github.com/bytom/errors"
)

const dirPath = "testdata/pseudo"

func TestCreateKeyWithUpperCase(t *testing.T) {
	hsm, _ := New(dirPath)

	alias := "UPPER"

	xpub, _, err := hsm.XCreate(alias, "password")
	if err != nil {
		t.Fatal(err)
	}

	if xpub.Alias != strings.ToLower(alias) {
		t.Fatal("the created key alias should be lowercase")
	}

	err = hsm.XDelete(xpub.XPub, "password")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateKeyWithWhiteSpaceTrimed(t *testing.T) {
	hsm, _ := New(dirPath)

	alias := " with space surrounding "

	xpub, _, err := hsm.XCreate(alias, "password")
	if err != nil {
		t.Fatal(err)
	}

	if xpub.Alias != strings.TrimSpace(alias) {
		t.Fatal("the created key alias should be lowercase")
	}

	err = hsm.XDelete(xpub.XPub, "password")
	if err != nil {
		t.Fatal(err)
	}
}

func TestPseudoHSMChainKDKeys(t *testing.T) {

	hsm, _ := New(dirPath)
	xpub, _, err := hsm.XCreate("bbs", "password")

	if err != nil {
		t.Fatal(err)
	}
	xpub2, _, err := hsm.XCreate("bytom", "nopassword")
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("In the face of ignorance and resistance I wrote financial systems into existence")
	sig, err := hsm.XSign(xpub.XPub, nil, msg, "password")
	if err != nil {
		t.Fatal(err)
	}
	if !xpub.XPub.Verify(msg, sig) {
		t.Error("expected verify to succeed")
	}
	if xpub2.XPub.Verify(msg, sig) {
		t.Error("expected verify with wrong pubkey to fail")
	}
	path := [][]byte{{3, 2, 6, 3, 8, 2, 7}}
	sig, err = hsm.XSign(xpub2.XPub, path, msg, "nopassword")
	if err != nil {
		t.Fatal(err)
	}
	if xpub2.XPub.Verify(msg, sig) {
		t.Error("expected verify with underived pubkey of sig from derived privkey to fail")
	}
	if !xpub2.XPub.Derive(path).Verify(msg, sig) {
		t.Error("expected verify with derived pubkey of sig from derived privkey to succeed")
	}

	xpubs := hsm.ListKeys()
	if len(xpubs) != 2 {
		t.Error("expected 2 entries in the db")
	}
	err = hsm.ResetPassword(xpub2.XPub, "nopassword", "1password")
	if err != nil {
		t.Fatal(err)
	}
	err = hsm.XDelete(xpub.XPub, "password")
	if err != nil {
		t.Fatal(err)
	}
	err = hsm.XDelete(xpub2.XPub, "1password")
	if err != nil {
		t.Fatal(err)
	}
}

func TestKeyWithEmptyAlias(t *testing.T) {
	hsm, _ := New(dirPath)
	for i := 0; i < 2; i++ {
		xpub, _, err := hsm.XCreate(fmt.Sprintf("xx%d", i), "xx")
		if errors.Root(err) != nil {
			t.Fatal(err)
		}
		err = hsm.XDelete(xpub.XPub, "xx")
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestSignAndVerifyMessage(t *testing.T) {
	hsm, _ := New(dirPath)
	xpub, _, err := hsm.XCreate("TESTKEY", "password")
	if err != nil {
		t.Fatal(err)
	}

	path := [][]byte{{3, 2, 6, 3, 8, 2, 7}}
	derivedXPub := xpub.XPub.Derive(path)

	msg := "this is a test message"
	sig, err := hsm.XSign(xpub.XPub, path, []byte(msg), "password")
	if err != nil {
		t.Fatal(err)
	}

	// derivedXPub verify success
	if !sm2.VerifyCompressedPubkey(derivedXPub.PublicKey(), []byte(msg), sig) {
		t.Fatal("right derivedXPub verify sign failed")
	}

	// rootXPub verify failed
	if sm2.VerifyCompressedPubkey(xpub.XPub.PublicKey(), []byte(msg), sig) {
		t.Fatal("right rootXPub verify derivedXPub sign succeed")
	}

	err = hsm.XDelete(xpub.XPub, "password")
	if err != nil {
		t.Fatal(err)
	}
}

func TestImportKeyFromMnemonic(t *testing.T) {
	dirPath, err := ioutil.TempDir(".", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirPath)

	hsm, _ := New(dirPath)
	xpub, mnemonic, err := hsm.XCreate("TESTKEY", "password")
	if err != nil {
		t.Fatal(err)
	}
	newXPub, err := hsm.ImportKeyFromMnemonic("TESTKEY1", "password", *mnemonic)
	if err != nil {
		t.Fatal(err)
	}
	if xpub.XPub != newXPub.XPub {
		t.Fatal("import key from mnemonic failed")
	}
}

func BenchmarkSign(b *testing.B) {
	b.StopTimer()
	auth := "nowpasswd"

	hsm, _ := New(dirPath)
	xpub, _, err := hsm.XCreate("TESTKEY", auth)
	if err != nil {
		b.Fatal(err)
	}

	msg := []byte("In the face of ignorance and resistance I wrote financial systems into existence")

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err := hsm.XSign(xpub.XPub, nil, msg, auth)
		if err != nil {
			b.Fatal(err)
		}
	}
	err = hsm.XDelete(xpub.XPub, auth)
	if err != nil {
		b.Fatal(err)
	}
}
