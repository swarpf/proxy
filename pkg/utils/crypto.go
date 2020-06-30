package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

var (
	key          = []byte("Gr4S2eiNl7zq5MrU") // key : AES encryption key for the SW API v2
	iv           = make([]byte, 16)           // iv : blank iv for the cipher
	aesCipher, _ = aes.NewCipher(key)         // aesCipher : the AES cipher itself
)

// EncryptBytes : Encrypts bytes for consumption by the Summoners War API
func EncryptBytes(plaintext []byte) ([]byte, error) {
	paddedPlaintext, err := pkcs7Pad(plaintext, aes.BlockSize)
	if err != nil {
		return nil, err
	}

	encrypter := cipher.NewCBCEncrypter(aesCipher, iv)

	ciphertext := make([]byte, len(paddedPlaintext))
	encrypter.CryptBlocks(ciphertext, paddedPlaintext)

	return ciphertext, nil
}

// EncryptString : Encrypts string for consumption by the Summoners War API
func EncryptString(plaintext string) (string, error) {
	ciphertext, err := EncryptBytes([]byte(plaintext))

	return string(ciphertext[:]), err
}

// DecryptBytes : Decrypts bytes sent by the Summoners War API
func DecryptBytes(ciphertext []byte) ([]byte, error) {
	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		return []byte{}, errors.New("ciphertext too short")
	}

	// CBC mode always works in whole blocks.
	if len(ciphertext)%aes.BlockSize != 0 {
		return []byte{}, errors.New("ciphertext is not a multiple of the block size")
	}

	decrypter := cipher.NewCBCDecrypter(aesCipher, iv)

	cipherBytes := make([]byte, len(ciphertext))
	copy(cipherBytes, ciphertext)

	decrypter.CryptBlocks(cipherBytes, cipherBytes)
	decryptedBytes := cipherBytes[:]

	return pkcs7Unpad(decryptedBytes, aes.BlockSize)
}

// DecryptString : Decrypts strings sent by the Summoners War API
func DecryptString(ciphertext string) (string, error) {
	decryptedBytes, err := DecryptBytes([]byte(ciphertext))

	return string(decryptedBytes[:]), err
}

// PKCS7 errors.
var (
	// ErrInvalidBlockSize indicates hash blocksize <= 0.
	ErrInvalidBlockSize = errors.New("invalid blocksize")

	// ErrInvalidPKCS7Data indicates bad input to PKCS7 pad or unpad.
	ErrInvalidPKCS7Data = errors.New("invalid PKCS7 data (empty or not padded)")

	// ErrInvalidPKCS7Padding indicates PKCS7 unpad fails to bad input.
	ErrInvalidPKCS7Padding = errors.New("invalid padding on input")
)

// pkcs7Pad : pkcs7Pad right-pads the given byte slice with 1 to n bytes, where
// n is the block size. The size of the result is x times n, where x
// is at least 1.
func pkcs7Pad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	n := blocksize - (len(b) % blocksize)
	pb := make([]byte, len(b)+n)
	copy(pb, b)
	copy(pb[len(b):], bytes.Repeat([]byte{byte(n)}, n))
	return pb, nil
}

func pkcs7Unpad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	if len(b)%blocksize != 0 {
		return nil, ErrInvalidPKCS7Padding
	}
	c := b[len(b)-1]
	n := int(c)
	if n == 0 || n > len(b) {
		return nil, ErrInvalidPKCS7Padding
	}

	return b[:len(b)-n], nil
}
