package decrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"testing"
)

func TestDecrypt(t *testing.T) {
	key := make([]byte, 16)
	rand.Read(key)

	plaintext := []byte("Hello, this is a test message for AES decryption!")

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("failed to create cipher: %v", err)
	}

	paddedPlaintext := pkcs7Pad(plaintext, aes.BlockSize)
	iv := make([]byte, aes.BlockSize)
	rand.Read(iv)

	ciphertext := make([]byte, len(paddedPlaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	decrypted, err := Decrypt(ciphertext, key, iv)
	if err != nil {
		t.Errorf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted data doesn't match original\ngot:  %v\nwant: %v", decrypted, plaintext)
	}
}

func TestDecryptInvalidKey(t *testing.T) {
	ciphertext := make([]byte, 16)
	rand.Read(ciphertext)

	_, err := Decrypt(ciphertext, nil, nil)
	if err == nil {
		t.Error("expected error for nil key")
	}
}

func TestDecryptInvalidIV(t *testing.T) {
	key := make([]byte, 16)
	rand.Read(key)
	ciphertext := make([]byte, 16)
	rand.Read(ciphertext)

	_, err := Decrypt(ciphertext, key, []byte{1, 2, 3})
	if err != nil {
		t.Errorf("unexpected error with short IV: %v", err)
	}
}

func TestPKCS7UnPadding(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		padding int
	}{
		{
			name:    "1 byte padding",
			data:    []byte("Hello\x01"),
			padding: 1,
		},
		{
			name:    "5 byte padding",
			data:    []byte("Hello\x05\x05\x05\x05\x05"),
			padding: 5,
		},
		{
			name:    "16 byte padding (full block)",
			data:    bytes.Repeat([]byte{16}, 16),
			padding: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pKCS7UnPadding(tt.data)

			expectedLen := len(tt.data) - tt.padding
			if len(result) != expectedLen {
				t.Errorf("got length %d, want %d", len(result), expectedLen)
			}
		})
	}
}

func TestRemoveSyncBytePrefix(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want []byte
	}{
		{
			name: "data starts with sync byte",
			data: []byte{0x47, 0x01, 0x02, 0x03},
			want: []byte{0x47, 0x01, 0x02, 0x03},
		},
		{
			name: "data has prefix before sync byte",
			data: []byte{0xFF, 0xFE, 0xFD, 0x47, 0x01, 0x02},
			want: []byte{0x47, 0x01, 0x02},
		},
		{
			name: "no sync byte in data",
			data: []byte{0x01, 0x02, 0x03, 0x04},
			want: []byte{0x01, 0x02, 0x03, 0x04},
		},
		{
			name: "empty data",
			data: []byte{},
			want: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveSyncBytePrefix(tt.data)

			if !bytes.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	if padding == 0 {
		padding = blockSize
	}
	padded := append([]byte{}, data...)
	for i := 0; i < padding; i++ {
		padded = append(padded, byte(padding))
	}
	return padded
}
