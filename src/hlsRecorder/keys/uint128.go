package vmx

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

const (
	lessThan = iota - 1
	equal
	greaterThan

	Len = 32
)

// Big endian uint128
type Uint128 struct {
	H, L uint64
}

func (u *Uint128) Compare(o *Uint128) int {
	if u.H < o.H {
		return lessThan
	} else if u.H > o.H {
		return greaterThan
	}

	if u.L < o.L {
		return lessThan
	} else if u.L > o.L {
		return greaterThan
	}

	return equal
}

func (u *Uint128) And(o *Uint128) {
	u.H &= o.H
	u.L &= o.L
}

func (u *Uint128) Or(o *Uint128) {
	u.H |= o.H
	u.L |= u.L
}

func (u *Uint128) Xor(o *Uint128) {
	u.H ^= o.H
	u.L ^= o.L
}

func (u *Uint128) Add(o *Uint128) {
	carry := u.L
	u.L += o.L
	u.H += o.H

	if u.L < carry {
		u.H += 1
	}
}

// Декодирует вектор из строки, передавать сюда бинарную строку числа
func NewUint128FromString(s string) (u *Uint128, err error) {
	if len(s) > Len {
		return nil, fmt.Errorf("s:%s length greater than 32", s)
	}

	b, err := hex.DecodeString(fmt.Sprintf("%032s", s))
	if err != nil {
		return nil, err
	}
	rdr := bytes.NewReader(b)
	u = new(Uint128)
	err = binary.Read(rdr, binary.BigEndian, u)
	return
}

func (u *Uint128) Bytes() []byte {
	b := make([]byte, 16)
	b[15] = byte(u.L)
	b[14] = byte(u.L >> 8)
	b[13] = byte(u.L >> 16)
	b[12] = byte(u.L >> 24)
	b[11] = byte(u.L >> 32)
	b[10] = byte(u.L >> 40)
	b[9] = byte(u.L >> 48)
	b[8] = byte(u.L >> 56)
	b[7] = byte(u.H)
	b[6] = byte(u.H >> 8)
	b[5] = byte(u.H >> 16)
	b[4] = byte(u.H >> 24)
	b[3] = byte(u.H >> 32)
	b[2] = byte(u.H >> 40)
	b[1] = byte(u.H >> 48)
	b[1] = byte(u.H >> 56)
	return b
}

func (u *Uint128) HexString() string {
	if u.H == 0 {
		return fmt.Sprintf("%x", u.L)
	}
	return fmt.Sprintf("%x%016x", u.H, u.L)
}

func (u *Uint128) String() string {
	return fmt.Sprintf("0x%032x", u.HexString())
}
