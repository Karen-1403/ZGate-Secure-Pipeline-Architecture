package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net"
	"time"

	"github.com/Karen-1403/ZGate-Secure-Pipeline-Architecture/config"
	"github.com/Karen-1403/ZGate-Secure-Pipeline-Architecture/internal/mvc"
	"github.com/Karen-1403/ZGate-Secure-Pipeline-Architecture/internal/pipeline"
	"github.com/Karen-1403/ZGate-Secure-Pipeline-Architecture/internal/protocol"
	"github.com/Karen-1403/ZGate-Secure-Pipeline-Architecture/internal/transport"
)

func main() {
	// Initialize logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	tlsConfig, err := transport.NewMTLSConfig(config.ServerCertPath, config.ServerKeyPath, config.CACertPath)
	if err != nil {
		log.Fatalf("failed to create tls config: %v", err)
	}

	listener, err := tls.Listen("tcp", config.ListenAddr, tlsConfig)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	log.Printf("Zero-trust Proxy listening on %s", config.ListenAddr)

	// Initialize pipeline and MVC controller
	pl := pipeline.NewQueryProcessingPipeline()
	ctrl := mvc.NewController(pl)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}

		go handleConn(conn, ctrl)
	}
}

func handleConn(c net.Conn, ctrl *mvc.Controller) {
	defer c.Close()

	r := bufio.NewReader(c)
	w := bufio.NewWriter(c)
	proto := protocol.NewJSONProtocol(r, w)

	// Simple loop - read JSONL messages and process them
	for {
		msgRaw, err := proto.ReadMessage()
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Printf("read message error: %v", err)
			return
		}

		// Convert raw message to map for flexibility
		var msg map[string]interface{}
		if err := json.Unmarshal(msgRaw, &msg); err != nil {
			log.Printf("invalid json message: %v", err)
			proto.WriteMessage(map[string]interface{}{"status": "error", "error": "invalid json"})
			continue
		}

		// Pass to controller which uses pipeline and models
		ctx, resp := ctrl.HandleRequest(context.Background(), msg)

		// attach optionally time
		if resp == nil {
			resp = map[string]interface{}{"status": "error", "error": "internal_error"}
		}

		// add metadata
		if m, ok := resp.(map[string]interface{}); ok {
			m["processed_at"] = time.Now().UTC().Format(time.RFC3339)
			// include which user if present
			if ctx != nil && ctx.User != "" {
				m["user"] = ctx.User
			}
		}

		if err := proto.WriteMessage(resp); err != nil {
			log.Printf("write response error: %v", err)
			return
		}
	}
}
