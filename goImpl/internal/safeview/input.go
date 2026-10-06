package safeview

import "log"

type InputEvent struct {
    Kind   string  `json:"kind"`
    Action string  `json:"action"`
    X      float64 `json:"x,omitempty"`
    Y      float64 `json:"y,omitempty"`
    Button string  `json:"button,omitempty"`
    Key    string  `json:"key,omitempty"`
}

// HandleInputEvent deliberately does not inject input into the host OS.
// It provides a safe end-to-end transport harness for validating the protocol.
func HandleInputEvent(event InputEvent) {
    log.Printf("input event received: kind=%s action=%s x=%.2f y=%.2f button=%s key=%s",
        event.Kind, event.Action, event.X, event.Y, event.Button, event.Key)
}