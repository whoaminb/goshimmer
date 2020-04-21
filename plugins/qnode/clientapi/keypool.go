package clientapi

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/model/generic"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto"
	"github.com/rogpeppe/go-internal/cache"
)

type mockKeysPool struct {
}

// for testing

func NewDummyKeyPool() generic.KeyPool {
	return &mockKeysPool{}
}

func (kp *mockKeysPool) SignBlock(sigBlk generic.SignedBlock) error {
	var sig [128]byte
	h := hashing.HashData(sigBlk.SignedHash().Bytes(), sigBlk.Account().Bytes())
	copy(sig[:], h.Bytes())
	sigBlk.SetSignature(sig[:], generic.SIG_TYPE_MOCKED)
	return nil
}

func (kp *mockKeysPool) VerifySignature(sigBlk generic.SignedBlock) error {
	sig, typ := sigBlk.GetSignature()
	var err error
	switch typ {
	case generic.SIG_TYPE_BLS_FINAL:
		err = tcrypto.VerifyWithPublicKey(sigBlk.SignedHash().Bytes(), sig, sigBlk.GetPublicKey())

	case generic.SIG_TYPE_MOCKED:
		h := hashing.HashData(sigBlk.SignedHash().Bytes(), sigBlk.Account().Bytes())
		if bytes.Compare(sig[:cache.HashSize], h.Bytes()) != 0 {
			err = fmt.Errorf("invalid mocked signature")
		}
	default:
		err = fmt.Errorf("not suported signature type")
	}
	return err
}

func (kp *mockKeysPool) GetKeyData(_ *hashing.HashValue) (interface{}, error) {
	return nil, nil
}
