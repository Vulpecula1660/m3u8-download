package decrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"m3u8-download/pkg/m3u8"
	"sync"
)

type Decryptor struct {
	block cipher.Block
	iv    []byte
	mu    sync.Mutex
}

func NewDecryptor(key []byte, iv []byte) (*Decryptor, error) {
	if len(key) == 0 {
		return nil, m3u8.ErrInvalidKey
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, m3u8.ErrDecryptFailed
	}

	blockSize := block.BlockSize()

	if len(iv) == 0 {
		iv = make([]byte, blockSize)
		copy(iv, key)
	} else if len(iv) < blockSize {
		temp := make([]byte, blockSize)
		copy(temp, iv)
		iv = temp
	}

	return &Decryptor{
		block: block,
		iv:    iv[:blockSize],
	}, nil
}

func (d *Decryptor) Decrypt(data []byte) ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	blockMode := cipher.NewCBCDecrypter(d.block, d.iv)
	origData := make([]byte, len(data))
	blockMode.CryptBlocks(origData, data)
	origData = pKCS7UnPadding(origData)

	return origData, nil
}

func Decrypt(data, key []byte, iv []byte) ([]byte, error) {
	decryptor, err := NewDecryptor(key, iv)
	if err != nil {
		return nil, err
	}
	return decryptor.Decrypt(data)
}

func pKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	if length == 0 {
		return origData
	}
	unpadding := int(origData[length-1])
	if unpadding > length {
		return origData
	}
	return origData[:(length - unpadding)]
}

func RemoveSyncBytePrefix(data []byte) []byte {
	syncByte := uint8(71)
	bLen := len(data)

	for j := 0; j < bLen; j++ {
		if data[j] == syncByte {
			return data[j:]
		}
	}

	return data
}
