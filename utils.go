package webterm

import (
	"encoding/base64"
	"encoding/binary"
	"os"
	"sync"
	"time"
)

var (
	idMutex   sync.Mutex
	lastTime  int64
	sequence  uint16
	machineID uint16
)

func init() {
	// Use process ID as machine identifier (modulo to fit in 10 bits)
	machineID = uint16(os.Getpid() % 1024)
}

func genID() string {
	idMutex.Lock()
	defer idMutex.Unlock()

	// Get current timestamp in milliseconds
	now := time.Now().UnixMilli()

	// If same millisecond, increment sequence
	if now == lastTime {
		sequence = (sequence + 1) & 0xFFF // 12 bits for sequence
		// If sequence overflows, wait for next millisecond
		if sequence == 0 {
			for now <= lastTime {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		sequence = 0
	}
	lastTime = now

	// Snowflake-like structure (64 bits):
	// 41 bits: timestamp (milliseconds since custom epoch)
	// 10 bits: machine/process ID
	// 12 bits: sequence number
	// 1 bit: unused (sign bit)

	// Use a custom epoch (2020-01-01) to save bits
	const customEpoch = 1577836800000 // 2020-01-01 00:00:00 UTC in milliseconds
	timestamp := now - customEpoch

	// Construct the ID
	id := uint64(timestamp&0x1FFFFFFFFFF) << 22 // 41 bits timestamp
	id |= uint64(machineID&0x3FF) << 12         // 10 bits machine ID
	id |= uint64(sequence & 0xFFF)              // 12 bits sequence

	// Encode to base64 for shorter string representation
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, id)

	// Use URL-safe base64 encoding without padding for shorter IDs
	return base64.RawURLEncoding.EncodeToString(buf)
}
