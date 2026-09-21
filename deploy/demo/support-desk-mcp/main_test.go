package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListTickets_ReturnsSeedData(t *testing.T) {
	s := NewServer()
	handler := s.Handler()

	req := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_tickets","arguments":{}}}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString(req))
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp jsonrpcResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
}

func TestUpdateTicketStatus_MutatesState(t *testing.T) {
	s := NewServer()
	handler := s.Handler()

	req := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"update_ticket_status","arguments":{"ticket_id":"TCK-1001","new_status":"resolved"}}}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString(req))
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	s.mu.Lock()
	got := s.tickets["TCK-1001"].Status
	s.mu.Unlock()
	if got != "resolved" {
		t.Fatalf("expected ticket status to be updated to resolved, got %q", got)
	}
}

func TestUpdateTicketStatus_UnknownTicketErrors(t *testing.T) {
	s := NewServer()
	handler := s.Handler()

	req := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"update_ticket_status","arguments":{"ticket_id":"NO-SUCH-TICKET","new_status":"resolved"}}}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString(req))
	handler.ServeHTTP(w, r)

	var resp jsonrpcResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Error == nil {
		t.Fatalf("expected an error for an unknown ticket id")
	}
}

func TestDeleteTicket_RemovesFromState(t *testing.T) {
	s := NewServer()
	handler := s.Handler()

	req := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"delete_ticket","arguments":{"ticket_id":"TCK-1001"}}}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString(req))
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	s.mu.Lock()
	_, stillExists := s.tickets["TCK-1001"]
	s.mu.Unlock()
	if stillExists {
		t.Fatalf("expected TCK-1001 to be removed after delete_ticket")
	}
}

func TestReset_RestoresSeedData(t *testing.T) {
	s := NewServer()

	s.mu.Lock()
	delete(s.tickets, "TCK-1001")
	s.mu.Unlock()

	s.reset()

	s.mu.Lock()
	_, exists := s.tickets["TCK-1001"]
	invocationCount := len(s.invocations)
	s.mu.Unlock()

	if !exists {
		t.Fatalf("expected reset to restore TCK-1001")
	}
	if invocationCount != 0 {
		t.Fatalf("expected reset to clear invocation history, got %d entries", invocationCount)
	}
}

func TestUnknownTool_ReturnsError(t *testing.T) {
	s := NewServer()
	handler := s.Handler()

	req := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"not_a_real_tool","arguments":{}}}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString(req))
	handler.ServeHTTP(w, r)

	var resp jsonrpcResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Error == nil {
		t.Fatalf("expected an error for an unknown tool")
	}
}
