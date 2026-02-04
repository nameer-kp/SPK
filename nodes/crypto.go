package nodes

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"

	"github.com/nameer-kp/atem/pkg/node"
)

// CryptoNode provides hashing and basic cryptographic operations
type CryptoNode struct{}

func NewCryptoNode() *CryptoNode {
	return &CryptoNode{}
}

func (n *CryptoNode) Name() string {
	return "crypto"
}

func (n *CryptoNode) Description() string {
	return "Cryptographic operations: hashing, encoding, random generation"
}

func (n *CryptoNode) ConfigSchema() map[string]any {
	return map[string]any{
		"operation": "string", // hash, hmac, encode, decode, random
		"input":     "any",    // input data
		"algorithm": "string", // md5, sha1, sha256, sha512
		"key":       "string", // key for HMAC
		"encoding":  "string", // base64, hex
		"length":    "number", // length for random generation
	}
}

func (n *CryptoNode) Execute(ctx node.Context, config map[string]any) (node.Result, error) {
	operation, _ := config["operation"].(string)
	if operation == "" {
		return node.Result{}, fmt.Errorf("operation is required")
	}

	switch operation {
	case "hash":
		input, err := getInputBytes(config)
		if err != nil {
			return node.Result{}, err
		}

		algorithm, _ := config["algorithm"].(string)
		if algorithm == "" {
			algorithm = "sha256"
		}

		var h hash.Hash
		switch algorithm {
		case "md5":
			h = md5.New()
		case "sha1":
			h = sha1.New()
		case "sha256":
			h = sha256.New()
		case "sha512":
			h = sha512.New()
		default:
			return node.Result{}, fmt.Errorf("unknown hash algorithm: %s", algorithm)
		}

		h.Write(input)
		hashBytes := h.Sum(nil)

		return node.Result{
			Success: true,
			Data: map[string]any{
				"hex":       hex.EncodeToString(hashBytes),
				"base64":    base64.StdEncoding.EncodeToString(hashBytes),
				"algorithm": algorithm,
			},
			Status: algorithm,
		}, nil

	case "hmac":
		input, err := getInputBytes(config)
		if err != nil {
			return node.Result{}, err
		}

		key, _ := config["key"].(string)
		if key == "" {
			return node.Result{}, fmt.Errorf("key is required for HMAC")
		}

		algorithm, _ := config["algorithm"].(string)
		if algorithm == "" {
			algorithm = "sha256"
		}

		var h func() hash.Hash
		switch algorithm {
		case "md5":
			h = md5.New
		case "sha1":
			h = sha1.New
		case "sha256":
			h = sha256.New
		case "sha512":
			h = sha512.New
		default:
			return node.Result{}, fmt.Errorf("unknown HMAC algorithm: %s", algorithm)
		}

		mac := hmac.New(h, []byte(key))
		mac.Write(input)
		macBytes := mac.Sum(nil)

		return node.Result{
			Success: true,
			Data: map[string]any{
				"hex":       hex.EncodeToString(macBytes),
				"base64":    base64.StdEncoding.EncodeToString(macBytes),
				"algorithm": algorithm,
			},
			Status: "hmac-" + algorithm,
		}, nil

	case "encode":
		input, err := getInputBytes(config)
		if err != nil {
			return node.Result{}, err
		}

		encoding, _ := config["encoding"].(string)
		if encoding == "" {
			encoding = "base64"
		}

		var result string
		switch encoding {
		case "base64":
			result = base64.StdEncoding.EncodeToString(input)
		case "base64url":
			result = base64.URLEncoding.EncodeToString(input)
		case "hex":
			result = hex.EncodeToString(input)
		default:
			return node.Result{}, fmt.Errorf("unknown encoding: %s", encoding)
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"result":   result,
				"encoding": encoding,
				"length":   len(result),
			},
			Status: encoding,
		}, nil

	case "decode":
		input, _ := config["input"].(string)
		if input == "" {
			return node.Result{}, fmt.Errorf("input string is required for decode")
		}

		encoding, _ := config["encoding"].(string)
		if encoding == "" {
			encoding = "base64"
		}

		var decoded []byte
		var err error
		switch encoding {
		case "base64":
			decoded, err = base64.StdEncoding.DecodeString(input)
		case "base64url":
			decoded, err = base64.URLEncoding.DecodeString(input)
		case "hex":
			decoded, err = hex.DecodeString(input)
		default:
			return node.Result{}, fmt.Errorf("unknown encoding: %s", encoding)
		}

		if err != nil {
			return node.Result{}, fmt.Errorf("decode failed: %w", err)
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"result":   string(decoded),
				"bytes":    decoded,
				"encoding": encoding,
				"length":   len(decoded),
			},
			Status: encoding,
		}, nil

	case "random":
		length := 32
		if l, ok := config["length"].(float64); ok && l > 0 {
			length = int(l)
		}
		if length > 1024 {
			return node.Result{}, fmt.Errorf("random length exceeds 1024 bytes")
		}

		randomBytes := make([]byte, length)
		if _, err := rand.Read(randomBytes); err != nil {
			return node.Result{}, fmt.Errorf("failed to generate random bytes: %w", err)
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"hex":    hex.EncodeToString(randomBytes),
				"base64": base64.StdEncoding.EncodeToString(randomBytes),
				"bytes":  randomBytes,
				"length": length,
			},
			Status: fmt.Sprintf("%d random bytes", length),
		}, nil

	case "uuid":
		// Generate UUID v4
		uuid := make([]byte, 16)
		if _, err := rand.Read(uuid); err != nil {
			return node.Result{}, fmt.Errorf("failed to generate UUID: %w", err)
		}

		// Set version (4) and variant (RFC 4122)
		uuid[6] = (uuid[6] & 0x0f) | 0x40
		uuid[8] = (uuid[8] & 0x3f) | 0x80

		uuidStr := fmt.Sprintf("%x-%x-%x-%x-%x",
			uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])

		return node.Result{
			Success: true,
			Data: map[string]any{
				"uuid": uuidStr,
			},
			Status: "uuid",
		}, nil

	default:
		return node.Result{}, fmt.Errorf("unknown operation: %s", operation)
	}
}

func getInputBytes(config map[string]any) ([]byte, error) {
	input := config["input"]
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	switch v := input.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	case []any:
		// Try to interpret as byte array
		bytes := make([]byte, len(v))
		for i, b := range v {
			if num, ok := toFloat64(b); ok {
				bytes[i] = byte(int(num))
			} else {
				return nil, fmt.Errorf("invalid byte at index %d", i)
			}
		}
		return bytes, nil
	default:
		return []byte(fmt.Sprintf("%v", v)), nil
	}
}
