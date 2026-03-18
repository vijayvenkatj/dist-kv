package wal

import "encoding/binary"

func encode(data *LogEntry) []byte {

	operationBytes, keyBytes, valBytes := []byte(data.Operation), []byte(data.Key), []byte(data.Value)

	operationLen, keyLen, valLen := uint32(len(operationBytes)), uint32(len(keyBytes)), uint32(len(valBytes))
	total := 12 + operationLen + keyLen + valLen

	entry := make([]byte, total)

	temp := make([]byte, 4)

	binary.LittleEndian.PutUint32(temp, operationLen)
	entry = append(temp, operationBytes...)
	binary.LittleEndian.PutUint32(temp, keyLen)
	entry = append(temp, keyBytes...)
	binary.LittleEndian.PutUint32(temp, valLen)
	entry = append(temp, valBytes...)

	entry = append(entry, operationBytes...)
	entry = append(entry, keyBytes...)
	entry = append(entry, valBytes...)

	return entry
}

func decode(data []byte) (*LogEntry, error) {
	offset := 0

	opLen := binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	keyLen := binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	valLen := binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	op := string(data[offset : offset+int(opLen)])
	offset += int(opLen)

	key := string(data[offset : offset+int(keyLen)])
	offset += int(keyLen)

	value := string(data[offset : offset+int(valLen)])

	return &LogEntry{
		Operation: op,
		Key:       key,
		Value:     value,
	}, nil
}
