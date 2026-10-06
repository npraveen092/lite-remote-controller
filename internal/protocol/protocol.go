package protocol

import "encoding/json"

type Envelope struct {
    Type string          `json:"type"`
    Data json.RawMessage `json:"data,omitempty"`
}

type Signal struct {
    Kind string `json:"kind"`
    SDP  string `json:"sdp,omitempty"`
}

type ControlMessage struct {
    Type      string `json:"type"`
    RequestID string `json:"requestId,omitempty"`
    Payload   any    `json:"payload,omitempty"`
}
