package safeview

import "testing"

func TestInputEventHarness(t *testing.T) {
    HandleInputEvent(InputEvent{Kind: "mouse", Action: "move", X: 10, Y: 20})
    HandleInputEvent(InputEvent{Kind: "keyboard", Action: "down", Key: "A"})
}