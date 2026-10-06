package main

import (
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "strings"
    "sync"

    "github.com/gorilla/websocket"
)

const (
    roleAgent      = "agent"
    roleController = "controller"
)

type peer struct {
    role string
    conn *websocket.Conn
    mu   sync.Mutex
}

type room struct {
    mu    sync.Mutex
    peers map[string]*peer
}

var (
    roomsMu sync.Mutex
    rooms   = map[string]*room{}

    upgrader = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool { return true },
    }
)

type envelope struct {
    Type string          `json:"type"`
    Data json.RawMessage `json:"data,omitempty"`
}

func getOrCreateRoom(session string) *room {
    roomsMu.Lock()
    defer roomsMu.Unlock()
    if r, ok := rooms[session]; ok {
        return r
    }
    r := &room{peers: make(map[string]*peer)}
    rooms[session] = r
    return r
}

func cleanupRoom(session string, p *peer) {
    roomsMu.Lock()
    r, ok := rooms[session]
    roomsMu.Unlock()
    if !ok {
        return
    }

    r.mu.Lock()
    defer r.mu.Unlock()

    if current, exists := r.peers[p.role]; exists && current == p {
        delete(r.peers, p.role)
    }
    if len(r.peers) == 0 {
        roomsMu.Lock()
        delete(rooms, session)
        roomsMu.Unlock()
    }
}

func sendJSON(p *peer, payload any) error {
    p.mu.Lock()
    defer p.mu.Unlock()
    return p.conn.WriteJSON(payload)
}

func handleWS(w http.ResponseWriter, r *http.Request) {
    role := r.URL.Query().Get("role")
    const prefix = "Bearer "
    auth := r.Header.Get("Authorization")
    if !strings.HasPrefix(auth, prefix) {
        http.Error(w, "authorization required", http.StatusUnauthorized)
        return
    }
    session := strings.TrimSpace(strings.TrimPrefix(auth, prefix))

    if session == "" || (role != roleAgent && role != roleController) {
        http.Error(w, "session and role are required", http.StatusBadRequest)
        return
    }
    if len(session) < 16 || len(session) > 128 {
        http.Error(w, "invalid session", http.StatusBadRequest)
        return
    }

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("upgrade failed: %v", err)
        return
    }
    defer conn.Close()

    p := &peer{role: role, conn: conn}
    room := getOrCreateRoom(session)

    room.mu.Lock()
    if len(room.peers) >= 2 || room.peers[role] != nil {
        room.mu.Unlock()
        _ = sendJSON(p, envelope{
            Type: "error",
            Data: json.RawMessage([]byte("{"message":"session is full"}")),
        })
        return
    }
    room.peers[role] = p
    room.mu.Unlock()

    log.Printf("peer joined session=%s role=%s", session, role)
    defer cleanupRoom(session, p)

    _ = sendJSON(p, envelope{
        Type: "ready",
        Data: json.RawMessage([]byte("{"message":"signaling channel ready"}")),
    })

    for {
        _, payload, err := conn.ReadMessage()
        if err != nil {
            if !websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
                log.Printf("read failed session=%s role=%s: %v", session, role, err)
            }
            return
        }

        var msg envelope
        if err := json.Unmarshal(payload, &msg); err != nil {
            _ = sendJSON(p, envelope{
                Type: "error",
                Data: json.RawMessage([]byte("{"message":"invalid json"}")),
            })
            continue
        }

        switch msg.Type {
        case "signal":
            room.mu.Lock()
            otherRole := roleAgent
            if role == roleAgent {
                otherRole = roleController
            }
            target := room.peers[otherRole]
            room.mu.Unlock()

            if target == nil {
                continue
            }
            if err := sendJSON(target, msg); err != nil {
                log.Printf("forward failed session=%s role=%s: %v", session, role, err)
                return
            }
        case "ping":
            _ = sendJSON(p, envelope{Type: "pong"})
        default:
            _ = sendJSON(p, envelope{
                Type: "error",
                Data: json.RawMessage([]byte("{"message":"unsupported message type"}")),
            })
        }
    }
}

func health(w http.ResponseWriter, _ *http.Request) {
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte("ok"))
}

func randomSession() (string, error) {
    raw := make([]byte, 16)
    if _, err := rand.Read(raw); err != nil {
        return "", err
    }
    return hex.EncodeToString(raw), nil
}

func main() {
    addr := getenv("SIGNALING_ADDR", ":8080")
    mux := http.NewServeMux()
    mux.HandleFunc("/healthz", health)
    mux.HandleFunc("/ws", handleWS)
    mux.HandleFunc("/session", func(w http.ResponseWriter, _ *http.Request) {
        session, err := randomSession()
        if err != nil {
            http.Error(w, "session generation failed", http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "text/plain")
        _, _ = w.Write([]byte(session))
    })

    server := &http.Server{Addr: addr, Handler: mux}
    log.Printf("signaling server listening on %s", addr)

    if cert, key := os.Getenv("TLS_CERT_FILE"), os.Getenv("TLS_KEY_FILE"); cert != "" && key != "" {
        log.Printf("TLS enabled")
        log.Fatal(server.ListenAndServeTLS(cert, key))
        return
    }

    log.Printf("TLS disabled; use a reverse proxy for WSS in public deployments")
    log.Fatal(server.ListenAndServe())
}

func getenv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}
