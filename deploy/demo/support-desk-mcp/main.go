// Command support-desk-mcp is a small, realistic (non-toy) MCP backend for
// AgentGate's G7 Task B demonstration system. It exposes three tools with
// genuine business meaning and deliberately mixed risk classifications —
// read, write, destructive — so the governance story (deny-by-default,
// policy activation, rollback) is visibly meaningful rather than abstract:
//
//   - list_tickets          (read)        — never mutates state
//   - update_ticket_status  (write)       — mutates an existing ticket
//   - delete_ticket         (destructive) — permanently removes a ticket
//
// This is additive to deploy/g6 (deploy/demo/), not a replacement — see
// deploy/demo/README.md. Modeled structurally on deploy/g6/probe-mcp (same
// JSON-RPC MCP shape, same inspection endpoints) but with business-meaningful
// tools and state instead of a single always-succeeding counter.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Ticket is the demo's only piece of durable state — deliberately simple
// in-memory storage, reset on process restart (this is a demo backend, not
// a production one).
type Ticket struct {
	ID      string `json:"id"`
	Subject string `json:"subject"`
	Status  string `json:"status"` // open | in_progress | resolved
}

// InvocationRecord mirrors probe-mcp's inspection shape, for the same
// "what actually reached the backend" verification the G6 evidence pattern
// established.
type InvocationRecord struct {
	Timestamp  string            `json:"timestamp"`
	ToolName   string            `json:"tool_name"`
	Arguments  map[string]any    `json:"arguments"`
	Headers    map[string]string `json:"headers"`
	RawRequest string            `json:"raw_request"`
}

type Server struct {
	mu          sync.Mutex
	tickets     map[string]*Ticket
	invocations []InvocationRecord
}

func NewServer() *Server {
	s := &Server{
		tickets: make(map[string]*Ticket),
	}
	s.seed()
	return s
}

// seed populates realistic starting data — always reset to this exact set on
// every /_demo/reset call, so scenario runs are reproducible.
func (s *Server) seed() {
	s.tickets = map[string]*Ticket{
		"TCK-1001": {ID: "TCK-1001", Subject: "Password reset request", Status: "open"},
		"TCK-1002": {ID: "TCK-1002", Subject: "Billing discrepancy on invoice #4471", Status: "in_progress"},
		"TCK-1003": {ID: "TCK-1003", Subject: "Feature request: dark mode", Status: "open"},
	}
}

func (s *Server) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seed()
	s.invocations = nil
}

func (s *Server) recordInvocation(rec InvocationRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.invocations = append(s.invocations, rec)
}

func (s *Server) snapshot() (tickets []*Ticket, invocations []InvocationRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.tickets {
		tCopy := *t
		tickets = append(tickets, &tCopy)
	}
	invocations = append(invocations, s.invocations...)
	return tickets, invocations
}

type jsonrpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonrpcResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      any           `json:"id"`
	Result  any           `json:"result,omitempty"`
	Error   *jsonrpcError `json:"error,omitempty"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var toolSchemas = []map[string]any{
	{
		"name":        "list_tickets",
		"description": "List support tickets, optionally filtered by status. Read-only.",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"status_filter": map[string]any{"type": "string"},
			},
		},
	},
	{
		"name":        "update_ticket_status",
		"description": "Update an existing ticket's status. Mutates state.",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"ticket_id":  map[string]any{"type": "string"},
				"new_status": map[string]any{"type": "string"},
			},
			"required": []string{"ticket_id", "new_status"},
		},
	},
	{
		"name":        "delete_ticket",
		"description": "Permanently delete a ticket. Irreversible.",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"ticket_id": map[string]any{"type": "string"},
			},
			"required": []string{"ticket_id"},
		},
	},
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Inspection endpoints — same convention as deploy/g6/probe-mcp's
	// /_g6/* routes, under /_demo/* to avoid ambiguity between topologies.
	mux.HandleFunc("/_demo/state", func(w http.ResponseWriter, r *http.Request) {
		tickets, invocations := s.snapshot()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tickets":     tickets,
			"invocations": invocations,
			"count":       len(invocations),
		})
	})

	mux.HandleFunc("/_demo/reset", func(w http.ResponseWriter, r *http.Request) {
		s.reset()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("reset"))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var req jsonrpcRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			writeJSON(w, jsonrpcResponse{JSONRPC: "2.0", Error: &jsonrpcError{Code: -32700, Message: "parse error: " + err.Error()}})
			return
		}

		headers := make(map[string]string)
		for k, v := range r.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		switch req.Method {
		case "initialize":
			writeJSON(w, jsonrpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]any{
					"protocolVersion": "2026-07-28",
					"serverInfo":      map[string]any{"name": "support-desk-mcp", "version": "1.0.0"},
					"capabilities":    map[string]any{"tools": map[string]any{}},
				},
			})

		case "tools/list":
			writeJSON(w, jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"tools": toolSchemas}})

		case "tools/call":
			s.handleToolCall(w, req, bodyBytes, headers)

		default:
			writeJSON(w, jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonrpcError{Code: -32601, Message: fmt.Sprintf("method %q not found", req.Method)}})
		}
	})

	return mux
}

func (s *Server) handleToolCall(w http.ResponseWriter, req jsonrpcRequest, rawBody []byte, headers map[string]string) {
	var callParams struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &callParams); err != nil {
		writeJSON(w, jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonrpcError{Code: -32602, Message: "invalid params"}})
		return
	}

	s.recordInvocation(InvocationRecord{
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
		ToolName:   callParams.Name,
		Arguments:  callParams.Arguments,
		Headers:    headers,
		RawRequest: string(rawBody),
	})

	switch callParams.Name {
	case "list_tickets":
		s.mu.Lock()
		statusFilter, _ := callParams.Arguments["status_filter"].(string)
		var out []*Ticket
		for _, t := range s.tickets {
			if statusFilter == "" || t.Status == statusFilter {
				tCopy := *t
				out = append(out, &tCopy)
			}
		}
		s.mu.Unlock()
		writeJSON(w, jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{
			"content": []map[string]any{{"type": "text", "text": fmt.Sprintf("Found %d ticket(s)", len(out))}},
			"tickets": out,
		}})

	case "update_ticket_status":
		ticketID, _ := callParams.Arguments["ticket_id"].(string)
		newStatus, _ := callParams.Arguments["new_status"].(string)
		s.mu.Lock()
		t, ok := s.tickets[ticketID]
		if ok {
			t.Status = strings.TrimSpace(newStatus)
		}
		s.mu.Unlock()
		if !ok {
			writeJSON(w, jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonrpcError{Code: -32000, Message: "ticket not found"}})
			return
		}
		writeJSON(w, jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{
			"content": []map[string]any{{"type": "text", "text": fmt.Sprintf("Ticket %s status updated to %s", ticketID, newStatus)}},
		}})

	case "delete_ticket":
		ticketID, _ := callParams.Arguments["ticket_id"].(string)
		s.mu.Lock()
		_, ok := s.tickets[ticketID]
		if ok {
			delete(s.tickets, ticketID)
		}
		s.mu.Unlock()
		if !ok {
			writeJSON(w, jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonrpcError{Code: -32000, Message: "ticket not found"}})
			return
		}
		writeJSON(w, jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{
			"content": []map[string]any{{"type": "text", "text": fmt.Sprintf("Ticket %s permanently deleted", ticketID)}},
		}})

	default:
		writeJSON(w, jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonrpcError{Code: -32601, Message: fmt.Sprintf("tool %q not found", callParams.Name)}})
	}
}

func writeJSON(w http.ResponseWriter, resp jsonrpcResponse) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func main() {
	port := os.Getenv("DEMO_MCP_PORT")
	if port == "" {
		port = "9200"
	}

	server := NewServer()
	log.Printf("support-desk-mcp listening on :%s (3 tools: list_tickets [read], update_ticket_status [write], delete_ticket [destructive])", port)
	if err := http.ListenAndServe(":"+port, server.Handler()); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
