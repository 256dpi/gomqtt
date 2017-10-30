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

package transport

import (
	"crypto/tls"
	"fmt"
	"net"
)

var serverTLSConfig *tls.Config
var clientTLSConfig *tls.Config

var testDialer *Dialer
var testLauncher *Launcher

var testKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAtS5QzEX9hvsDMe9LQEpLI5vlLFjTxSMF8WIuAnsNc/racssY
7WS0VJy0Bdy/xvzUhsFDxcvML3jbfx9TG0ljsN5v4ExTE4AH68ZuVJNUuZjtu/6S
Tyy5HyNUKOCGhStOCLlDVrj+9d9IhwE74wSs0wWAz5xHYX3vBgthw+tQR9AX8T5l
PKKLeXyxO6145AUIx+4h9uW1a8l5twZqJfaPuUoHouL2pyznHKVbp8NMjb61otLP
Qv1uXDKe1Ish3sYxor8goE57/vKsawMmKvt0NQ3HjR1r9Q0kXnzQzFv7RBaeP8YY
SLrypnvRhZau6zTNElx6gD+dKH13ogRwF4U2ZwIDAQABAoIBAQCD7Ahq5J94dDbc
WheZ5T6W3AFfbAIubZT7MHi916lxBHug0B8DY6smXL978UM2eYup3vkhAYZYHpD8
G+vgV2yeMSiW0hw/e57I30AglvI+/YpEs1UHD1JvyMKTzQBF4MU99t+AOs1ya7Uc
kBpx70qDkGM9R9Z1t0OeTBVQmXtn/XRLpOD5mHn4or3yJrrsiOa7tVJUMZkwPkZb
hYpaBByKoNspVGfX5MbS4pyaR/BMDH8IR7tB0t0F8YtB0vUot0K4OMraCei7sjl0
s7wcc4FTflGdwxhLxGL3KCmQi8935ZhJuHbUltCrm9JvJ6PpB45DE8iZUEyF2ONn
/vVJ7lC5AoGBAN9MXIvnldp9OUNgwytC8L+jnNU9tamyvRvRb6PHKvFRzr3j7E4R
svUfNWcuntDLDpyh/Dd9i61MlEmP0xZmHeJfbpiz10TxwLdM02n42UcQSpZU/VUY
e/9JEohhQjjjW4CwAT0F8XNk5NRLyruGqeJdyPjZcNp1gonn5V2XSDQ1AoGBAM+2
8GJS27ws905DRDth33EMARi93boof4+EMrJEAFtFcg/VJMc8hTD37kafwUfrzIqE
WcpAH+OYSZyujwgORF4YB3O7UhIgzxGE7PH0c6g+MBiL3tfQ36CHXI5WrM/kqQZL
BAt32KVOWDHVIH+qlmlxppytuo1tLscII4ohPdurAoGBAMOEEhhx2vUm5gfH5ruv
biruMDlKQhDnLsv3sp+dmU0ZC0ObGOI2fxI/lDvXRzmcQhwFfVh89dr0jXZnlzUq
jooScz0DYr68Srh0bTyBpoWhjx1YJ0TmHlQlgDOnrRswf4MLGNp8pLAcLHxyLH3L
6F4LLhguE7FEaNslD+DFwHPtAoGAHsP+6OFS8JVMcAgglBcUDF91zh9ZNxiGhFBV
XfgagWpQ0a2FTAlXxQAEB+vjqe5wFlgyIHXNA13sEUv9nXjXgYRXnjmxl0PKx9KD
cfb+Jn5Hi2s0L40dgl5qRB7sa8J3kpoL2FMBAMMQ5xilPqOasbWGsxA8YHQ6iHpZ
IT15Nw8CgYEAtNLTthUVqMK5fy2t2iZWxndzfA+CmR+Gra5yTerPpQcQQoVS2bj9
K+Tg0aL8jcGRE4gBtdu0mAbQCQQM95orGfja3zD7LXOL9ljcPuYr70CfBWK9NQTo
2wJUIMgnoE9TCUiKa22GPH97dWX1q/CiK7vDFV0of+bOWx3CFJYsVSE=
-----END RSA PRIVATE KEY-----
`

var testCert = `-----BEGIN CERTIFICATE-----
MIIDtTCCAp2gAwIBAgIJAL8lt8+sd1TQMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIEwpTb21lLVN0YXRlMSEwHwYDVQQKExhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTYwMTA2MTYxMjE1WhcNMjYwMTAzMTYxMjE1WjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECBMKU29tZS1TdGF0ZTEhMB8GA1UEChMYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEAtS5QzEX9hvsDMe9LQEpLI5vlLFjTxSMF8WIuAnsNc/racssY7WS0VJy0
Bdy/xvzUhsFDxcvML3jbfx9TG0ljsN5v4ExTE4AH68ZuVJNUuZjtu/6STyy5HyNU
KOCGhStOCLlDVrj+9d9IhwE74wSs0wWAz5xHYX3vBgthw+tQR9AX8T5lPKKLeXyx
O6145AUIx+4h9uW1a8l5twZqJfaPuUoHouL2pyznHKVbp8NMjb61otLPQv1uXDKe
1Ish3sYxor8goE57/vKsawMmKvt0NQ3HjR1r9Q0kXnzQzFv7RBaeP8YYSLrypnvR
hZau6zTNElx6gD+dKH13ogRwF4U2ZwIDAQABo4GnMIGkMB0GA1UdDgQWBBQJem3s
aJLvZIO0e6TLzPTCrAc54DB1BgNVHSMEbjBsgBQJem3saJLvZIO0e6TLzPTCrAc5
4KFJpEcwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgTClNvbWUtU3RhdGUxITAfBgNV
BAoTGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZIIJAL8lt8+sd1TQMAwGA1UdEwQF
MAMBAf8wDQYJKoZIhvcNAQEFBQADggEBAI+XSsbSaB5juWzuqUKoCjVe1uN4sA+i
2xpz+IOqJ2MWR38QVQIpy7bF/21TYsv37U26Mt+yjs81QEYdjfZ+uZbvCKbJMBt6
z+EWC9mLUT65uG3tTBTgati8aIEmtKCXCr9R/S4QM7xBMqSF8gQCSC3jmKxmNCGX
BSfOCo8VrG4aXEYVKVUwSnzLTdNX0O6jJ2wJwm2/dFuMwKqRs6x7K3VVPoV2oTsn
xKQPmhbGAG14Fdu7g9I5jn/sAgTcW97FIPjX+wXwd6VU8K2kU5MBbrq3thVCQrZh
+9rdeUbNbDPsTlg24NlY9anbc6bG9WfrQgCkSA1VFUj6yYaFqT0sgAg=
-----END CERTIFICATE-----
`

func init() {
	cer, err := tls.X509KeyPair([]byte(testCert), []byte(testKey))
	if err != nil {
		panic(err)
	}

	serverTLSConfig = &tls.Config{Certificates: []tls.Certificate{cer}}
	clientTLSConfig = &tls.Config{InsecureSkipVerify: true}

	testDialer = NewDialer()
	testDialer.TLSConfig = clientTLSConfig

	testLauncher = NewLauncher()
	testLauncher.TLSConfig = serverTLSConfig
}

// returns a client-ish and server-ish pair of connections
func connectionPair(protocol string, handler func(Conn)) (Conn, chan struct{}) {
	done := make(chan struct{})

	server, err := testLauncher.Launch(protocol + "://localhost:0")
	if err != nil {
		panic(err)
	}

	go func() {
		conn, err := server.Accept()
		if err != nil {
			panic(err)
		}

		handler(conn)

		server.Close()
		close(done)
	}()

	conn, err := testDialer.Dial(getURL(server, protocol))
	if err != nil {
		panic(err)
	}

	return conn, done
}

func getPort(s Server) string {
	_, port, _ := net.SplitHostPort(s.Addr().String())
	return port
}

func getURL(s Server, protocol string) string {
	return fmt.Sprintf("%s://%s", protocol, s.Addr().String())
}
