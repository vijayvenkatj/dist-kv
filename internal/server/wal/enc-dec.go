package wal

import (
	"encoding/binary"
	"fmt"
)

func encode(e *LogEntry) []byte {

	op := []byte(e.Operation)
	key := []byte(e.Key)
	val := []byte(e.Value)

	total := 20 + len(op) + len(key) + len(val)
	buf := make([]byte, total)

	binary.LittleEndian.PutUint32(buf[0:4], e.Term)
	binary.LittleEndian.PutUint32(buf[4:8], e.LogIndex)
	binary.LittleEndian.PutUint32(buf[8:12], uint32(len(op)))
	binary.LittleEndian.PutUint32(buf[12:16], uint32(len(key)))
	binary.LittleEndian.PutUint32(buf[16:20], uint32(len(val)))

	offset := 20
	copy(buf[offset:], op)
	offset += len(op)

	copy(buf[offset:], key)
	offset += len(key)

	copy(buf[offset:], val)

	return buf
}

func decode(data []byte) (*LogEntry, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("invalid header")
	}

	offset := 0

	term := binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	idx := binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	opLen := binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	keyLen := binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	valLen := binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	total := int(opLen + keyLen + valLen)
	if len(data) < 20+total {
		return nil, fmt.Errorf("incomplete entry")
	}

	op := string(data[offset : offset+int(opLen)])
	offset += int(opLen)

	key := string(data[offset : offset+int(keyLen)])
	offset += int(keyLen)

	value := string(data[offset : offset+int(valLen)])

	return &LogEntry{
		Term:      term,
		LogIndex:  idx,
		Operation: op,
		Key:       key,
		Value:     value,
	}, nil
}
