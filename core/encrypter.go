package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

func Encrypt(text, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		CreateErrorLog("", "Unable to create new cypher:", err)
		return nil
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		CreateErrorLog("", "Error during encryption: ", err)
		return nil
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext
}

func Decrypt(text, key []byte) (out []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		CreateErrorLog("", "Error creating encryption cypher", len(text))
		return nil
	}
	if len(text) < aes.BlockSize {
		CreateErrorLog("", "Encryption cypher is too short: ", len(text))
		return nil
	}

	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	out = make([]byte, len(text))
	cfb.XORKeyStream(out, text)
	data, err := base64.StdEncoding.DecodeString(string(out))
	if err != nil {
		CreateErrorLog("", "Error during dencryption: ", err)
		return nil
	}
	return data
}
