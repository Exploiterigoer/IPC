// Package common byte converted to integer or  integer converted to byte.
package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

// PadZero Fills the zero into the byte slice.
// parameter "count", the number of zreo to be filled.
func PadZero(count int) []byte {
	zero := make([]byte, 0)
	for i := 0; i < count; i++ {
		zero = append(zero, 0x00)
	}
	return zero
}

// TohexString Converts a byte slice to a hex string.
// parameter "msg",the slice to be converted.
func TohexString(msg []byte) []string {
	hexString := make([]string, len(msg))
	for k, v := range msg {
		tmp := fmt.Sprintf("%X", v)
		if len(tmp) < 2 {
			tmp = "0" + tmp
		}
		hexString[k] = tmp
	}
	return hexString
}

// ToNormalString converts a hex string to a normal string
// parameter "hexStr",the char to be converted
func ToNormalString(hexStr string) string {
	hexs := strings.Split(hexStr, "")
	hexb := make([]byte, 0)
	var hb int64
	tmp := ""
	for k, v := range hexs {
		if k%2 == 0 && len(tmp) == 2 {
			hb, _ = strconv.ParseInt(tmp, 16, 32)
			hexb = append(hexb, byte(hb))
			tmp = ""
		}
		tmp += v
	}

	// the last cycle,wihich can't be eliminated
	hb, _ = strconv.ParseInt(tmp, 16, 32)
	hexb = append(hexb, byte(hb))
	return string(hexb)
}

// ByteToInt converts a byte slice to an integer.
// parameter "b",the slice to be converted.
// parameter "ord",an integer with bigEndian or littleEndian,1 for bigEndian,0 for littleEndian.
// parameter "intType",the type of the integer to be converted.There is 16,32,64 for choosing.
func ByteToInt(b []byte, ord, intType int) uint64 {
	if len(b) == 1 {
		return uint64(b[0])
	}
	var i uint64
	switch intType {
	case 16:
		var tmp uint16
		if ord == 1 {
			binary.Read(bytes.NewBuffer(b), binary.BigEndian, &tmp)
		} else {
			binary.Read(bytes.NewBuffer(b), binary.LittleEndian, &tmp)
		}
		i = uint64(tmp)
	case 32:
		var tmp uint32
		if ord == 1 {
			binary.Read(bytes.NewBuffer(b), binary.BigEndian, &tmp)
		} else {
			binary.Read(bytes.NewBuffer(b), binary.LittleEndian, &tmp)
		}
		i = uint64(tmp)
	case 64:
		var tmp uint64
		if ord == 1 {
			binary.Read(bytes.NewBuffer(b), binary.BigEndian, &tmp)
		} else {
			binary.Read(bytes.NewBuffer(b), binary.LittleEndian, &tmp)
		}
		i = tmp
	}
	return i
}

// IntToByte converts an integer to a byte slice.
// parameter "i", the integer to be converted.
// parameter "ord",a byte slice with bigEndian or littleEndian,1 for bigEndian,0 for littleEndian.
// parameter "bitSize",the length of the byte slice,1 ror 1 bit,2 for 2bits,4 for 4 bits,8 for 8 bits.
func IntToByte(i, ord, bitSize int) []byte {
	buffer := bytes.NewBuffer([]byte{})
	switch bitSize {
	case 1:
		tmp := uint8(i)
		if ord == 1 {
			// a byte slice with 1 byte and bigEndian
			binary.Write(buffer, binary.BigEndian, &tmp)
		} else {
			// a byte slice with 1 byte and LittleEndian
			binary.Write(buffer, binary.LittleEndian, &tmp)
		}
	case 2:
		tmp := uint16(i)
		if ord == 1 {
			// a byte slice with 1 byte and bigEndian
			binary.Write(buffer, binary.BigEndian, &tmp)
		} else {
			// a byte slice with 1 byte and LittleEndian
			binary.Write(buffer, binary.LittleEndian, &tmp)
		}
	case 4:
		tmp := uint32(i)
		if ord == 1 {
			binary.Write(buffer, binary.BigEndian, &tmp)
		} else {
			binary.Write(buffer, binary.LittleEndian, &tmp)
		}
	case 8:
		tmp := uint64(i)
		if ord == 1 {
			binary.Write(buffer, binary.BigEndian, &tmp)
		} else {
			binary.Write(buffer, binary.LittleEndian, &tmp)
		}
	}
	return buffer.Bytes()
}

// ByteToUCS2 converts a byte slice to a byte of UCS2 format slice.
// patameter "b",the slice to be converted
// Note:it's the fact that UCS2 is just an uint16 integer,which has two bytes.
func ByteToUCS2(b []byte) []interface{} {
	ucs2 := make([]interface{}, 0)
	for k := range b {
		if (k+2)%2 == 0 {
			x := ByteToInt(b[k:k+2], 1, 16)
			ucs2 = append(ucs2, uint16(x))
		}
	}
	return ucs2
}

// UCS2ToByte converts an UCS2 to a byte slice.
// parameter "u", the UCS2 char to be converted.
// parameter "ord",a byte slice with bigEndian or littleEndian,1 for bigEndian,0 for littleEndian.
func UCS2ToByte(u uint16, ord int) []byte {
	buffer := bytes.NewBuffer([]byte{})
	if ord == 1 {
		binary.Write(buffer, binary.BigEndian, u)
	} else {
		binary.Write(buffer, binary.LittleEndian, u)
	}
	return buffer.Bytes()
}
