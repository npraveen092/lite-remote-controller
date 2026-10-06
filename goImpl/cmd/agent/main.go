package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/exec"
    "runtime"
    "strings"
    "time"

    "github.com/gorilla/websocket"
    "github.com/pion/webrtc/v4"

    "github.com/npraveen092/lite-remote-controller/internal/protocol"
)

func main() {
    server := flag.String("server", "ws://localhost:8080/ws", "signaling WebSocket endpoint")
    session := flag.String("session", "", "shared session token")
    stun := flag.String("stun", "stun:stun.cloudflare.com:3478", "STUN URL; empty disables STUN")
    turn := flag.String("turn", "", "TURN URL; optional")
    turnUser := flag.String("turn-user", "", "TURN username; optional")
    turnCredential := flag.String("turn-credential", "", "TURN credential; optional")
    flag.Parse()

    if *session == "" {
        log.Fatal("--session is required")
    }
    if runtime.GOOS != "windows" {
        log.Printf("warning: this agent is intended for Windows; current OS=%s", runtime.GOOS)
    }

    signalingURL := fmt.Sprintf("%s?role=agent", *server)
    headers := http.Header{"Authorization": []string{"Bearer " + *session}}
    signalConn, _, err := websocket.DefaultDialer.Dial(signalingURL, headers)
    if err != nil {
        log.Fatalf("signaling connection failed: %v", err)
    }
    defer signalConn.Close()

    config := webrtc.Configuration{}
    if *stun != "" {
        config.ICEServers = append(config.ICEServers, webrtc.ICEServer{URLs: []string{*stun}})
    }
    if *turn != "" {
        config.ICEServers = append(config.ICEServers, webrtc.ICEServer{
            URLs:       []string{*turn},
            Username:   *turnUser,
            Credential: *turnCredential,
        })
    }

    pc, err := webrtc.NewPeerConnection(config)
    if err != nil {
        log.Fatalf("peer connection failed: %v", err)
    }
    defer pc.Close()

    pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
        log.Printf("WebRTC state: %s", state.String())
    })

    pc.OnDataChannel(func(dc *webrtc.DataChannel) {
        log.Printf("Data channel received: %s", dc.Label())

        dc.OnOpen(func() {
            log.Println("Controller connected.")
            sendResponse(dc, protocol.ControlMessage{
                Type:      "AGENT_READY",
                RequestID: requestID(),
                Payload: map[string]any{
                    "message":  "agent is ready",
                    "platform": runtime.GOOS,
                    "hostname": hostname(),
                },
            })
        })

        dc.OnMessage(func(msg webrtc.DataChannelMessage) {
            var control protocol.ControlMessage
            if err := json.Unmarshal(msg.Data, &control); err != nil {
                sendResponse(dc, protocol.ControlMessage{
                    Type:      "ERROR",
                    RequestID: requestID(),
                    Payload:   map[string]string{"error": "invalid control message"},
                })
                return
            }
            handleControl(dc, control)
        })
    })

    for {
        _, payload, err := signalConn.ReadMessage()
        if err != nil {
            log.Fatalf("signaling read failed: %v", err)
        }

        var env protocol.Envelope
        if err := json.Unmarshal(payload, &env); err != nil {
            log.Printf("ignored invalid signaling message: %v", err)
            continue
        }

        switch env.Type {
        case "ready":
            log.Println("Signaling channel ready; waiting for controller.")
        case "signal":
            var signal protocol.Signal
            if err := json.Unmarshal(env.Data, &signal); err != nil {
                log.Printf("invalid signal: %v", err)
                continue
            }
            if signal.Kind != "offer" {
                continue
            }

            if err := pc.SetRemoteDescription(webrtc.SessionDescription{
                Type: webrtc.SDPTypeOffer,
                SDP:  signal.SDP,
            }); err != nil {
                log.Fatalf("set remote description failed: %v", err)
            }

            answer, err := pc.CreateAnswer(nil)
            if err != nil {
                log.Fatalf("answer creation failed: %v", err)
            }

            gatherComplete := webrtc.GatheringCompletePromise(pc)
            if err := pc.SetLocalDescription(answer); err != nil {
                log.Fatalf("set local description failed: %v", err)
            }
            <-gatherComplete

            local := pc.LocalDescription()
            if local == nil {
                log.Fatal("local description is empty")
            }

            if err := writeSignal(signalConn, protocol.Signal{Kind: "answer", SDP: local.SDP}); err != nil {
                log.Fatalf("answer signaling failed: %v", err)
            }
        case "error":
            log.Fatalf("signaling error: %s", string(env.Data))
        }
    }
}

func writeSignal(conn *websocket.Conn, signal protocol.Signal) error {
    payload, err := json.Marshal(signal)
    if err != nil {
        return err
    }
    return conn.WriteJSON(protocol.Envelope{Type: "signal", Data: payload})
}

func handleControl(dc *webrtc.DataChannel, control protocol.ControlMessage) {
    switch control.Type {
    case "PING":
        sendResponse(dc, protocol.ControlMessage{
            Type:      "PONG",
            RequestID: control.RequestID,
            Payload:   map[string]any{"time": time.Now().UTC().Format(time.RFC3339Nano)},
        })
    case "INFO":
        sendResponse(dc, protocol.ControlMessage{
            Type:      "INFO_RESULT",
            RequestID: control.RequestID,
            Payload: map[string]string{
                "os":       runtime.GOOS,
                "arch":     runtime.GOARCH,
                "hostname": hostname(),
            },
        })
    case "POWERSHELL":
        payload, ok := control.Payload.(map[string]any)
        if !ok {
            sendResponse(dc, protocol.ControlMessage{
                Type:      "ERROR",
                RequestID: control.RequestID,
                Payload:   map[string]string{"error": "invalid powershell payload"},
            })
            return
        }

        command, _ := payload["command"].(string)
        command = strings.TrimSpace(command)
        if command == "" {
            sendResponse(dc, protocol.ControlMessage{
                Type:      "ERROR",
                RequestID: control.RequestID,
                Payload:   map[string]string{"error": "command is required"},
            })
            return
        }

        if runtime.GOOS != "windows" {
            sendResponse(dc, protocol.ControlMessage{
                Type:      "ERROR",
                RequestID: control.RequestID,
                Payload:   map[string]string{"error": "PowerShell control is only implemented for Windows"},
            })
            return
        }

        cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", command)
        output, err := cmd.CombinedOutput()
        sendResponse(dc, protocol.ControlMessage{
            Type:      "POWERSHELL_RESULT",
            RequestID: control.RequestID,
            Payload: map[string]any{
                "success": err == nil,
                "output":  string(output),
            },
        })
    case "CLOSE":
        _ = dc.Close()
    default:
        sendResponse(dc, protocol.ControlMessage{
            Type:      "ERROR",
            RequestID: control.RequestID,
            Payload:   map[string]string{"error": "unsupported control type"},
        })
    }
}

func sendResponse(dc *webrtc.DataChannel, message protocol.ControlMessage) {
    payload, err := json.Marshal(message)
    if err != nil {
        return
    }
    if err := dc.Send(payload); err != nil {
        log.Printf("response send failed: %v", err)
    }
}

func hostname() string {
    value, err := os.Hostname()
    if err != nil {
        return "unknown"
    }
    return value
}

func requestID() string {
    return time.Now().UTC().Format("20060102T150405.000000000")
}
