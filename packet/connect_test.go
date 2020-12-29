package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		pkt := NewConnect()
		pkt.ClientID = "c"
		pkt.KeepAlive = 5
		pkt.Username = "u"
		pkt.Password = "p"
		pkt.CleanSession = true
		pkt.Will = &Message{
			Topic:   "w",
			Payload: []byte("m"),
			QOS:     QOSAtLeastOnce,
			Retain:  true,
		}

		assert.Equal(t, pkt.Type(), CONNECT)
		assert.Equal(t, `<Connect ClientID="c" KeepAlive=5 Username="u" Password="p" CleanSession=true Will=<Message Topic="w" QOS=1 Retain=true Payload=6d> Version=4>`, pkt.String())

		buf := make([]byte, pkt.Len(m))
		n1, err := pkt.Encode(m, buf)
		assert.NoError(t, err)

		pkt2 := NewConnect()
		n2, err := pkt2.Decode(m, buf)
		assert.NoError(t, err)

		assert.Equal(t, pkt, pkt2)
		assert.Equal(t, n1, n2)
	})
}

func TestConnectDecode(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(CONNECT << 4),
			58, // remaining length
			0,  // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4,   // protocol level
			206, // connect flags
			0,   // keep alive
			10,
			0, // client id
			6,
			'g', 'o', 'm', 'q', 't', 't',
			0, // will topic
			4,
			'w', 'i', 'l', 'l',
			0, // will message
			12,
			's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
			0, // username
			6,
			'g', 'o', 'm', 'q', 't', 't',
			0, // password
			10,
			'v', 'e', 'r', 'y', 's', 'e', 'c', 'r', 'e', 't',
		}

		pkt := NewConnect()
		n, err := pkt.Decode(m, packet)
		assert.NoError(t, err)
		assert.Equal(t, len(packet), n)
		assert.Equal(t, uint16(10), pkt.KeepAlive)
		assert.Equal(t, "gomqtt", pkt.ClientID)
		assert.Equal(t, "will", pkt.Will.Topic)
		assert.Equal(t, []byte("send me home"), pkt.Will.Payload)
		assert.Equal(t, "gomqtt", pkt.Username)
		assert.Equal(t, "verysecret", pkt.Password)
		assert.Equal(t, Version311, pkt.Version)

		packet = []byte{
			byte(CONNECT << 4),
			60, // remaining length
			0,  // protocol string
			6,
			'M', 'Q', 'I', 's', 'd', 'p',
			3,   // protocol level
			206, // connect flags
			0,   // keep alive
			10,
			0, // client id
			6,
			'g', 'o', 'm', 'q', 't', 't',
			0, // will topic
			4,
			'w', 'i', 'l', 'l',
			0, // will message
			12,
			's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
			0, // username
			6,
			'g', 'o', 'm', 'q', 't', 't',
			0, // password
			10,
			'v', 'e', 'r', 'y', 's', 'e', 'c', 'r', 'e', 't',
		}

		pkt = NewConnect()
		n, err = pkt.Decode(m, packet)
		assert.NoError(t, err)
		assert.Equal(t, len(packet), n)
		assert.Equal(t, uint16(10), pkt.KeepAlive)
		assert.Equal(t, "gomqtt", pkt.ClientID)
		assert.Equal(t, "will", pkt.Will.Topic)
		assert.Equal(t, []byte("send me home"), pkt.Will.Payload)
		assert.Equal(t, "gomqtt", pkt.Username)
		assert.Equal(t, "verysecret", pkt.Password)
		assert.Equal(t, Version31, pkt.Version)

		assertDecodeError(t, m, CONNECT, 2, []byte{
			byte(CONNECT << 4),
			60, // < wrong remaining length
			0,  // protocol string
			5,
			'M', 'Q', 'T', 'T',
		})

		assertDecodeError(t, m, CONNECT, 4, []byte{
			byte(CONNECT << 4),
			6, // remaining length
			0, // protocol string
			5, // < invalid size
			'M', 'Q', 'T', 'T',
		})

		assertDecodeError(t, m, CONNECT, 8, []byte{
			byte(CONNECT << 4),
			6, // remaining length
			0, // protocol string
			4,
			'M', 'Q', 'T', 'T',
			// < protocol level: missing
		})

		assertDecodeError(t, m, CONNECT, 10, []byte{
			byte(CONNECT << 4),
			8, // remaining length
			0, // protocol string
			5,
			'M', 'Q', 'T', 'T', 'X', // < wrong protocol string
			4,
		})

		assertDecodeError(t, m, CONNECT, 9, []byte{
			byte(CONNECT << 4),
			7, // remaining length
			0, // protocol string
			4,
			'M', 'Q', 'T', 'T',
			7, // < protocol level: wrong level
		})

		assertDecodeError(t, m, CONNECT, 9, []byte{
			byte(CONNECT << 4),
			7, // remaining length
			0, // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4, // protocol level
			// < connect flags: missing
		})

		assertDecodeError(t, m, CONNECT, 10, []byte{
			byte(CONNECT << 4),
			8, // remaining length
			0, // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4, // protocol level
			1, // < connect flags: reserved bit set to one
		})

		assertDecodeError(t, m, CONNECT, 10, []byte{
			byte(CONNECT << 4),
			8, // remaining length
			0, // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4,  // protocol level
			24, // < connect flags: invalid qos
		})

		assertDecodeError(t, m, CONNECT, 10, []byte{
			byte(CONNECT << 4),
			8, // remaining length
			0, // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4, // protocol level
			8, // < connect flags: will flag set to zero but others not
		})

		assertDecodeError(t, m, CONNECT, 10, []byte{
			byte(CONNECT << 4),
			8, // remaining length
			0, // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4,  // protocol level
			64, // < connect flags: password flag set but not username
		})

		assertDecodeError(t, m, CONNECT, 10, []byte{
			byte(CONNECT << 4),
			9, // remaining length
			0, // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4, // protocol level
			0, // connect flags
			0, // < keep alive: missing
			// < keep alive: missing
		})

		assertDecodeError(t, m, CONNECT, 14, []byte{
			byte(CONNECT << 4),
			13, // remaining length
			0,  // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4, // protocol level
			0, // connect flags
			0, // keep alive
			1,
			0, // client id
			2, // < wrong size
			'x',
		})

		assertDecodeError(t, m, CONNECT, 14, []byte{
			byte(CONNECT << 4),
			12, // remaining length
			0,  // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4, // protocol level
			0, // < connect flags: clean session false
			0, // keep alive
			1,
			0, // client id
			0,
		})

		assertDecodeError(t, m, CONNECT, 16, []byte{
			byte(CONNECT << 4),
			14, // remaining length
			0,  // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4, // protocol level
			6, // connect flags
			0, // keep alive
			1,
			0, // client id
			0,
			0, // will topic
			1, // < wrong size
		})

		assertDecodeError(t, m, CONNECT, 18, []byte{
			byte(CONNECT << 4),
			16, // remaining length
			0,  // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4, // protocol level
			6, // connect flags
			0, // keep alive
			1,
			0, // client id
			0,
			0, // will topic
			0,
			0, // will payload
			1, // < wrong size
		})

		assertDecodeError(t, m, CONNECT, 16, []byte{
			byte(CONNECT << 4),
			14, // remaining length
			0,  // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4,   // protocol level
			194, // connect flags
			0,   // keep alive
			1,
			0, // client id
			0,
			0, // username
			1, // < wrong size
		})

		assertDecodeError(t, m, CONNECT, 18, []byte{
			byte(CONNECT << 4),
			16, // remaining length
			0,  // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4,   // protocol level
			194, // connect flags
			0,   // keep alive
			1,
			0, // client id
			0,
			0, // username
			0,
			0, // password
			1, // < wrong size
		})
	})
}

