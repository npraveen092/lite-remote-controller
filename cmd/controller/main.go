package main

import (
    "bufio"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/gorilla/websocket"
    "github.com/pion/webrtc/v4"

    "github.com/npraveen092/lite-remote-controller/internal/protocol"
)

func main() {
    server := flag.String("server", "ws://localhost:8080/ws", "signaling WebSocket endpoint")
    session := flag.String("session", "", "shared session token")
    stun := flag.String("stun", "stun:stun.cloudflare.com:3478", "STUN URL")
    flag.Parse()

    if *session == "" {
        log.Fatal("--session is required")
    }

    signalingURL := fmt.Sprintf("%s?role=controller", *server)
    headers := http.Header{"Authorization": []string{"Bearer " + *session}}
    signalConn, _, err := websocket.DefaultDialer.Dial(signalingURL, headers)
    if err != nil {
        log.Fatalf("signaling connection failed: %v", err)
    }
    defer signalConn.Close()

    config := webrtc.Configuration{}
    if *stun != "" {
        config.ICEServers = []webrtc.ICEServer{{URLs: []string{*stun}}}
    }

    pc, err := webrtc.NewPeerConnection(config)
    if err != nil {
        log.Fatalf("peer connection failed: %v", err)
    }
    defer pc.Close()

    pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
        log.Printf("WebRTC state: %s", state.String())
    })

    opened := make(chan struct{})
    dataChannel, err := pc.CreateDataChannel("control", nil)
    if err != nil {
        log.Fatalf("data channel creation failed: %v", err)
    }

    dataChannel.OnOpen(func() {
        select {
        case <-opened:
        default:
            close(opened)
        }
        log.Println("Remote control channel is open.")
    })

    dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
        var response map[string]any
        if err := json.Unmarshal(msg.Data, &response); err == nil {
            encoded, _ := json.MarshalIndent(response, "", "  ")
            fmt.Println(string(encoded))
            return
        }
        fmt.Println(string(msg.Data))
    })

    offer, err := pc.CreateOffer(nil)
    if err != nil {
        log.Fatalf("offer creation failed: %v", err)
    }

    gatherComplete := webrtc.GatheringCompletePromise(pc)
    if err := pc.SetLocalDescription(offer); err != nil {
        log.Fatalf("set local description failed: %v", err)
    }
    <-gatherComplete

    local := pc.LocalDescription()
    if local == nil {
        log.Fatal("local description is empty")
    }

    if err := writeSignal(signalConn, protocol.Signal{Kind: "offer", SDP: local.SDP}); err != nil {
        log.Fatalf("offer signaling failed: %v", err)
    }

    answered := false
    for !answered {
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
            log.Println("Signaling channel ready; waiting for remote peer.")
        case "signal":
            var signal protocol.Signal
            if err := json.Unmarshal(env.Data, &signal); err != nil {
                log.Printf("invalid signal: %v", err)
                continue
            }
            if signal.Kind == "answer" {
                if err := pc.SetRemoteDescription(webrtc.SessionDescription{
                    Type: webrtc.SDPTypeAnswer,
                    SDP:  signal.SDP,
                }); err != nil {
                    log.Fatalf("set remote description failed: %v", err)
                }
                answered = true
            }
        case "error":
            log.Fatalf("signaling error: %s", string(env.Data))
        }
    }

    select {
    case <-opened:
        runConsole(dataChannel)
    case <-time.After(30 * time.Second):
        log.Fatal("timed out waiting for WebRTC data channel")
    }
}

func writeSignal(conn *websocket.Conn, signal protocol.Signal) error {
    payload, err := json.Marshal(signal)
    if err != nil {
        return err
    }
    return conn.WriteJSON(protocol.Envelope{Type: "signal", Data: payload})
}

func runConsole(dc *webrtc.DataChannel) {
    reader := bufio.NewReader(os.Stdin)

    fmt.Println("Commands: info | ping | ps <PowerShell> | quit")
    for {
        fmt.Print("remote> ")
        line, err := reader.ReadString('
')
        if err != nil {
            return
        }

        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }

        switch {
        case line == "info":
            sendControl(dc, protocol.ControlMessage{Type: "INFO", RequestID: requestID()})
        case line == "ping":
            sendControl(dc, protocol.ControlMessage{Type: "PING", RequestID: requestID()})
        case line == "quit":
            sendControl(dc, protocol.ControlMessage{Type: "CLOSE", RequestID: requestID()})
            return
        case strings.HasPrefix(line, "ps "):
            command := strings.TrimSpace(strings.TrimPrefix(line, "ps "))
            if command == "" {
                fmt.Println("usage: ps <PowerShell command>")
                continue
            }
            sendControl(dc, protocol.ControlMessage{
                Type:      "POWERSHELL",
                RequestID: requestID(),
                Payload:   map[string]string{"command": command},
            })
        default:
            fmt.Println("unknown command")
        }
    }
}

func sendControl(dc *webrtc.DataChannel, message protocol.ControlMessage) {
    payload, err := json.Marshal(message)
    if err != nil {
        log.Printf("encode control message: %v", err)
        return
    }
    if err := dc.Send(payload); err != nil {
        log.Printf("send failed: %v", err)
    }
}

func requestID() string {
    return time.Now().UTC().Format("20060102T150405.000000000")
}
