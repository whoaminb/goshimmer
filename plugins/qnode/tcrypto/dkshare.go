package tcrypto

import (
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto/tbdn"
	"github.com/pkg/errors"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"time"
)

// Distributed key set for (T,N) threshold signatures, T out f N

type DKShare struct {
	Suite   *bn256.Suite
	N       uint16
	T       uint16
	Index   uint16
	Address *HashValue // used as permanent id = hash(pubkey)

	Created    int64
	Aggregated bool
	Committed  bool
	//
	PriShares []*share.PriShare // nil after commit
	//
	PubKeys      []kyber.Point // all public shares by peers
	PubPoly      *share.PubPoly
	PriKey       kyber.Scalar // own private key (sum of private shares)
	PubKeyOwn    kyber.Point  // public key from own private key
	PubKeyMaster kyber.Point
}

func ValidateDKSParams(t, n, index uint16) error {
	if t < 1 || t > n || index < 0 || index >= n {
		return errors.New("wrong N, T or Index parameters")
	}
	// probably not necessary
	if t < (2*n)/3+1 {
		return errors.New(fmt.Sprintf("value T must be at least floor(2*N/3)+1"))
	}
	return nil
}

func NewRndDKShare(t, n, index uint16) (*DKShare, error) {
	if err := ValidateDKSParams(t, n, index); err != nil {
		return nil, err
	}
	suite := bn256.NewSuite()
	// create seed secret
	secret := suite.G1().Scalar().Pick(suite.RandomStream())
	// create random polynomial of degree t
	priPoly := share.NewPriPoly(suite.G2(), int(t), secret, suite.RandomStream())
	// create private shares of the random polynomial
	// with index n corresponds to p(n+1)
	shares := priPoly.Shares(int(n))
	ret := &DKShare{
		Suite:     suite,
		Created:   time.Now().UnixNano(),
		N:         n,
		T:         t,
		Index:     index,
		PriShares: shares,
	}
	return ret, nil
}

func (ks *DKShare) AggregateDKS(priShares []kyber.Scalar) error {
	if ks.Aggregated {
		return errors.New("already Aggregated")
	}
	// aggregate (add up) secret shares
	ks.PriKey = ks.Suite.G2().Scalar().Zero()
	for i, pshare := range priShares {
		if uint16(i) == ks.Index {
			ks.PriKey = ks.PriKey.Add(ks.PriKey, ks.PriShares[ks.Index].V)
			continue
		}
		ks.PriKey = ks.PriKey.Add(ks.PriKey, pshare)
	}
	// calculate own public key
	ks.PubKeyOwn = ks.Suite.G2().Point().Mul(ks.PriKey, nil)
	ks.Aggregated = true
	return nil
}

func (ks *DKShare) FinalizeDKS(pubKeys []kyber.Point) error {
	if ks.Committed {
		return errors.New("already Committed")
	}
	ks.PubKeys = pubKeys
	var err error
	ks.PubPoly, err = RecoverPubPoly(ks.Suite, ks.PubKeys, ks.T, ks.N)
	if err != nil {
		return err
	}
	pubKeyMaster := ks.PubPoly.Commit()
	pubKeyBin, err := pubKeyMaster.MarshalBinary()
	if err != nil {
		return err
	}
	// calculate address, the permanent key ID
	ks.Address = HashData(pubKeyBin)

	ks.PriShares = nil // not needed anymore
	ks.Committed = true
	return nil
}

func (ks *DKShare) SignShare(data []byte) (tbdn.SigShare, error) {
	priShare := share.PriShare{
		I: int(ks.Index),
		V: ks.PriKey,
	}
	return tbdn.Sign(ks.Suite, &priShare, data)
}

// for testing
func (ks *DKShare) VerifyOwnSigShare(data []byte, sigshare tbdn.SigShare) error {
	if !ks.Committed {
		return errors.New("key set is not Committed")
	}
	idx, err := sigshare.Index()
	if err != nil || uint16(idx) != ks.Index {
		return err
	}
	return bdn.Verify(ks.Suite, ks.PubKeyOwn, data, sigshare[2:])
}

func (ks *DKShare) VerifySigShare(data []byte, sigshare tbdn.SigShare) error {
	if !ks.Committed {
		return errors.New("key set is not Committed")
	}
	idx, err := sigshare.Index()
	if err != nil || idx >= int(ks.N) || idx < 0 {
		return err
	}
	return bdn.Verify(ks.Suite, ks.PubKeys[idx], data, sigshare.Value())
}

func (ks *DKShare) VerifyMasterSignature(data []byte, signature []byte) error {
	if !ks.Committed {
		return errors.New("key set is not Committed")
	}
	return bdn.Verify(ks.Suite, ks.PubKeyMaster, data, signature)
}

var suiteLoc *bn256.Suite

func init() {
	suiteLoc = bn256.NewSuite()
}

func VerifyWithPublicKey(data, signature, pubKeyBin []byte) error {
	pubKey := suiteLoc.G2().Point()
	err := pubKey.UnmarshalBinary(pubKeyBin)
	if err != nil {
		return err
	}
	return bdn.Verify(suiteLoc, pubKey, data, signature)
}

func RecoverPubPoly(suite *bn256.Suite, pubKeys []kyber.Point, t, n uint16) (*share.PubPoly, error) {
	pubShares := make([]*share.PubShare, len(pubKeys))
	for i, v := range pubKeys {
		pubShares[i] = &share.PubShare{
			I: i,
			V: v,
		}
	}
	return share.RecoverPubPoly(suite.G2(), pubShares, int(t), int(n))
}
