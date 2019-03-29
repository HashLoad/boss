package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/denisbrodbeck/machineid"
	"github.com/hashload/boss/msg"
	"io"
)

func Encrypt(key []byte, message string) (cyphred string, err error) {
	plainText := []byte(message)

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	cyphred = base64.URLEncoding.EncodeToString(cipherText)
	return
}

func Decrypt(key []byte, securemess string) (decrypted string, err error) {
	cipherText, err := base64.URLEncoding.DecodeString(securemess)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return

	}

	if len(cipherText) < aes.BlockSize {
		err = errors.New("cipher text block size is too short")
		return
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	decrypted = string(cipherText)
	return
}
func GetMachineID() string {
	id, e := machineid.ID()
	if e != nil {
		msg.Err("Error on get machine ID")
		id = "AAAA"
	}
	return id
}

func Md5MachineID() string {
	hash := md5.New()
	if _, err := io.WriteString(hash, GetMachineID()); err != nil {
		msg.Warn("Failed on  write machine id to hash")
	}
	return hex.EncodeToString(hash.Sum(nil))
}
