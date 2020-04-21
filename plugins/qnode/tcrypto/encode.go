package tcrypto

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/util"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"io"
)

func (ks *DKShare) Write(w io.Writer) error {
	err := util.WriteUint16(w, ks.N)
	if err != nil {
		return err
	}
	err = util.WriteUint16(w, ks.T)
	if err != nil {
		return err
	}
	err = util.WriteUint16(w, ks.Index)
	if err != nil {
		return err
	}
	_, err = w.Write(ks.Address.Bytes())
	if err != nil {
		return err
	}
	err = util.WriteUint64(w, uint64(ks.Created))
	if err != nil {
		return err
	}
	err = util.WriteUint16(w, uint16(len(ks.PubKeys)))
	if err != nil {
		return err
	}
	for _, pk := range ks.PubKeys {
		pkdata, err := pk.MarshalBinary()
		if err != nil {
			return err
		}
		err = util.WriteBytes16(w, pkdata)
		if err != nil {
			return err
		}
	}
	pkdata, err := ks.PriKey.MarshalBinary()
	if err != nil {
		return err
	}
	err = util.WriteBytes16(w, pkdata)
	if err != nil {
		return err
	}
	return nil
}

func (ks *DKShare) Read(r io.Reader) error {
	*ks = DKShare{}
	ks.Suite = bn256.NewSuite()

	var n, t, index uint16
	err := util.ReadUint16(r, &n)
	if err != nil {
		return err
	}
	err = util.ReadUint16(r, &t)
	if err != nil {
		return err
	}
	err = util.ReadUint16(r, &index)
	if err != nil {
		return err
	}
	var account hashing.HashValue
	_, err = r.Read(account.Bytes())
	if err != nil {
		return err
	}
	var created uint64
	err = util.ReadUint64(r, &created)
	if err != nil {
		return err
	}
	var num uint16
	err = util.ReadUint16(r, &num)
	if err != nil {
		return err
	}
	pubKeys := make([]kyber.Point, num)
	for i := range pubKeys {
		data, err := util.ReadBytes16(r)
		if err != nil {
			return err
		}
		pubKeys[i] = ks.Suite.G2().Point()
		err = pubKeys[i].UnmarshalBinary(data)
		if err != nil {
			return err
		}
	}
	data, err := util.ReadBytes16(r)
	if err != nil {
		return err
	}
	priKey := ks.Suite.G2().Scalar()
	err = priKey.UnmarshalBinary(data)
	if err != nil {
		return err
	}
	ks.N = n
	ks.T = t
	ks.Index = index
	ks.Address = &account
	ks.Created = int64(created)
	ks.PubKeys = pubKeys
	ks.PriKey = priKey
	return nil
}
