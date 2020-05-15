package waspconn

import (
	"encoding/binary"
	"github.com/pkg/errors"
	"io"
	"time"
)

func Uint16To2Bytes(val uint16) []byte {
	var tmp2 [2]byte
	binary.LittleEndian.PutUint16(tmp2[:], val)
	return tmp2[:]
}

func Uint32To4Bytes(val uint32) []byte {
	var tmp4 [4]byte
	binary.LittleEndian.PutUint32(tmp4[:], val)
	return tmp4[:]
}

func Uint32From4Bytes(b []byte) uint32 {
	if len(b) != 4 {
		panic("len(b) != 4")
	}
	return binary.LittleEndian.Uint32(b[:])
}

func Uint64To8Bytes(val uint64) []byte {
	var tmp8 [8]byte
	binary.LittleEndian.PutUint64(tmp8[:], val)
	return tmp8[:]
}

func WriteUint16(w io.Writer, val uint16) error {
	_, err := w.Write(Uint16To2Bytes(val))
	return err
}

func ReadUint16(r io.Reader, pval *uint16) error {
	var tmp2 [2]byte
	_, err := r.Read(tmp2[:])
	if err != nil {
		return err
	}
	*pval = binary.LittleEndian.Uint16(tmp2[:])
	return nil
}

func WriteUint32(w io.Writer, val uint32) error {
	_, err := w.Write(Uint32To4Bytes(val))
	return err
}

func ReadUint32(r io.Reader, pval *uint32) error {
	var tmp4 [4]byte
	_, err := r.Read(tmp4[:])
	if err != nil {
		return err
	}
	*pval = Uint32From4Bytes(tmp4[:])
	return nil
}

func WriteUint64(w io.Writer, val uint64) error {
	_, err := w.Write(Uint64To8Bytes(val))
	return err
}

func ReadUint64(r io.Reader, pval *uint64) error {
	var tmp8 [8]byte
	_, err := r.Read(tmp8[:])
	if err != nil {
		return err
	}
	*pval = binary.LittleEndian.Uint64(tmp8[:])
	return nil
}

func WriteBoolByte(w io.Writer, cond bool) error {
	var b [1]byte
	if cond {
		b[0] = 0xFF
	}
	_, err := w.Write(b[:])
	return err
}

func ReadBoolByte(r io.Reader, cond *bool) error {
	var b [1]byte
	_, err := r.Read(b[:])
	if err != nil {
		return err
	}
	*cond = b[0] == 0xFF
	if !*cond && b[0] != 0x00 {
		return errors.New("ReadBoolByte: unexpected value")
	}
	return nil
}

func WriteTime(w io.Writer, ts time.Time) error {
	return WriteUint64(w, uint64(ts.UnixNano()))
}

func ReadTime(r io.Reader, ts *time.Time) error {
	var nano uint64
	err := ReadUint64(r, &nano)
	if err != nil {
		return err
	}
	*ts = time.Unix(0, int64(nano))
	return nil
}
