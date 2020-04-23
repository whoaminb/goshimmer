package sctransaction

import (
	"errors"
	valuetransaction "github.com/iotaledger/goshimmer/packages/binary/valuetransfer/transaction"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/util"
	"io"
)

const RequestIdSize = hashing.HashSize + 2

type RequestId [RequestIdSize]byte

type RequestRef struct {
	reqId *RequestId
	tx    Transaction
}

type RequestBlock struct {
	scid *ScId
	body *RequestBody
}

// RequestBlock

func NewRequestBlock(scid *ScId) *RequestBlock {
	return &RequestBlock{
		scid: scid,
	}
}

func (req *RequestBlock) ScId() *ScId {
	return req.ScId()
}

// encoding
// important: each block starts with 65 bytes of scid

func (req *RequestBlock) Write(w io.Writer) error {
	if err := req.scid.Write(w); err != nil {
		return err
	}
	// TODO write body
	return nil
}

func (req *RequestBlock) Read(r io.Reader) error {
	scid := new(ScId)
	if err := scid.Read(r); err != nil {
		return err
	}
	// TODO read body
	req.scid = scid
	return nil
}

// TODO the rest of request body

// Request Id

func (rid *RequestId) Bytes() []byte {
	return rid[:]
}

func (rid *RequestId) TransactionId() *valuetransaction.Id {
	var ret valuetransaction.Id
	copy(ret[:], rid[:valuetransaction.IdLength])
	return &ret
}

func (rid *RequestId) Index() uint16 {
	return util.Uint16From2Bytes(rid[valuetransaction.IdLength:])
}

func (rid *RequestId) Write(w io.Writer) error {
	_, err := w.Write(rid.Bytes())
	return err
}

func (rid *RequestId) Read(r io.Reader) error {
	n, err := r.Read(rid.Bytes())
	if err != nil {
		return err
	}
	if n != RequestIdSize {
		return errors.New("not enough data for RequestId")
	}
	return nil
}
