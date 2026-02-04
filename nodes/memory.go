package nodes

import (
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/nameer-kp/atem/pkg/node"
)

// Memory buffers for simulation (workflow-scoped)
var (
	memBuffers   = make(map[string][]byte)
	memBuffersMu sync.RWMutex
)

// MemoryNode provides array/buffer operations for memory simulation
// Essential for simulating RAM, registers, and other memory structures
type MemoryNode struct{}

func NewMemoryNode() *MemoryNode {
	return &MemoryNode{}
}

func (n *MemoryNode) Name() string {
	return "memory"
}

func (n *MemoryNode) Description() string {
	return "Manage memory buffers (allocate, read, write bytes)"
}

func (n *MemoryNode) ConfigSchema() map[string]any {
	return map[string]any{
		"operation": "string", // alloc, free, read, write, read16, write16, read32, write32, fill, copy, dump, list
		"name":      "string", // buffer name
		"size":      "number", // size in bytes (for alloc)
		"address":   "number", // byte address
		"value":     "any",    // value to write
		"length":    "number", // length for operations
		"from":      "string", // source buffer (for copy)
		"from_addr": "number", // source address (for copy)
		"to":        "string", // destination buffer (for copy)
		"to_addr":   "number", // destination address (for copy)
	}
}

func (n *MemoryNode) Execute(ctx node.Context, config map[string]any) (node.Result, error) {
	operation, _ := config["operation"].(string)
	if operation == "" {
		return node.Result{}, fmt.Errorf("operation is required")
	}

	name, _ := config["name"].(string)

	switch operation {
	case "alloc":
		if name == "" {
			return node.Result{}, fmt.Errorf("buffer name is required")
		}
		size := 256
		if s, ok := config["size"].(float64); ok && s > 0 {
			size = int(s)
		}
		if size > 1024*1024 { // 1MB limit
			return node.Result{}, fmt.Errorf("buffer size exceeds 1MB limit")
		}

		memBuffersMu.Lock()
		memBuffers[name] = make([]byte, size)
		memBuffersMu.Unlock()

		return node.Result{
			Success: true,
			Data: map[string]any{
				"name": name,
				"size": size,
			},
			Status: fmt.Sprintf("allocated %s (%d bytes)", name, size),
			Logs: []node.LogEntry{
				{Level: "info", Message: fmt.Sprintf("Allocated buffer %s with %d bytes", name, size)},
			},
		}, nil

	case "free":
		if name == "" {
			return node.Result{}, fmt.Errorf("buffer name is required")
		}

		memBuffersMu.Lock()
		_, existed := memBuffers[name]
		delete(memBuffers, name)
		memBuffersMu.Unlock()

		return node.Result{
			Success: true,
			Data: map[string]any{
				"name":  name,
				"freed": existed,
			},
			Status: fmt.Sprintf("freed %s", name),
			Logs: []node.LogEntry{
				{Level: "info", Message: fmt.Sprintf("Freed buffer %s (existed: %v)", name, existed)},
			},
		}, nil

	case "read", "read8":
		if name == "" {
			return node.Result{}, fmt.Errorf("buffer name is required")
		}
		addr, err := getAddress(config)
		if err != nil {
			return node.Result{}, err
		}

		memBuffersMu.RLock()
		buf, exists := memBuffers[name]
		if !exists {
			memBuffersMu.RUnlock()
			return node.Result{}, fmt.Errorf("buffer %s not found", name)
		}
		if addr >= len(buf) {
			memBuffersMu.RUnlock()
			return node.Result{}, fmt.Errorf("address %d out of bounds (size: %d)", addr, len(buf))
		}
		value := buf[addr]
		memBuffersMu.RUnlock()

		return node.Result{
			Success: true,
			Data: map[string]any{
				"value":   int(value),
				"address": addr,
				"hex":     fmt.Sprintf("0x%02X", value),
			},
			Status: fmt.Sprintf("[%d] = 0x%02X", addr, value),
		}, nil

	case "write", "write8":
		if name == "" {
			return node.Result{}, fmt.Errorf("buffer name is required")
		}
		addr, err := getAddress(config)
		if err != nil {
			return node.Result{}, err
		}
		val, err := getValue(config)
		if err != nil {
			return node.Result{}, err
		}

		memBuffersMu.Lock()
		buf, exists := memBuffers[name]
		if !exists {
			memBuffersMu.Unlock()
			return node.Result{}, fmt.Errorf("buffer %s not found", name)
		}
		if addr >= len(buf) {
			memBuffersMu.Unlock()
			return node.Result{}, fmt.Errorf("address %d out of bounds (size: %d)", addr, len(buf))
		}
		buf[addr] = byte(val)
		memBuffersMu.Unlock()

		return node.Result{
			Success: true,
			Data: map[string]any{
				"address": addr,
				"value":   val & 0xFF,
			},
			Status: fmt.Sprintf("[%d] <- 0x%02X", addr, val&0xFF),
		}, nil

	case "read16":
		if name == "" {
			return node.Result{}, fmt.Errorf("buffer name is required")
		}
		addr, err := getAddress(config)
		if err != nil {
			return node.Result{}, err
		}

		memBuffersMu.RLock()
		buf, exists := memBuffers[name]
		if !exists {
			memBuffersMu.RUnlock()
			return node.Result{}, fmt.Errorf("buffer %s not found", name)
		}
		if addr+1 >= len(buf) {
			memBuffersMu.RUnlock()
			return node.Result{}, fmt.Errorf("address %d out of bounds for 16-bit read", addr)
		}
		// Little-endian
		value := uint16(buf[addr]) | (uint16(buf[addr+1]) << 8)
		memBuffersMu.RUnlock()

		return node.Result{
			Success: true,
			Data: map[string]any{
				"value":   int(value),
				"address": addr,
				"hex":     fmt.Sprintf("0x%04X", value),
			},
			Status: fmt.Sprintf("[%d:16] = 0x%04X", addr, value),
		}, nil

	case "write16":
		if name == "" {
			return node.Result{}, fmt.Errorf("buffer name is required")
		}
		addr, err := getAddress(config)
		if err != nil {
			return node.Result{}, err
		}
		val, err := getValue(config)
		if err != nil {
			return node.Result{}, err
		}

		memBuffersMu.Lock()
		buf, exists := memBuffers[name]
		if !exists {
			memBuffersMu.Unlock()
			return node.Result{}, fmt.Errorf("buffer %s not found", name)
		}
		if addr+1 >= len(buf) {
			memBuffersMu.Unlock()
			return node.Result{}, fmt.Errorf("address %d out of bounds for 16-bit write", addr)
		}
		// Little-endian
		buf[addr] = byte(val)
		buf[addr+1] = byte(val >> 8)
		memBuffersMu.Unlock()

		return node.Result{
			Success: true,
			Data: map[string]any{
				"address": addr,
				"value":   val & 0xFFFF,
			},
			Status: fmt.Sprintf("[%d:16] <- 0x%04X", addr, val&0xFFFF),
		}, nil

	case "read32":
		if name == "" {
			return node.Result{}, fmt.Errorf("buffer name is required")
		}
		addr, err := getAddress(config)
		if err != nil {
			return node.Result{}, err
		}

		memBuffersMu.RLock()
		buf, exists := memBuffers[name]
		if !exists {
			memBuffersMu.RUnlock()
			return node.Result{}, fmt.Errorf("buffer %s not found", name)
		}
		if addr+3 >= len(buf) {
			memBuffersMu.RUnlock()
			return node.Result{}, fmt.Errorf("address %d out of bounds for 32-bit read", addr)
		}
		// Little-endian
		value := uint32(buf[addr]) | (uint32(buf[addr+1]) << 8) | (uint32(buf[addr+2]) << 16) | (uint32(buf[addr+3]) << 24)
		memBuffersMu.RUnlock()

		return node.Result{
			Success: true,
			Data: map[string]any{
				"value":   int64(value),
				"address": addr,
				"hex":     fmt.Sprintf("0x%08X", value),
			},
			Status: fmt.Sprintf("[%d:32] = 0x%08X", addr, value),
		}, nil

	case "write32":
		if name == "" {
			return node.Result{}, fmt.Errorf("buffer name is required")
		}
		addr, err := getAddress(config)
		if err != nil {
			return node.Result{}, err
		}
		val, err := getValue(config)
		if err != nil {
			return node.Result{}, err
		}

		memBuffersMu.Lock()
		buf, exists := memBuffers[name]
		if !exists {
			memBuffersMu.Unlock()
			return node.Result{}, fmt.Errorf("buffer %s not found", name)
		}
		if addr+3 >= len(buf) {
			memBuffersMu.Unlock()
			return node.Result{}, fmt.Errorf("address %d out of bounds for 32-bit write", addr)
		}
		// Little-endian
		buf[addr] = byte(val)
		buf[addr+1] = byte(val >> 8)
		buf[addr+2] = byte(val >> 16)
		buf[addr+3] = byte(val >> 24)
		memBuffersMu.Unlock()

		return node.Result{
			Success: true,
			Data: map[string]any{
				"address": addr,
				"value":   val & 0xFFFFFFFF,
			},
			Status: fmt.Sprintf("[%d:32] <- 0x%08X", addr, val&0xFFFFFFFF),
		}, nil

	case "fill":
		if name == "" {
			return node.Result{}, fmt.Errorf("buffer name is required")
		}
		val := 0
		if v, ok := config["value"].(float64); ok {
			val = int(v)
		}

		memBuffersMu.Lock()
		buf, exists := memBuffers[name]
		if !exists {
			memBuffersMu.Unlock()
			return node.Result{}, fmt.Errorf("buffer %s not found", name)
		}
		for i := range buf {
			buf[i] = byte(val)
		}
		memBuffersMu.Unlock()

		return node.Result{
			Success: true,
			Data: map[string]any{
				"name":  name,
				"value": val & 0xFF,
				"size":  len(buf),
			},
			Status: fmt.Sprintf("filled %s with 0x%02X", name, val&0xFF),
		}, nil

	case "copy":
		from, _ := config["from"].(string)
		to, _ := config["to"].(string)
		if from == "" || to == "" {
			return node.Result{}, fmt.Errorf("from and to buffer names required")
		}

		fromAddr := 0
		if a, ok := config["from_addr"].(float64); ok {
			fromAddr = int(a)
		}
		toAddr := 0
		if a, ok := config["to_addr"].(float64); ok {
			toAddr = int(a)
		}
		length := -1
		if l, ok := config["length"].(float64); ok {
			length = int(l)
		}

		memBuffersMu.Lock()
		srcBuf, srcExists := memBuffers[from]
		dstBuf, dstExists := memBuffers[to]
		if !srcExists {
			memBuffersMu.Unlock()
			return node.Result{}, fmt.Errorf("source buffer %s not found", from)
		}
		if !dstExists {
			memBuffersMu.Unlock()
			return node.Result{}, fmt.Errorf("destination buffer %s not found", to)
		}

		if length < 0 {
			length = len(srcBuf) - fromAddr
		}
		if fromAddr+length > len(srcBuf) {
			length = len(srcBuf) - fromAddr
		}
		if toAddr+length > len(dstBuf) {
			length = len(dstBuf) - toAddr
		}

		copy(dstBuf[toAddr:toAddr+length], srcBuf[fromAddr:fromAddr+length])
		memBuffersMu.Unlock()

		return node.Result{
			Success: true,
			Data: map[string]any{
				"from":      from,
				"to":        to,
				"from_addr": fromAddr,
				"to_addr":   toAddr,
				"length":    length,
			},
			Status: fmt.Sprintf("copied %d bytes", length),
		}, nil

	case "dump":
		if name == "" {
			return node.Result{}, fmt.Errorf("buffer name is required")
		}

		addr := 0
		if a, ok := config["address"].(float64); ok {
			addr = int(a)
		}
		length := 64
		if l, ok := config["length"].(float64); ok {
			length = int(l)
		}

		memBuffersMu.RLock()
		buf, exists := memBuffers[name]
		if !exists {
			memBuffersMu.RUnlock()
			return node.Result{}, fmt.Errorf("buffer %s not found", name)
		}

		if addr >= len(buf) {
			memBuffersMu.RUnlock()
			return node.Result{}, fmt.Errorf("address %d out of bounds", addr)
		}

		end := addr + length
		if end > len(buf) {
			end = len(buf)
		}

		data := make([]byte, end-addr)
		copy(data, buf[addr:end])
		memBuffersMu.RUnlock()

		// Format as hex dump
		hexDump := formatHexDump(data, addr)

		return node.Result{
			Success: true,
			Data: map[string]any{
				"hex":    fmt.Sprintf("%X", data),
				"base64": base64.StdEncoding.EncodeToString(data),
				"bytes":  toIntSlice(data),
				"dump":   hexDump,
			},
			Status: fmt.Sprintf("dumped %d bytes", len(data)),
		}, nil

	case "list":
		memBuffersMu.RLock()
		names := make([]string, 0, len(memBuffers))
		info := make(map[string]int)
		for k, v := range memBuffers {
			names = append(names, k)
			info[k] = len(v)
		}
		memBuffersMu.RUnlock()

		return node.Result{
			Success: true,
			Data: map[string]any{
				"buffers": names,
				"sizes":   info,
				"count":   len(names),
			},
			Status: fmt.Sprintf("%d buffers", len(names)),
		}, nil

	default:
		return node.Result{}, fmt.Errorf("unknown operation: %s", operation)
	}
}

