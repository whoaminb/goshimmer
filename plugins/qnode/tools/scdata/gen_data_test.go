package main

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"testing"
)

func TestGenData(t *testing.T) {
	t.Logf("dummy owner's pub key: %s", hashing.HashStrings("dummy pub key").String())
	t.Logf("dummy program hash: %s", hashing.HashStrings("dummy program").String())
}
