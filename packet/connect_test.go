package packet

import "testing"

func TestConnect(t *testing.T) {
	multiTest(t, func(t *testing.T, m Mode) {
		assertEncodeDecode(t, m, []byte{
			byte(CONNECT << 4),
			14, // remaining length
			0,  // protocol string
			6,
			'M', 'Q', 'I', 's', 'd', 'p',
			3, // protocol level
			2, // connect flags
			0, // keep alive
			10,
			0, // client id
			0,
		}, &Connect{
			KeepAlive:    10,
			CleanSession: true,
			Version:      Version3,
		}, `<Connect ClientID="" KeepAlive=10 Username="" Password="" CleanSession=true Will=nil Version=3>`)

		assertEncodeDecode(t, m, []byte{
			byte(CONNECT << 4),
			12, // remaining length
			0,  // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4, // protocol level
			2, // connect flags
			0, // keep alive
			10,
			0, // client id
			0,
		}, &Connect{
			KeepAlive:    10,
			CleanSession: true,
			Version:      Version4,
		}, `<Connect ClientID="" KeepAlive=10 Username="" Password="" CleanSession=true Will=nil Version=4>`)

		assertEncodeDecode(t, m, []byte{
			byte(CONNECT << 4),
			41, // remaining length
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
			3,
			'b', 'y', 'e',
			0, // username
			4,
			'u', 's', 'e', 'r',
			0, // password
			4,
			'p', 'a', 's', 's',
		}, &Connect{
			ClientID:     "gomqtt",
			KeepAlive:    10,
			Username:     "user",
			Password:     "pass",
			CleanSession: true,
			Will: &Message{
				Topic:   "will",
				Payload: []byte("bye"),
				QOS:     QOSAtLeastOnce,
				Retain:  false,
			},
			Version: Version4,
		}, `<Connect ClientID="gomqtt" KeepAlive=10 Username="user" Password="pass" CleanSession=true Will=<Message Topic="will" QOS=1 Retain=false Payload=627965> Version=4>`)

		assertEncodeDecode(t, m, []byte{
			byte(CONNECT << 4),
			43, // remaining length
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
			3,
			'b', 'y', 'e',
			0, // username
			4,
			'u', 's', 'e', 'r',
			0, // password
			4,
			'p', 'a', 's', 's',
		}, &Connect{
			ClientID:     "gomqtt",
			KeepAlive:    10,
			Username:     "user",
			Password:     "pass",
			CleanSession: true,
			Will: &Message{
				Topic:   "will",
				Payload: []byte("bye"),
				QOS:     QOSAtLeastOnce,
				Retain:  false,
			},
			Version: Version3,
		}, `<Connect ClientID="gomqtt" KeepAlive=10 Username="user" Password="pass" CleanSession=true Will=<Message Topic="will" QOS=1 Retain=false Payload=627965> Version=3>`)

		assertDecodeError(t, m, CONNECT, 2, []byte{
			byte(CONNECT << 4),
			60, // < wrong remaining length
			0,  // protocol string
			4,
			'M', 'Q', 'T', 'T',
		}, ErrRemainingLengthMismatch)

		assertDecodeError(t, m, CONNECT, 4, []byte{
			byte(CONNECT << 4),
			6, // remaining length
			0, // protocol string
			5, // < invalid size
			'M', 'Q', 'T', 'T',
		}, ErrInsufficientBufferSize)

		assertDecodeError(t, m, CONNECT, 8, []byte{
			byte(CONNECT << 4),
			6, // remaining length
			0, // protocol string
			4,
			'M', 'Q', 'T', 'T',
			// < protocol level: missing
		}, ErrInsufficientBufferSize)

		assertDecodeError(t, m, CONNECT, 10, []byte{
			byte(CONNECT << 4),
			8, // remaining length
			0, // protocol string
			5,
			'M', 'Q', 'T', 'T', 'X', // < wrong protocol string
			4,
		}, "invalid protocol string")

		assertDecodeError(t, m, CONNECT, 9, []byte{
			byte(CONNECT << 4),
			7, // remaining length
			0, // protocol string
			4,
			'M', 'Q', 'T', 'T',
			7, // < protocol level: wrong level
		}, "invalid protocol version")

		assertDecodeError(t, m, CONNECT, 9, []byte{
			byte(CONNECT << 4),
			7, // remaining length
			0, // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4, // protocol level
			// < connect flags: missing
		}, ErrInsufficientBufferSize)

		assertDecodeError(t, m, CONNECT, 10, []byte{
			byte(CONNECT << 4),
			8, // remaining length
			0, // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4, // protocol level
			1, // < connect flags: reserved bit set to one
		}, "invalid connect flags")

		assertDecodeError(t, m, CONNECT, 10, []byte{
			byte(CONNECT << 4),
			8, // remaining length
			0, // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4,  // protocol level
			24, // < connect flags: invalid qos
		}, ErrInvalidQOSLevel)

		assertDecodeError(t, m, CONNECT, 10, []byte{
			byte(CONNECT << 4),
			8, // remaining length
			0, // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4, // protocol level
			8, // < connect flags: will flag set to zero but others not
		}, "invalid connect flags")

		assertDecodeError(t, m, CONNECT, 10, []byte{
			byte(CONNECT << 4),
			8, // remaining length
			0, // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4,  // protocol level
			64, // < connect flags: password flag set but not username
		}, "invalid connect flags")

		assertDecodeError(t, m, CONNECT, 10, []byte{
			byte(CONNECT << 4),
			9, // remaining length
			0, // protocol string
			4,
			'M', 'Q', 'T', 'T',
			4, // protocol level
			0, // connect flags
			0, // < keep alive: incomplete
		}, ErrInsufficientBufferSize)

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
		}, ErrInsufficientBufferSize)

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
		}, "missing client id")

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
		}, ErrInsufficientBufferSize)

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
		}, ErrInsufficientBufferSize)

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
		}, ErrInsufficientBufferSize)

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
		}, ErrInsufficientBufferSize)

		assertDecodeError(t, m, CONNECT, 16, []byte{
			byte(CONNECT << 4),
			15, // remaining length
			0,  // protocol string
			6,
			'M', 'Q', 'I', 's', 'd', 'p',
			3, // protocol level
			2, // connect flags
			0, // keep alive
			10,
			0, // client id
			0,
			0, // < superfluous byte
		}, "leftover buffer length")

		// small buffer
		assertEncodeError(t, m, 1, 1, &Connect{}, ErrInsufficientBufferSize)

		assertEncodeError(t, m, 0, 9, &Connect{
			CleanSession: true,
			Version:      Version4,
			Will: &Message{
				Topic: "t",
				QOS:   3, // < wrong qos
			},
		}, ErrInvalidQOSLevel)

		assertEncodeError(t, m, 0, 14, &Connect{
			CleanSession: true,
			Version:      Version4,
			ClientID:     longString, // < too big
		}, ErrPrefixedBytesOverflow)

		assertEncodeError(t, m, 0, 16, &Connect{
			CleanSession: true,
			Version:      Version4,
			Will: &Message{
				Topic: longString, // < too big
			},
		}, ErrPrefixedBytesOverflow)

		assertEncodeError(t, m, 0, 19, &Connect{
			CleanSession: true,
			Version:      Version4,
			Will: &Message{
				Topic:   "t",
				Payload: longBytes, // < too big
			},
		}, ErrPrefixedBytesOverflow)

		assertEncodeError(t, m, 0, 16, &Connect{
			CleanSession: true,
			Version:      Version4,
			Username:     longString, // < too big
		}, ErrPrefixedBytesOverflow)

		assertEncodeError(t, m, 0, 14, &Connect{
			CleanSession: true,
			Version:      Version4,
			Username:     "", // < missing
			Password:     "p",
		}, "missing username")

		assertEncodeError(t, m, 0, 19, &Connect{
			CleanSession: true,
			Version:      Version4,
			Username:     "u",
			Password:     longString, // < too big
		}, ErrPrefixedBytesOverflow)

		assertEncodeError(t, m, 0, 9, &Connect{
			CleanSession: true,
			Version:      Version4,
			Will: &Message{
				Topic: "", // < missing
			},
		}, "missing will topic")

		assertEncodeError(t, m, 0, 2, &Connect{
			Version: 255, // < invalid
		}, "invalid protocol version")

		assertEncodeError(t, m, 0, 9, &Connect{
			CleanSession: false,
			ClientID:     "", // < missing
		}, "missing client id")
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
		Version: Version4,
	})
}
