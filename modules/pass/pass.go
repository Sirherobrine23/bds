package pass

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256" // Using SHA256 for HMAC
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	// To read environment variables
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

	// Environment variable to store the global key
	// !! IN PRODUCTION: Use a secure secret management system !!
	globalKeyEnvVar = "MY_APP_GLOBAL_SECRET_KEY"
)

var (
	ErrInvalidHashLength    = errors.New("invalid hash length for decryption")
	ErrInvalidCiphertextLen = errors.New("ciphertext is not a multiple of the block size")
	ErrInvalidPadding       = errors.New("invalid pkcs7 padding")
	ErrGlobalKeyNotSet      = errors.New("global key environment variable not set or empty")
	ErrFailedToCombineKeys  = errors.New("failed to combine global key and secret using HMAC")
)

// --- pkcs7Pad and pkcs7Unpad functions remain the same as before ---

// pkcs7Pad adds PKCS#7 padding to a byte slice.
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7Unpad removes PKCS#7 padding from a byte slice.
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

// combineKeys uses HMAC-SHA256 to combine the global key and the per-message secret.
// This result will be used as the password input for scrypt.
func combineKeys(globalKey, secret []byte) ([]byte, error) {
	if len(globalKey) == 0 || len(secret) == 0 {
		// Prevent HMAC with empty key or message which might behave unexpectedly
		return nil, ErrFailedToCombineKeys
	}
	mac := hmac.New(sha256.New, globalKey)
	_, err := mac.Write(secret)
	if err != nil {
		// This error is unlikely for hmac but check anyway
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

	// 1. Generate random IV
	iv := make([]byte, ivLen)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}

	// 2. Generate random per-message secret
	randomSecret := make([]byte, secretLen)
	if _, err := io.ReadFull(rand.Reader, randomSecret); err != nil {
		return "", fmt.Errorf("failed to generate random secret: %w", err)
	}

	// 3. Combine Global Key and random secret using HMAC -> This is the input for Scrypt
	scryptInput, err := combineKeys([]byte(globalKey), randomSecret)
	if err != nil {
		return "", err
	}

	// 4. Derive final encryption key using Scrypt
	finalEncryptionKey, err := scrypt.Key(scryptInput, []byte(scryptSalt), scryptN, scryptR, scryptP, keyLen)
	if err != nil {
		return "", fmt.Errorf("failed to derive key using scrypt: %w", err)
	}

	// 5. Create AES cipher block
	block, err := aes.NewCipher(finalEncryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// 6. Pad plaintext
	paddedPlaintext := pkcs7Pad(plaintext, aes.BlockSize)

	// 7. Encrypt using CBC mode
	ciphertext := make([]byte, len(paddedPlaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	// 8. Concatenate IV, *random* Secret, and Ciphertext, then Hex encode
	ivHex := hex.EncodeToString(iv)
	randomSecretHex := hex.EncodeToString(randomSecret) // Store the original random secret
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
	secretHexLen := secretLen * 2 // Length of the *random* secret in hex
	minHashLen := ivHexLen + secretHexLen

	if len(hash) < minHashLen {
		return "", ErrInvalidHashLength
	}

	// 1. Extract and decode IV, *random* Secret, Ciphertext from hex hash
	ivHex := hash[:ivHexLen]
	randomSecretHex := hash[ivHexLen : ivHexLen+secretHexLen]
	ciphertextHex := hash[ivHexLen+secretHexLen:]

	iv, err := hex.DecodeString(ivHex)
	if err != nil || len(iv) != ivLen {
		return "", fmt.Errorf("failed to decode IV: %w", err)
	}

	randomSecret, err := hex.DecodeString(randomSecretHex) // Decode the stored random secret
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

	// 2. Combine Global Key and the extracted random secret using HMAC -> This is the input for Scrypt
	scryptInput, err := combineKeys([]byte(globalKey), randomSecret)
	if err != nil {
		return "", err
	}


	// 3. Derive final encryption key using Scrypt (must use same salt and parameters)
	finalEncryptionKey, err := scrypt.Key(scryptInput, []byte(scryptSalt), scryptN, scryptR, scryptP, keyLen)
	if err != nil {
		return "", fmt.Errorf("failed to derive key using scrypt: %w", err)
	}

	// 4. Create AES cipher block
	block, err := aes.NewCipher(finalEncryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// 5. Decrypt using CBC mode
	if len(ciphertext) < aes.BlockSize {
		 return "", errors.New("ciphertext too short")
	}
	decryptedPadded := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decryptedPadded, ciphertext)

	// 6. Unpad the decrypted data
	plaintext, err := pkcs7Unpad(decryptedPadded, aes.BlockSize)
	if err != nil {
		return "", fmt.Errorf("failed to unpad data (wrong global key or corrupted hash?): %w", err)
	}

	// 7. Return as string
	return string(plaintext), nil
}
