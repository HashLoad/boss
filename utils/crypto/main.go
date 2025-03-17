package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"

	//nolint:gosec // MD5 is used for hash comparison
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/denisbrodbeck/machineid"
	"github.com/hashload/boss/pkg/msg"
)

func Encrypt(key []byte, message string) (string, error) {
	plainText := []byte(message)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("error on create cipher: %w", err)
	}

	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("error on read random: %w", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	return base64.URLEncoding.EncodeToString(cipherText), nil
}

func Decrypt(key []byte, securemess string) (string, error) {
	cipherText, err := base64.URLEncoding.DecodeString(securemess)
	if err != nil {
		return "", fmt.Errorf("error on decode base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("error on create cipher: %w", err)
	}

	if len(cipherText) < aes.BlockSize {
		return "", errors.New("error on check block size: cipher text block size is too short")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}
func GetMachineID() string {
	id, err := machineid.ID()
	if err != nil {
		msg.Err("Error on get machine ID")
		id = "AAAA"
	}
	return id
}

func Md5MachineID() string {
	//nolint:gosec // MD5 is used for hash comparison
	hash := md5.New()
	if _, err := io.WriteString(hash, GetMachineID()); err != nil {
		msg.Warn("Failed on  write machine id to hash")
	}
	return hex.EncodeToString(hash.Sum(nil))
}
