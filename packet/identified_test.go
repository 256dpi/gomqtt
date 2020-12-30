package packet

import (
	"fmt"
	"testing"
)

func testIdentified(t *testing.T, pkt Generic) {
	multiTest(t, func(t *testing.T, m Mode) {
		assertEncodeDecode(t, m, []byte{
			byte(pkt.Type()<<4) | pkt.Type().defaultFlags(),
			2, // remaining length
			0, // packet id
			1,
		}, pkt, fmt.Sprintf("<%s ID=1>", pkt.Type().String()))

		assertDecodeError(t, m, pkt.Type(), 2, []byte{
			byte(pkt.Type()<<4) | pkt.Type().defaultFlags(),
			1, // < wrong remaining length
			0, // packet id
			7,
		})

		assertDecodeError(t, m, pkt.Type(), 2, []byte{
			byte(pkt.Type()<<4) | pkt.Type().defaultFlags(),
			2, // remaining length
			7, // packet id
			// < insufficient bytes
		})

		assertDecodeError(t, m, pkt.Type(), 4, []byte{
			byte(pkt.Type()<<4) | pkt.Type().defaultFlags(),
			2, // remaining length
			0, // packet id
			0, // < zero id
		})

		assertDecodeError(t, m, pkt.Type(), 4, []byte{
			byte(pkt.Type()<<4) | pkt.Type().defaultFlags(),
			3, // remaining length
			0, // packet id
			7,
			0, // < superfluous byte
		})

		// small buffer
		assertEncodeError(t, m, 1, 1, &Puback{})

		assertEncodeError(t, m, 0, 2, &Puback{
			ID: 0, // < zero id
		})
	})
}

func TestPuback(t *testing.T) {
	testIdentified(t, &Puback{
		ID: 1,
	})
}

func TestPubcomp(t *testing.T) {
	testIdentified(t, &Pubcomp{
		ID: 1,
	})
}

func TestPubrec(t *testing.T) {
	testIdentified(t, &Pubrec{
		ID: 1,
	})
}

func TestPubrel(t *testing.T) {
	testIdentified(t, &Pubrel{
		ID: 1,
	})
}

func TestUnsuback(t *testing.T) {
	testIdentified(t, &Unsuback{
		ID: 1,
	})
}

func BenchmarkIdentified(b *testing.B) {
	benchPacket(b, &Puback{
		ID: 1,
	})
}
