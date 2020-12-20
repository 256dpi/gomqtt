package packet

import "unsafe"

func cast(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}
