package bitutils

import (
	"unsafe"
)

type Offset int

func NewOffset() *Offset {
	offset := Offset(0)

	return &offset
}

func (offset *Offset) Inc(delta int) (newOffset int) {
	newOffset = *(*int)(unsafe.Pointer(offset)) + delta

	*offset = *(*Offset)(unsafe.Pointer(&newOffset))

	return
}
