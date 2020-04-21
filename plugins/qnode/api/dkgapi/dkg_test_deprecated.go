package dkgapi

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto/tbdn"
	"github.com/pkg/errors"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"testing"
)

const (
	N = 5
	T = 4
)

func ___TestDKGGeneral(t *testing.T) {
	setId := HashStrings("KLMN12345")

	dicts := make([]map[HashValue]*tcrypto.DKShare, N)
	for i := range dicts {
		dicts[i] = make(map[HashValue]*tcrypto.DKShare)
	}
	resps := make([]*NewDKSResponse, N)
	errb := false
	for i := range dicts {
		resps[i] = NewDKSetReq(&NewDKSRequest{
			Id:    setId,
			N:     N,
			T:     T,
			Index: uint16(i),
		})
		if resps[i].Err != "" {
			errb = true
		}
		js, _ := json.MarshalIndent(resps[i], "", " ")
		fmt.Printf("%s\n", js)
	}
	if errb {
		fmt.Println("errors!")
		return
	}
	aggrReqs := make([]*AggregateDKSRequest, N)
	for i := range aggrReqs {
		aggrReqs[i] = &AggregateDKSRequest{
			Id:        setId,
			PriShares: make([]string, N),
		}
	}

	//for i, r := range resps {
	//	for j := range aggrReqs {
	//		if j == i {
	//			aggrReqs[j].PriShares[r.Index] = ""
	//		} else {
	//			aggrReqs[j].PriShares[r.Index] = r.PriShares[j]
	//		}
	//	}
	//}
	// print it
	for j := range aggrReqs {
		js, _ := json.MarshalIndent(aggrReqs[j], "", " ")
		fmt.Printf("%s\n", js)
	}
	// call aggregate
	errb = false
	aggrResp := make([]*AggregateDKSResponse, N)
	for j, aggrReq := range aggrReqs {
		aggrResp[j] = AggregateDKSReq(aggrReq)
		if aggrResp[j].Err != "" {
			errb = true
		}
		fmt.Printf("%+v\n", *aggrResp[j])
	}
	if errb {
		fmt.Println("errors!")
		return
	}

	pubKeysS := make([]string, N)
	for j, r := range aggrResp {
		pubKeysS[j] = r.PubShare
	}
	suite := bn256.NewSuite()
	pubKeys, err := DecodePubKeys(suite, pubKeysS)
	if err != nil {
		panic(err)
	}

	pubShares := make([]*share.PubShare, len(pubKeys))
	for i, v := range pubKeys {
		pubShares[i] = &share.PubShare{
			I: i,
			V: v,
		}
	}
	pubPoly, err := share.RecoverPubPoly(suite.G2(), pubShares, T, N)
	if err != nil {
		panic(err)
	}
	// sign

	fmt.Println("----------------------------")
	msgs := "Hello, world"
	msgb := []byte(msgs)
	sigs := make([][]byte, N)
	for i, d := range dicts {
		ks := d[*setId]
		sigs[i], err = ks.SignShare(msgb)
		fmt.Printf("sig share: %s len = %d\n", hex.EncodeToString(sigs[i]), len(sigs[i]))

		err := tbdn.Verify(suite, pubPoly, msgb, sigs[i])
		if err != nil {
			fmt.Printf("sigShare verification #%d: %v\n", i, err)
		} else {
			fmt.Printf("sigShare verification: ok\n")
		}
	}

	ok, err := checkCombi(suite, []int{0, 1, 2, 3, 4}, sigs, pubPoly, msgb)
	expectResult(t, ok, err, true)
	ok, err = checkCombi(suite, []int{1, 2, 3, 4}, sigs, pubPoly, msgb)
	expectResult(t, ok, err, true)
	ok, err = checkCombi(suite, []int{0, 1, 2, 3}, sigs, pubPoly, msgb)
	expectResult(t, ok, err, true)
	ok, err = checkCombi(suite, []int{0, 1, 2}, sigs, pubPoly, msgb)
	expectResult(t, ok, err, false)

	// TODO add commit testing
}

func expectResult(t *testing.T, ok bool, err error, expect bool) {
	if err != nil {
		t.Error(err)
		return
	}
	if ok != expect {
		t.Errorf("failed signature verification: expected %v, result %v", expect, ok)
	}
}

func checkCombi(suite *bn256.Suite, indices []int, allSigs [][]byte, pubPoly *share.PubPoly, msg []byte) (bool, error) {
	fmt.Printf("checkCombi %+v", indices)
	ss, err := makeSubset(indices, allSigs)
	if err != nil {
		fmt.Printf("checkCombi: %v\n", err)
		return false, err
	}
	sig, err := tbdn.Recover(suite, pubPoly, msg, ss, T, N)
	if err != nil {
		fmt.Printf("sigShare Recover %v\n", err)
	} else {
		fmt.Printf("Recovered signature %s\n", hex.EncodeToString(sig))
	}

	pubKey := pubPoly.Commit()
	pkb, _ := pubKey.MarshalBinary()
	fmt.Printf("++++++++++++++++++ Public key length: %d bytes, signature length: %d bytes\n", len(pkb), len(sig))

	err = bdn.Verify(suite, pubKey, msg, sig)
	if err != nil {
		fmt.Printf("Recovered signature verification: %v\n", err)
	} else {
		fmt.Printf("Recovered signature verification: ok\n")
	}
	return err == nil, nil
}

func makeSubset(indices []int, allSigs [][]byte) ([][]byte, error) {
	ret := make([][]byte, 0)
	for _, idx := range indices {
		if idx < 0 || idx >= len(allSigs) {
			return nil, errors.New("wrong index")
		}
		ret = append(ret, allSigs[idx])
	}
	return ret, nil
}

func VerifyCombi(suite *bn256.Suite, t int, n int, shares []*share.PriShare, pubKey kyber.Point, msg []byte) {
	priPoly, err := share.RecoverPriPoly(suite.G2(), shares, t, n)
	if err != nil {
		fmt.Printf("RecoverPriPoly: %v\n", err)
	}
	recSig := priPoly.Secret()
	recSigB, err := recSig.MarshalBinary()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	} else {
		fmt.Printf("Recovered signature: %s", hex.EncodeToString(recSigB))
	}
}

func DecodePubKeys(suite *bn256.Suite, pubKeysS []string) ([]kyber.Point, error) {
	ret := make([]kyber.Point, len(pubKeysS))
	for i, s := range pubKeysS {
		b, err := hex.DecodeString(s)
		if err != nil {
			return nil, err
		}
		p := suite.G2().Point()
		if err := p.UnmarshalBinary(b); err != nil {
			return nil, err
		}
		ret[i] = p
	}
	return ret, nil
}
