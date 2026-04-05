package mcp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	stdruntime "runtime"
	"strings"
	"testing"
)

func TestExplainEndpoint(t *testing.T) {
	server := NewServer(8080)

	reqBody := `{"command": "npm run build"}`
	req, _ := http.NewRequest("POST", "/explain", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	server.handleExplain(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %v", rr.Code)
	}

	var res map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&res)

	if res["type"] != "build" || res["tool"] != "vite" {
		t.Errorf("Expected build/vite payload, got %v", res)
	}
}

func TestDoctorEndpoint(t *testing.T) {
	server := NewServer(8080)

	req, _ := http.NewRequest("GET", "/doctor", nil)
	rr := httptest.NewRecorder()

	server.handleDoctor(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %v", rr.Code)
	}

	var res map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&res)

	if _, ok := res["node"]; !ok {
		t.Errorf("Expected node key in doctor JSON response")
	}
}

func TestTraceEndpoint(t *testing.T) {
	server := NewServer(8080)

	req, _ := http.NewRequest("GET", "/trace", nil)
	rr := httptest.NewRecorder()

	server.handleTrace(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %v", rr.Code)
	}
}
func TestJSONEncoding(t *testing.T) {
	server := NewServer(8080)

	// Create a dummy request to /doctor to check headers
	req, _ := http.NewRequest("GET", "/doctor", nil)
	rr := httptest.NewRecorder()

	// We need to use the handler that has the middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/doctor", server.handleDoctor)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		mux.ServeHTTP(w, r)
	})

	handler.ServeHTTP(rr, req)

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("Expected application/json; charset=utf-8, got %q", contentType)
	}

	// Check for HTML escaping
	// We can manually call writeJSON with suspicious characters
	rr2 := httptest.NewRecorder()
	server.writeJSON(rr2, http.StatusOK, map[string]string{"sym": "<>&"})

	body := rr2.Body.String()
	if !strings.Contains(body, "<>&") {
		t.Errorf("Expected symbols to be unescaped, got %q", body)
	}
}

func TestParseCommandForExecution(t *testing.T) {
	t.Run("simple command keeps argv split", func(t *testing.T) {
		cmd, args := parseCommandForExecution("go test ./...")
		if cmd != "go" {
			t.Fatalf("expected cmd=go, got %q", cmd)
		}
		if len(args) != 2 || args[0] != "test" || args[1] != "./..." {
			t.Fatalf("unexpected args: %#v", args)
		}
	})

	if stdruntime.GOOS == "windows" {
		t.Run("complex powershell command routed to powershell", func(t *testing.T) {
			raw := `cd C:\Users\Eric\Projects\almathurat; Select-String -Pattern "^version:" -Path pubspec.yaml`
			cmd, args := parseCommandForExecution(raw)

			if cmd != "powershell" && cmd != "pwsh" {
				t.Fatalf("expected powershell wrapper, got cmd=%q args=%#v", cmd, args)
			}
			if len(args) != 4 {
				t.Fatalf("expected 4 wrapper args, got %#v", args)
			}
			if args[0] != "-NoProfile" || args[1] != "-NonInteractive" || args[2] != "-Command" {
				t.Fatalf("unexpected wrapper args: %#v", args)
			}
			if args[3] != raw {
				t.Fatalf("raw command changed during parse: %q", args[3])
			}
		})

		t.Run("complex cmd command routed to cmd /C", func(t *testing.T) {
			raw := `cd /d C:\Users\Eric\Projects\almathurat && findstr /R "^version:" pubspec.yaml`
			cmd, args := parseCommandForExecution(raw)
			if cmd != "cmd" {
				t.Fatalf("expected cmd wrapper, got cmd=%q args=%#v", cmd, args)
			}
			if len(args) != 2 || args[0] != "/C" || args[1] != raw {
				t.Fatalf("unexpected cmd wrapper args: %#v", args)
			}
		})
	}
}
