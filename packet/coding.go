package packet

const maxVarUint = 268435455

func varUintLen(n uint64) int {
	if n < 128 {
		return 1
	} else if n < 16384 {
		return 2
	} else if n < 2097152 {
		return 3
	} else if n <= maxVarUint {
		return 4
	}

	return 0
}
