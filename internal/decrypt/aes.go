package decrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"m3u8-download/pkg/m3u8"
)

func Decrypt(data, key []byte, iv []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, m3u8.ErrInvalidKey
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, m3u8.ErrDecryptFailed
	}

	blockSize := block.BlockSize()

	if len(iv) == 0 {
		iv = key
	}

	if len(iv) < blockSize {
		iv = make([]byte, blockSize)
		copy(iv, key)
	}

	blockMode := cipher.NewCBCDecrypter(block, iv[:blockSize])
	origData := make([]byte, len(data))
	blockMode.CryptBlocks(origData, data)
	origData = pKCS7UnPadding(origData)

	return origData, nil
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
