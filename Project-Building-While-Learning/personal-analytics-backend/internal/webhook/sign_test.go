package webhook

import "testing"

func Test_HexCode(t *testing.T) {

	hexString := Sign([]byte("hello"), "mysecret")

	hardCodedHexString := "f09399f0c446d84b31a080e57ec483392d41e6f512f3e7ada5027abbcd358c2a"

	if hexString != hardCodedHexString {
		t.Errorf("expected %s, got %s", hardCodedHexString, hexString)
	}
}
