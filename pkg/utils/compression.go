package utils

import (
	"bytes"
	"compress/zlib"
	"io/ioutil"
)

// CompressBytes : Compresses bytes using the zlib deflate method
func CompressBytes(plaintext []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)

	if _, err := w.Write(plaintext); err != nil {
		return nil, err
	}

	if err := w.Flush(); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// CompressString : Compresses strings using the zlib deflate method
func CompressString(plaintext string) (string, error) {
	compressed, err := CompressBytes([]byte(plaintext))

	return string(compressed[:]), err
}

// DecompressBytes : Decompresses zlib-compressed bytes
func DecompressBytes(compressed []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}

	decompressed, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return decompressed, nil
}

// DecompressString : Decompresses zlib-compressed strings
func DecompressString(compressed string) (string, error) {
	decompressedBytes, err := DecompressBytes([]byte(compressed))

	return string(decompressedBytes[:]), err
}
