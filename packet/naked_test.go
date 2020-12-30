package packet

import (
	"fmt"
	"testing"
)

func testNaked(t *testing.T, pkt Generic) {
	multiTest(t, func(t *testing.T, m Mode) {
		assertEncodeDecode(t, m, []byte{
			byte(pkt.Type() << 4),
			0, // remaining length
		}, pkt, fmt.Sprintf("<%s>", pkt.Type().String()))

		assertDecodeError(t, m, pkt.Type(), 2, []byte{
			byte(pkt.Type() << 4),
			1, // remaining length
			0, // < superfluous byte
		})
	})
}

func TestDisconnect(t *testing.T) {
	testNaked(t, &Disconnect{})
}

func TestPingreq(t *testing.T) {
	testNaked(t, &Pingreq{})
}

func TestPingresp(t *testing.T) {
	testNaked(t, &Pingresp{})
}

func TestAuth(t *testing.T) {
	testNaked(t, &Auth{})
}

func BenchmarkNaked(b *testing.B) {
	benchPacket(b, &Disconnect{})
}
