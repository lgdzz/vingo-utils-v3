package cryptor

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
)

// Constants for default layer
const defaultLayer = 3

// Generates the S string (A-Z, a-z, 0-9)
func generateCharset() string {
	var charset string
	for i := 65; i <= 90; i++ {
		charset += string(i) // A-Z
	}
	for i := 97; i <= 122; i++ {
		charset += string(i) // a-z
	}
	for i := 48; i <= 57; i++ {
		charset += string(i) // 0-9
	}
	return charset
}

// Generates a random string with a length prefix (1-9) and random characters
func generateRandomPrefix() string {
	charset := generateCharset()
	n, _ := rand.Int(rand.Reader, big.NewInt(9)) // Random number between 1-9
	n = n.Add(n, big.NewInt(1))                  // Ensure n is between 1 and 9
	prefix := fmt.Sprintf("%d", n.Int64())

	for i := int64(0); i < n.Int64(); i++ {
		randomIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		prefix += string(charset[randomIndex.Int64()]) // Randomly pick characters from charset
	}

	return prefix
}

// Base64 encoding function similar to btoa
func encodeBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

// Function to swap base64 character pairs while preserving the padding
func swapBase64Pairs(input string) string {
	// Preserve padding
	padding := ""
	for len(input) > 0 && input[len(input)-1] == '=' {
		padding = "=" + padding
		input = input[:len(input)-1]
	}

	// Swap pairs of characters
	result := ""
	for i := 0; i < len(input); i += 2 {
		if i+1 < len(input) {
			result += string(input[i+1]) + string(input[i])
		} else {
			result += string(input[i])
		}
	}
	return result + padding
}

// Reverse the base64 pair swap operation while preserving padding
func reverseBase64PairSwap(s string) string {
	// Split off padding
	padding := ""
	for strings.HasSuffix(s, "=") {
		s = s[:len(s)-1]
		padding += "="
	}

	// Reverse the character pair swaps
	runes := []rune(s)
	for i := 0; i+1 < len(runes); i += 2 {
		runes[i], runes[i+1] = runes[i+1], runes[i]
	}

	return string(runes) + padding
}

// Decode function to reverse the encoding and pair swapping
func decodeBase64Pairs(encoded string, times int) (string, error) {
	result := encoded

	// Perform the reverse pair swap and decode operations for the given number of times
	for t := 0; t < times; t++ {
		result = reverseBase64PairSwap(result)
		decoded, err := base64.StdEncoding.DecodeString(result)
		if err != nil {
			return "", fmt.Errorf("base64 decode failed at loop %d: %v", t, err)
		}
		result = string(decoded)
	}

	// Remove the prefix (random length and characters)
	if len(result) < 1 {
		return "", fmt.Errorf("invalid encoded string")
	}

	prefixLength := result[0]
	if prefixLength < '1' || prefixLength > '9' {
		return "", fmt.Errorf("invalid prefix first character")
	}

	n := int(prefixLength - '0')
	if len(result) < 1+n {
		return "", fmt.Errorf("encoded string too short")
	}
	return result[1+n:], nil
}

// ExchangeEncode Main encoding function to encode the string with multiple layers of encoding and pair swapping
func ExchangeEncode(text string, times ...int) string {
	layer := defaultLayer
	if len(times) > 0 {
		layer = times[0]
	}

	// Generate a random prefix and concatenate it with the original text
	randomPrefix := generateRandomPrefix()
	encoded := fmt.Sprintf("%d%s%s", len(randomPrefix)-1, randomPrefix, text)

	// Apply Base64 encoding and swap pairs for the specified number of layers
	for i := 0; i < layer; i++ {
		encoded = encodeBase64(encoded)
		encoded = swapBase64Pairs(encoded)
	}

	return encoded
}

// ExchangeDecode Main decoding function to decode the encoded text using multiple layers of decoding and reverse pair swapping
func ExchangeDecode(text string, times ...int) string {
	layer := defaultLayer
	if len(times) > 0 {
		layer = times[0]
	}
	decoded, err := decodeBase64Pairs(text, layer)
	if err != nil {
		fmt.Println("Error during decoding:", err)
		return ""
	}
	return decoded
}
