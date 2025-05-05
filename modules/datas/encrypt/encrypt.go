package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/scrypt"
)

const (
	ivLen     = 16 // AES block size
	secretLen = 24 // Per-message random secret length
	keyLen    = 24 // Final AES-192 key length

	// Scrypt parameters
	scryptN    = 16384
	scryptR    = 8
	scryptP    = 1
	scryptSalt = "salt" // Use a unique, consistent salt for your application
)

var (
	ErrInvalidHashLength    = errors.New("invalid hash length for decryption")
	ErrInvalidCiphertextLen = errors.New("ciphertext is not a multiple of the block size")
	ErrInvalidPadding       = errors.New("invalid pkcs7 padding")
	ErrGlobalKeyNotSet      = errors.New("global key environment variable not set or empty")
	ErrFailedToCombineKeys  = errors.New("failed to combine global key and secret using HMAC")
)

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, ErrInvalidPadding
	}
	unpadding := int(data[length-1])
	if unpadding > blockSize || unpadding == 0 {
		return nil, ErrInvalidPadding
	}
	if length < unpadding {
		return nil, ErrInvalidPadding
	}
	// Check that the padding bytes are correct
	for i := length - unpadding; i < length; i++ {
		if data[i] != byte(unpadding) {
			return nil, ErrInvalidPadding
		}
	}
	return data[:(length - unpadding)], nil
}

func combineKeys(globalKey, secret []byte) ([]byte, error) {
	if len(globalKey) == 0 || len(secret) == 0 {
		return nil, ErrFailedToCombineKeys
	}
	mac := hmac.New(sha256.New, globalKey)
	_, err := mac.Write(secret)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToCombineKeys, err)
	}
	return mac.Sum(nil), nil
}

// Encrypt encrypts text using AES-192-CBC with a key derived from the provided globalKey.
// Accepts globalKey as an argument.
// Returns hex-encoded string: ivHex + randomSecretHex + ciphertextHex
func Encrypt(globalKey, text string) (string, error) {
	if len(globalKey) == 0 {
		return "", ErrGlobalKeyNotSet // Check passed key
	}
	plaintext := []byte(text)

	iv := make([]byte, ivLen)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}

	randomSecret := make([]byte, secretLen)
	if _, err := io.ReadFull(rand.Reader, randomSecret); err != nil {
		return "", fmt.Errorf("failed to generate random secret: %w", err)
	}

	scryptInput, err := combineKeys([]byte(globalKey), randomSecret)
	if err != nil {
		return "", err
	}

	finalEncryptionKey, err := scrypt.Key(scryptInput, []byte(scryptSalt), scryptN, scryptR, scryptP, keyLen)
	if err != nil {
		return "", fmt.Errorf("failed to derive key using scrypt: %w", err)
	}

	block, err := aes.NewCipher(finalEncryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	paddedPlaintext := pkcs7Pad(plaintext, aes.BlockSize)

	ciphertext := make([]byte, len(paddedPlaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	ivHex := hex.EncodeToString(iv)
	randomSecretHex := hex.EncodeToString(randomSecret)
	ciphertextHex := hex.EncodeToString(ciphertext)

	return ivHex + randomSecretHex + ciphertextHex, nil
}

// Decrypt decrypts a hash string created by Encrypt, using the provided globalKey.
// Accepts globalKey as an argument.
// Parses hex-encoded: ivHex + randomSecretHex + ciphertextHex
func Decrypt(globalKey, hash string) (string, error) {
	if len(globalKey) == 0 {
		return "", ErrGlobalKeyNotSet // Check passed key
	}

	ivHexLen := ivLen * 2
	secretHexLen := secretLen * 2
	minHashLen := ivHexLen + secretHexLen

	if len(hash) < minHashLen {
		return "", ErrInvalidHashLength
	}

	ivHex := hash[:ivHexLen]
	randomSecretHex := hash[ivHexLen : ivHexLen+secretHexLen]
	ciphertextHex := hash[ivHexLen+secretHexLen:]

	iv, err := hex.DecodeString(ivHex)
	if err != nil || len(iv) != ivLen {
		return "", fmt.Errorf("failed to decode IV: %w", err)
	}

	randomSecret, err := hex.DecodeString(randomSecretHex)
	if err != nil || len(randomSecret) != secretLen {
		return "", fmt.Errorf("failed to decode random secret: %w", err)
	}

	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", ErrInvalidCiphertextLen
	}

	scryptInput, err := combineKeys([]byte(globalKey), randomSecret)
	if err != nil {
		return "", err
	}

	finalEncryptionKey, err := scrypt.Key(scryptInput, []byte(scryptSalt), scryptN, scryptR, scryptP, keyLen)
	if err != nil {
		return "", fmt.Errorf("failed to derive key using scrypt: %w", err)
	}

	block, err := aes.NewCipher(finalEncryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	decryptedPadded := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decryptedPadded, ciphertext)

	plaintext, err := pkcs7Unpad(decryptedPadded, aes.BlockSize)
	if err != nil {
		return "", fmt.Errorf("failed to unpad data (wrong global key or corrupted hash?): %w", err)
	}

	return string(plaintext), nil
}
