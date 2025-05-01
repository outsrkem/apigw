package crypto

import (
	"apigw/src/slog"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
)

const (
	keyStr = "Npf4zWUvqDp6LmQtNxkorgn1qSAgSMGW"
)

var key = []byte(keyStr)

// =================== CFB ======================

// aesEncryptCFB encrypts the original data using AES in Cipher Feedback (CFB) mode.
func aesEncryptCFB(origData []byte, key []byte) (encrypted []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	encrypted = make([]byte, aes.BlockSize+len(origData))
	iv := encrypted[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(encrypted[aes.BlockSize:], origData)
	return encrypted
}

// aesDecryptCFB decrypts the encrypted data using AES in CFB mode.
// TODO 携程中解密失败（panic）会导致主进程奔溃
func aesDecryptCFB(encrypted []byte, key []byte) (decrypted []byte) {
	block, _ := aes.NewCipher(key)
	if len(encrypted) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(encrypted, encrypted)
	return encrypted
}

// Encryption encrypts a plain text string.
func Encryption(plain string) string {
	plaByt := []byte(plain)
	encrypted := aesEncryptCFB(plaByt, key)
	_cipher := hex.EncodeToString(encrypted)
	return _cipher
}

// Decryption decrypts a cipher text string.
func Decryption(cipherText string) (string, error) {
	klog := slog.FromContext(nil)
	encryptedHex, err := hex.DecodeString(cipherText)
	if err != nil {
		klog.Error("Unable to decode hexadecimal string: ", err)
		return "", err
	}
	plain := aesDecryptCFB(encryptedHex, key)
	return string(plain), nil
}
