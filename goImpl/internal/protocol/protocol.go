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

type InputEvent struct {
    Kind   string  `json:"kind"`
    Action string  `json:"action,omitempty"`
    X      float64 `json:"x,omitempty"`
    Y      float64 `json:"y,omitempty"`
    Button string  `json:"button,omitempty"`
    Key    string  `json:"key,omitempty"`
}

const (
    InputMouse    = "mouse"
    InputKeyboard = "keyboard"
)
