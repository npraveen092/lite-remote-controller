package safeview

import (
    "bytes"
    "testing"
)

func TestChunkRoundTrip(t *testing.T) {
    frame := bytes.Repeat([]byte("frame"), 10000)
    chunks := ChunkFrame(42, frame)
    if len(chunks) < 2 {
        t.Fatal("expected multiple chunks")
    }
    rebuilt := make([]byte, 0, len(frame))
    for i, chunk := range chunks {
        id, index, count, payload, err := DecodeFrameChunk(chunk)
        if err != nil {
            t.Fatalf("chunk %d: %v", i, err)
        }
        if id != 42 || int(index) != i || int(count) != len(chunks) {
            t.Fatalf("unexpected metadata: id=%d index=%d count=%d", id, index, count)
        }
        rebuilt = append(rebuilt, payload...)
    }
    if !bytes.Equal(frame, rebuilt) {
        t.Fatal("rebuilt frame differs from original")
    }
}