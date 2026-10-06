package safeview

import (
    "encoding/binary"
    "errors"
)

const (
    HeaderSize = 16
    ChunkSize  = 12000
)

var ErrInvalidFrame = errors.New("invalid screen frame chunk")

func EncodeFrameChunk(frameID uint32, index, count uint16, payload []byte) []byte {
    out := make([]byte, HeaderSize+len(payload))
    copy(out[:4], []byte{'L', 'R', 'S', '1'})
    binary.BigEndian.PutUint32(out[4:8], frameID)
    binary.BigEndian.PutUint16(out[8:10], index)
    binary.BigEndian.PutUint16(out[10:12], count)
    binary.BigEndian.PutUint32(out[12:16], uint32(len(payload)))
    copy(out[HeaderSize:], payload)
    return out
}

func DecodeFrameChunk(data []byte) (uint32, uint16, uint16, []byte, error) {
    if len(data) < HeaderSize || string(data[:4]) != "LRS1" {
        return 0, 0, 0, nil, ErrInvalidFrame
    }
    frameID := binary.BigEndian.Uint32(data[4:8])
    index := binary.BigEndian.Uint16(data[8:10])
    count := binary.BigEndian.Uint16(data[10:12])
    size := binary.BigEndian.Uint32(data[12:16])
    if count == 0 || index >= count || uint32(len(data)-HeaderSize) != size {
        return 0, 0, 0, nil, ErrInvalidFrame
    }
    return frameID, index, count, data[HeaderSize:], nil
}

func ChunkFrame(frameID uint32, frame []byte) [][]byte {
    if len(frame) == 0 {
        return nil
    }
    count := (len(frame) + ChunkSize - 1) / ChunkSize
    if count > 65535 {
        panic("frame too large")
    }
    chunks := make([][]byte, 0, count)
    for i := 0; i < count; i++ {
        start := i * ChunkSize
        end := start + ChunkSize
        if end > len(frame) {
            end = len(frame)
        }
        chunks = append(chunks, EncodeFrameChunk(frameID, uint16(i), uint16(count), frame[start:end]))
    }
    return chunks
}