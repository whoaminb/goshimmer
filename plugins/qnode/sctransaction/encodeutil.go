package sctransaction

import (
	"errors"
	"github.com/iotaledger/goshimmer/plugins/qnode/util"
	"hash/crc32"
)

// the byte is needed for the parses to quickly recognize
// what kind of block it is: state or request
// max number of request blocks in the transaction is 127

const stateBlockMask = 0x80

func encodeMetaByte(hasState bool, numRequests byte) (byte, error) {
	if numRequests > 127 {
		return 0, errors.New("can't be more than 127 requests")
	}
	ret := numRequests
	if hasState {
		ret = ret | stateBlockMask
	}
	return ret, nil
}

func decodeMetaByte(b byte) (bool, byte) {
	return b|stateBlockMask != 0, b & stateBlockMask
}

func mustChecksum65Bytes(data []byte) uint32 {
	if len(data) != 65 {
		panic("mustChecksum65Bytes: wrong param")
	}
	return crc32.ChecksumIEEE(data)
}

// recognizes if the payload can be a parsed as SC payload, without parsing
// it must start with 1 meta byte, checksum and scid.
func CheckScPayloadPrefix(data []byte) bool {
	// 1 for meta byte
	// 4 for checksum
	// 65 for scid
	// minimum sc data payload is 71 bytes
	if len(data) < 1+4+ScIdLength {
		return false
	}
	checksumGiven := util.Uint32From4Bytes(data[1:5])
	// skip 2 bytes
	checksumCalculated := crc32.ChecksumIEEE(data[1+4 : 1+4+ScIdLength])
	return checksumGiven == checksumCalculated
}
