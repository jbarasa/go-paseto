package paseto

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestVectors struct {
	Name  string
	Tests []TestVector
}

type TestVector struct {
	Name                string
	Nonce               string
	Key                 string
	PublicKey           string `json:"public-key"`
	SecretKey           string `json:"secret-key"`
	Token               string
	Payload             string
	Footer              string
	ExpectFail          bool   `json:"expect-fail"`
	ImplicitAssertation string `json:"implicit-assertion"`
}

func TestV4(t *testing.T) {
	data, err := os.ReadFile("test-vectors/v4.json")
	require.NoError(t, err)

	var tests TestVectors
	err = json.Unmarshal(data, &tests)
	require.NoError(t, err)

	for _, test := range tests.Tests {
		t.Run(test.Name, func(t *testing.T) {
			var decoded Packet

			switch test.Key {
			// Local mode
			default:
				sk, err := V4SymmetricKeyFromHex(test.Key)
				require.NoError(t, err)

				message, err := NewMessage(V4Local, test.Token)
				if test.ExpectFail {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				decoded, err = V4LocalDecrypt(message, sk, []byte(test.ImplicitAssertation))
				if test.ExpectFail {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

			// Public mode
			case "":
				pk, err := NewV4AsymmetricPublicKeyFromHex(test.PublicKey)
				require.NoError(t, err)

				message, err := NewMessage(V4Public, test.Token)
				if test.ExpectFail {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				decoded, err = V4PublicVerify(message, pk, []byte(test.ImplicitAssertation))
				if test.ExpectFail {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
			}

			require.Equal(t, test.Payload, string(decoded.Content))
			require.Equal(t, test.Footer, string(decoded.Footer))

			packet := NewPacket([]byte(test.Payload), []byte(test.Footer))
			implicit := []byte(test.ImplicitAssertation)

			switch test.Key {
			// Local mode
			default:
				sk, err := V4SymmetricKeyFromHex(test.Key)
				require.NoError(t, err)

				unitTestNonce, err := hex.DecodeString(test.Nonce)
				require.NoError(t, err)

				encrypted := v4LocalEncrypt(packet, sk, implicit, unitTestNonce)
				require.NoError(t, err)

				require.Equal(t, test.Token, encrypted.Encoded())

			// Public mode
			case "":
				sk, err := NewV4AsymmetricSecretKeyFromHex(test.SecretKey)
				require.NoError(t, err)

				signed := V4PublicSign(packet, sk, implicit)
				require.NoError(t, err)

				require.Equal(t, test.Token, signed.Encoded())
			}
		})
	}
}