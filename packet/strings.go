// Copyright (c) 2014 The gomqtt Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package packet

import (
	"encoding/binary"
	"fmt"
)

const maxLPLength uint16 = 65535

// read length prefixed bytes
func readLPBytes(buf []byte, safe bool, t Type) ([]byte, int, error) {
	if len(buf) < 2 {
		return nil, 0, fmt.Errorf("[%s] insufficient buffer size, expected 2, got %d", t, len(buf))
	}

	n, total := 0, 0

	n = int(binary.BigEndian.Uint16(buf))
	total += 2
	total += n

	if len(buf) < total {
		return nil, total, fmt.Errorf("[%s] insufficient buffer size, expected %d, got %d", t, total, len(buf))
	}

	// copy buffer in safe mode
	if safe {
		newBuf := make([]byte, total-2)
		copy(newBuf, buf[2:total])
		return newBuf, total, nil
	}

	return buf[2:total], total, nil
}

// read length prefixed string
func readLPString(buf []byte, t Type) (string, int, error) {
	if len(buf) < 2 {
		return "", 0, fmt.Errorf("[%s] insufficient buffer size, expected 2, got %d", t, len(buf))
	}

	n, total := 0, 0

	n = int(binary.BigEndian.Uint16(buf))
	total += 2
	total += n

	if len(buf) < total {
		return "", total, fmt.Errorf("[%s] insufficient buffer size, expected %d, got %d", t, total, len(buf))
	}

	return string(buf[2:total]), total, nil
}

// write length prefixed bytes
func writeLPBytes(buf []byte, b []byte, t Type) (int, error) {
	total, n := 0, len(b)

	if n > int(maxLPLength) {
		return 0, fmt.Errorf("[%s] length (%d) greater than %d bytes", t, n, maxLPLength)
	}

	if len(buf) < 2+n {
		return 0, fmt.Errorf("[%s] insufficient buffer size, expected %d, got %d", t, 2+n, len(buf))
	}

	binary.BigEndian.PutUint16(buf, uint16(n))
	total += 2

	copy(buf[total:], b)
	total += n

	return total, nil
}

// write length prefixed string
func writeLPString(buf []byte, str string, t Type) (int, error) {
	return writeLPBytes(buf, []byte(str), t)
}

// checks the QOS value to see if it's valid
func validQOS(qos byte) bool {
	return qos == QOSAtMostOnce || qos == QOSAtLeastOnce || qos == QOSExactlyOnce
}
