package main

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name              string
		remoteAddr        string
		xForwardedFor     string
		xRealIP           string
		expectedIP        string
	}{
		{
			name:       "direct connection",
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:          "X-Forwarded-For single IP",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1",
			expectedIP:    "203.0.113.1",
		},
		{
			name:          "X-Forwarded-For multiple IPs",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1, 198.51.100.1, 10.0.0.1",
			expectedIP:    "203.0.113.1",
		},
		{
			name:          "X-Forwarded-For with spaces",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "  203.0.113.1  ",
			expectedIP:    "203.0.113.1",
		},
		{
			name:       "X-Real-IP",
			remoteAddr: "10.0.0.1:12345",
			xRealIP:    "203.0.113.1",
			expectedIP: "203.0.113.1",
		},
		{
			name:          "X-Forwarded-For takes precedence over X-Real-IP",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1",
			xRealIP:       "198.51.100.1",
			expectedIP:    "203.0.113.1",
		},
		{
			name:       "IPv6 address",
			remoteAddr: "[2001:db8::1]:12345",
			expectedIP: "2001:db8::1",
		},
		{
			name:       "RemoteAddr without port",
			remoteAddr: "192.168.1.1",
			expectedIP: "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/ip", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			got := getClientIP(req)
			if got != tt.expectedIP {
				t.Errorf("getClientIP() = %v, want %v", got, tt.expectedIP)
			}
		})
	}
}

func TestGetServerIP(t *testing.T) {
	// This test just verifies the function returns something valid
	ip := getServerIP()
	
	if ip == "" {
		t.Error("getServerIP() returned empty string")
	}
	
	if ip != "Unknown" {
		// If not "Unknown", should be a valid IP
		if net.ParseIP(ip) == nil {
			t.Errorf("getServerIP() returned invalid IP: %v", ip)
		}
	}
}

func TestHandleIP(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkBody      bool
	}{
		{
			name:           "GET request succeeds",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			checkBody:      true,
		},
		{
			name:           "POST request fails",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			checkBody:      false,
		},
		{
			name:           "PUT request fails",
			method:         http.MethodPut,
			expectedStatus: http.StatusMethodNotAllowed,
			checkBody:      false,
		},
		{
			name:           "DELETE request fails",
			method:         http.MethodDelete,
			expectedStatus: http.StatusMethodNotAllowed,
			checkBody:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/ip", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			w := httptest.NewRecorder()

			handleIP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("handleIP() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.checkBody {
				var response IPResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("handleIP() failed to decode response: %v", err)
				}
				if response.ClientIP == "" {
					t.Error("handleIP() ClientIP is empty")
				}
				if response.ServerIP == "" {
					t.Error("handleIP() ServerIP is empty")
				}
			}
		})
	}
}

func TestHandleConfig(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		containerID    string
		expectedStatus int
		checkBody      bool
	}{
		{
			name:           "GET request succeeds",
			method:         http.MethodGet,
			containerID:    "test-container-123",
			expectedStatus: http.StatusOK,
			checkBody:      true,
		},
		{
			name:           "GET request with empty container ID",
			method:         http.MethodGet,
			containerID:    "",
			expectedStatus: http.StatusOK,
			checkBody:      true,
		},
		{
			name:           "POST request fails",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			checkBody:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if provided
			if tt.containerID != "" {
				t.Setenv("CONTAINER_ID", tt.containerID)
			}

			req := httptest.NewRequest(tt.method, "/api/config", nil)
			w := httptest.NewRecorder()

			handleConfig(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("handleConfig() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.checkBody {
				var response ConfigResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("handleConfig() failed to decode response: %v", err)
				}
				if tt.containerID != "" {
					if response.ContainerID != tt.containerID {
						t.Errorf("handleConfig() ContainerID = %v, want %v", response.ContainerID, tt.containerID)
					}
				} else {
					if response.ContainerID != "N/A" {
						t.Errorf("handleConfig() ContainerID = %v, want N/A", response.ContainerID)
					}
				}
			}
		})
	}
}

func TestLogRequestMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrapped := logRequest(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("logRequest middleware changed status code: got %v, want %v", w.Code, http.StatusOK)
	}

	if w.Body.String() != "OK" {
		t.Errorf("logRequest middleware changed response body: got %v, want OK", w.Body.String())
	}
}
