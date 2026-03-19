package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
)

type Server struct {
	reader io.Reader
	writer io.Writer
}

func NewServer(reader io.Reader, writer io.Writer) *Server {
	return &Server{reader: reader, writer: writer}
}

func (s *Server) Run() error {
	scanner := bufio.NewScanner(s.reader)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			log.Printf("invalid JSON-RPC request: %v", err)
			continue
		}

		resp := s.handleRequest(req)
		respBytes, _ := json.Marshal(resp)
		fmt.Fprintln(s.writer, string(respBytes))
	}

	return scanner.Err()
}

func (s *Server) handleRequest(req JSONRPCRequest) JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]any{
					"tools": map[string]any{},
				},
				"serverInfo": map[string]any{
					"name":    "colign-mcp",
					"version": "0.1.0",
				},
			},
		}

	case "tools/list":
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"tools": ListTools(),
			},
		}

	case "tools/call":
		return s.handleToolCall(req)

	default:
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: -32601, Message: "method not found: " + req.Method},
		}
	}
}

func (s *Server) handleToolCall(req JSONRPCRequest) JSONRPCResponse {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: -32602, Message: "invalid params"},
		}
	}

	// TODO: implement tool handlers with actual service calls
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"content": []map[string]any{
				{"type": "text", "text": fmt.Sprintf("Tool %s called (not yet implemented)", params.Name)},
			},
		},
	}
}
