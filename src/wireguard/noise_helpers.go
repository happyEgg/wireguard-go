package wireguard

import (
	"crypto/hmac"
	"crypto/rand"
	"golang.org/x/crypto/blake2s"
	"golang.org/x/crypto/curve25519"
	"hash"
)

/* KDF related functions.
 * HMAC-based Key Derivation Function (HKDF)
 * https://tools.ietf.org/html/rfc5869
 */

func HMAC1(sum *[blake2s.Size]byte, key, in0 []byte) {
	mac := hmac.New(func() hash.Hash {
		h, _ := blake2s.New256(nil)
		return h
	}, key)
	mac.Write(in0)
	mac.Sum(sum[:0])
}

func HMAC2(sum *[blake2s.Size]byte, key, in0, in1 []byte) {
	mac := hmac.New(func() hash.Hash {
		h, _ := blake2s.New256(nil)
		return h
	}, key)
	mac.Write(in0)
	mac.Write(in1)
	mac.Sum(sum[:0])
}

func KDF1(t0 *[blake2s.Size]byte, key, input []byte) {
	HMAC1(t0, key, input)
	HMAC1(t0, t0[:], []byte{0x1})
	return
}

func KDF2(t0, t1 *[blake2s.Size]byte, key, input []byte) {
	var prk [blake2s.Size]byte
	HMAC1(&prk, key, input)
	HMAC1(t0, prk[:], []byte{0x1})
	HMAC2(t1, prk[:], t0[:], []byte{0x2})
	setZero(prk[:])
	return
}

func KDF3(t0, t1, t2 *[blake2s.Size]byte, key, input []byte) {
	var prk [blake2s.Size]byte
	HMAC1(&prk, key, input)
	HMAC1(t0, prk[:], []byte{0x1})
	HMAC2(t1, prk[:], t0[:], []byte{0x2})
	HMAC2(t2, prk[:], t1[:], []byte{0x3})
	setZero(prk[:])
	return
}

func isZero(val []byte) bool {
	var acc byte
	for _, b := range val {
		acc |= b
	}
	return acc == 0
}

func setZero(arr []byte) {
	for i := range arr {
		arr[i] = 0
	}
}

/* curve25519 wrappers */

func newPrivateKey() (sk NoisePrivateKey, err error) {
	// clamping: https://cr.yp.to/ecdh.html
	_, err = rand.Read(sk[:])
	sk[0] &= 248
	sk[31] &= 127
	sk[31] |= 64
	return
}

func (sk *NoisePrivateKey) publicKey() (pk NoisePublicKey) {
	apk := (*[NoisePublicKeySize]byte)(&pk)
	ask := (*[NoisePrivateKeySize]byte)(sk)
	curve25519.ScalarBaseMult(apk, ask)
	return
}

func (sk *NoisePrivateKey) sharedSecret(pk NoisePublicKey) (ss [NoisePublicKeySize]byte) {
	apk := (*[NoisePublicKeySize]byte)(&pk)
	ask := (*[NoisePrivateKeySize]byte)(sk)
	curve25519.ScalarMult(&ss, ask, apk)
	return ss
}