func getAddress(config map[string]any) (int, error) {
	if a, ok := config["address"].(float64); ok {
		return int(a), nil
	}
	if a, ok := config["addr"].(float64); ok {
		return int(a), nil
	}
	return 0, fmt.Errorf("address is required")
}

func getValue(config map[string]any) (int, error) {
	if v, ok := config["value"].(float64); ok {
		return int(v), nil
	}
	return 0, fmt.Errorf("value is required")
}

func toIntSlice(data []byte) []int {
	result := make([]int, len(data))
	for i, b := range data {
		result[i] = int(b)
	}
	return result
}

func formatHexDump(data []byte, startAddr int) string {
	var result string
	for i := 0; i < len(data); i += 16 {
		// Address
		result += fmt.Sprintf("%08X  ", startAddr+i)

		// Hex bytes
		for j := 0; j < 16; j++ {
			if i+j < len(data) {
				result += fmt.Sprintf("%02X ", data[i+j])
			} else {
				result += "   "
			}
			if j == 7 {
				result += " "
			}
		}

		// ASCII
		result += " |"
		for j := 0; j < 16 && i+j < len(data); j++ {
			b := data[i+j]
			if b >= 32 && b < 127 {
				result += string(b)
			} else {
				result += "."
			}
		}
		result += "|\n"
	}
	return result
}

// ResetMemory clears all memory buffers (for testing)
func ResetMemory() {
	memBuffersMu.Lock()
	memBuffers = make(map[string][]byte)
	memBuffersMu.Unlock()
}
