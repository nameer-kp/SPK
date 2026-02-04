package nodes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nameer-kp/atem/pkg/node"
)

// BitwiseNode provides bit manipulation operations
// Essential for low-level computing, hardware simulation, and binary protocols
type BitwiseNode struct{}

func NewBitwiseNode() *BitwiseNode {
	return &BitwiseNode{}
}

func (n *BitwiseNode) Name() string {
	return "bitwise"
}

func (n *BitwiseNode) Description() string {
	return "Perform bitwise operations (AND, OR, XOR, shifts, etc.)"
}

func (n *BitwiseNode) ConfigSchema() map[string]any {
	return map[string]any{
		"operation": "string", // and, or, xor, not, lshift, rshift, getbit, setbit, clearbit, togglebit, count, reverse
		"a":         "number", // first operand
		"b":         "number", // second operand (for binary ops)
		"bits":      "number", // bit position (for bit operations) or shift amount
		"width":     "number", // bit width (default 64, max 64)
		"mask":      "number", // mask value
	}
}

func (n *BitwiseNode) Execute(ctx node.Context, config map[string]any) (node.Result, error) {
	operation, _ := config["operation"].(string)
	if operation == "" {
		return node.Result{}, fmt.Errorf("operation is required")
	}

	width := 64
	if w, ok := config["width"].(float64); ok && w > 0 && w <= 64 {
		width = int(w)
	}

	// Get operand a
	var a uint64
	if aVal, ok := config["a"].(float64); ok {
		a = uint64(aVal)
	}

	var result uint64
	var resultDesc string

	switch operation {
	case "and":
		b, err := getUint64(config, "b")
		if err != nil {
			return node.Result{}, err
		}
		result = a & b
		resultDesc = fmt.Sprintf("0x%X & 0x%X = 0x%X", a, b, result)

	case "or":
		b, err := getUint64(config, "b")
		if err != nil {
			return node.Result{}, err
		}
		result = a | b
		resultDesc = fmt.Sprintf("0x%X | 0x%X = 0x%X", a, b, result)

	case "xor":
		b, err := getUint64(config, "b")
		if err != nil {
			return node.Result{}, err
		}
		result = a ^ b
		resultDesc = fmt.Sprintf("0x%X ^ 0x%X = 0x%X", a, b, result)

	case "not":
		mask := uint64((1 << width) - 1)
		result = (^a) & mask
		resultDesc = fmt.Sprintf("~0x%X = 0x%X (width: %d)", a, result, width)

	case "lshift", "shl":
		bits, err := getUint64(config, "bits")
		if err != nil {
			if b, err2 := getUint64(config, "b"); err2 == nil {
				bits = b
			} else {
				return node.Result{}, err
			}
		}
		result = a << bits
		mask := uint64((1 << width) - 1)
		result &= mask
		resultDesc = fmt.Sprintf("0x%X << %d = 0x%X", a, bits, result)

	case "rshift", "shr":
		bits, err := getUint64(config, "bits")
		if err != nil {
			if b, err2 := getUint64(config, "b"); err2 == nil {
				bits = b
			} else {
				return node.Result{}, err
			}
		}
		result = a >> bits
		resultDesc = fmt.Sprintf("0x%X >> %d = 0x%X", a, bits, result)

	case "getbit":
		bit, err := getUint64(config, "bits")
		if err != nil {
			if b, err2 := getUint64(config, "b"); err2 == nil {
				bit = b
			} else {
				return node.Result{}, err
			}
		}
		if bit >= 64 {
			return node.Result{}, fmt.Errorf("bit position must be < 64")
		}
		result = (a >> bit) & 1
		resultDesc = fmt.Sprintf("bit %d of 0x%X = %d", bit, a, result)

	case "setbit":
		bit, err := getUint64(config, "bits")
		if err != nil {
			if b, err2 := getUint64(config, "b"); err2 == nil {
				bit = b
			} else {
				return node.Result{}, err
			}
		}
		if bit >= 64 {
			return node.Result{}, fmt.Errorf("bit position must be < 64")
		}
		result = a | (1 << bit)
		resultDesc = fmt.Sprintf("set bit %d: 0x%X -> 0x%X", bit, a, result)

	case "clearbit":
		bit, err := getUint64(config, "bits")
		if err != nil {
			if b, err2 := getUint64(config, "b"); err2 == nil {
				bit = b
			} else {
				return node.Result{}, err
			}
		}
		if bit >= 64 {
			return node.Result{}, fmt.Errorf("bit position must be < 64")
		}
		result = a &^ (1 << bit)
		resultDesc = fmt.Sprintf("clear bit %d: 0x%X -> 0x%X", bit, a, result)

	case "togglebit":
		bit, err := getUint64(config, "bits")
		if err != nil {
			if b, err2 := getUint64(config, "b"); err2 == nil {
				bit = b
			} else {
				return node.Result{}, err
			}
		}
		if bit >= 64 {
			return node.Result{}, fmt.Errorf("bit position must be < 64")
		}
		result = a ^ (1 << bit)
		resultDesc = fmt.Sprintf("toggle bit %d: 0x%X -> 0x%X", bit, a, result)

	case "count", "popcount":
		// Count set bits (population count)
		temp := a
		for temp != 0 {
			result++
			temp &= temp - 1
		}
		resultDesc = fmt.Sprintf("popcount(0x%X) = %d", a, result)

	case "reverse":
		// Reverse bits
		for i := 0; i < width; i++ {
			if (a & (1 << i)) != 0 {
				result |= 1 << (width - 1 - i)
			}
		}
		resultDesc = fmt.Sprintf("reverse(%d bits): 0x%X -> 0x%X", width, a, result)

	case "mask":
		maskVal, err := getUint64(config, "mask")
		if err != nil {
			if b, err2 := getUint64(config, "b"); err2 == nil {
				maskVal = b
			} else {
				return node.Result{}, err
			}
		}
		result = a & maskVal
		resultDesc = fmt.Sprintf("0x%X & mask(0x%X) = 0x%X", a, maskVal, result)

	case "rotl":
		// Rotate left
		bits, err := getUint64(config, "bits")
		if err != nil {
			return node.Result{}, err
		}
		bits = bits % uint64(width)
		mask := uint64((1 << width) - 1)
		result = ((a << bits) | (a >> (uint64(width) - bits))) & mask
		resultDesc = fmt.Sprintf("rotl(%d): 0x%X -> 0x%X", bits, a, result)

	case "rotr":
		// Rotate right
		bits, err := getUint64(config, "bits")
		if err != nil {
			return node.Result{}, err
		}
		bits = bits % uint64(width)
		mask := uint64((1 << width) - 1)
		result = ((a >> bits) | (a << (uint64(width) - bits))) & mask
		resultDesc = fmt.Sprintf("rotr(%d): 0x%X -> 0x%X", bits, a, result)

	default:
		return node.Result{}, fmt.Errorf("unknown operation: %s", operation)
	}

	// Generate binary representation
	binary := formatBinary(result, width)

	return node.Result{
		Success: true,
		Data: map[string]any{
			"result": result,
			"hex":    fmt.Sprintf("0x%X", result),
			"binary": binary,
			"signed": int64(result),
		},
		Status: resultDesc,
		Logs: []node.LogEntry{
			{Level: "info", Message: resultDesc},
		},
	}, nil
}

func getUint64(config map[string]any, key string) (uint64, error) {
	if v, ok := config[key].(float64); ok {
		return uint64(v), nil
	}
	return 0, fmt.Errorf("%s is required", key)
}

func formatBinary(n uint64, width int) string {
	s := strconv.FormatUint(n, 2)
	if len(s) < width {
		s = strings.Repeat("0", width-len(s)) + s
	}
	// Add spacing every 8 bits
	var result strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%8 == 0 {
			result.WriteRune('_')
		}
		result.WriteRune(c)
	}
	return result.String()
}