func TestConnectEncode(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		packet := []byte{
			byte(CONNECT << 4),
			58, // remaining length
			0,  // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4,   // protocol level 4
			204, // connect flags
			0,   // keep alive
			10,
			0, // client id
			6,
			'g', 'o', 'm', 'q', 't', 't',
			0, // will topic
			4,
			'w', 'i', 'l', 'l',
			0, // will message
			12,
			's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
			0, // username
			6,
			'g', 'o', 'm', 'q', 't', 't',
			0, // password
			10,
			'v', 'e', 'r', 'y', 's', 'e', 'c', 'r', 'e', 't',
		}

		pkt := NewConnect()
		pkt.Will = &Message{
			QOS:     QOSAtLeastOnce,
			Topic:   "will",
			Payload: []byte("send me home"),
		}
		pkt.CleanSession = false
		pkt.ClientID = "gomqtt"
		pkt.KeepAlive = 10
		pkt.Username = "gomqtt"
		pkt.Password = "verysecret"

		dst := make([]byte, pkt.Len(m))
		n, err := pkt.Encode(m, dst)
		assert.NoError(t, err)
		assert.Equal(t, len(packet), n)
		assert.Equal(t, packet, dst[:n])

		packet = []byte{
			byte(CONNECT << 4),
			14, // remaining length
			0,  // protocol string
			6,
			'M', 'Q', 'I', 's', 'd', 'p',
			3, // protocol level 4
			2, // connect flags
			0, // keep alive
			10,
			0, // client id
			0,
		}

		pkt = NewConnect()
		pkt.CleanSession = true
		pkt.KeepAlive = 10
		pkt.Version = Version31

		dst = make([]byte, pkt.Len(m))
		n, err = pkt.Encode(m, dst)
		assert.NoError(t, err)
		assert.Equal(t, len(packet), n)
		assert.Equal(t, packet, dst[:n])

		packet = []byte{
			byte(CONNECT << 4),
			12, // remaining length
			0,  // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4, // protocol level 4
			2, // connect flags
			0, // keep alive
			10,
			0, // client id
			0,
		}

		pkt = NewConnect()
		pkt.CleanSession = true
		pkt.KeepAlive = 10
		pkt.Version = 0

		dst = make([]byte, pkt.Len(m))
		n, err = pkt.Encode(m, dst)
		assert.NoError(t, err)
		assert.Equal(t, len(packet), n)
		assert.Equal(t, packet, dst[:n])

		// small buffer
		assertEncodeError(t, m, 1, 1, &Connect{})

		assertEncodeError(t, m, 0, 9, &Connect{
			CleanSession: true,
			Version:      4,
			Will: &Message{
				Topic: "t",
				QOS:   3, // < wrong qos
			},
		})

		assertEncodeError(t, m, 0, 14, &Connect{
			CleanSession: true,
			Version:      4,
			ClientID:     longString, // < too big
		})

		assertEncodeError(t, m, 0, 16, &Connect{
			CleanSession: true,
			Version:      4,
			Will: &Message{
				Topic: longString, // < too big
			},
		})

		assertEncodeError(t, m, 0, 19, &Connect{
			CleanSession: true,
			Version:      4,
			Will: &Message{
				Topic:   "t",
				Payload: make([]byte, 65536), // < too big
			},
		})

		assertEncodeError(t, m, 0, 16, &Connect{
			CleanSession: true,
			Version:      4,
			Username:     longString, // < too big
		})

		assertEncodeError(t, m, 0, 14, &Connect{
			CleanSession: true,
			Version:      4,
			Username:     "", // < missing
			Password:     "p",
		})

		assertEncodeError(t, m, 0, 19, &Connect{
			CleanSession: true,
			Version:      4,
			Username:     "u",
			Password:     longString, // < too big
		})

		assertEncodeError(t, m, 0, 9, &Connect{
			CleanSession: true,
			Version:      4,
			Will: &Message{
				Topic: "", // < missing
			},
		})

		assertEncodeError(t, m, 0, 2, &Connect{
			Version: 255, // < invalid
		})

		assertEncodeError(t, m, 0, 9, &Connect{
			CleanSession: false,
			ClientID:     "", // < missing
		})
	})
}

func BenchmarkConnect(b *testing.B) {
	benchPacket(b, &Connect{
		ClientID:     "i",
		KeepAlive:    10,
		Username:     "u",
		Password:     "p",
		CleanSession: false,
		Will: &Message{
			Topic:   "t",
			Payload: []byte("m"),
			QOS:     QOSAtLeastOnce,
			Retain:  true,
		},
		Version: Version311,
	})
}
