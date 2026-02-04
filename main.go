package main

import (
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
)

type IPResponse struct {
	ServerIP string `json:"serverIP"`
	ClientIP string `json:"clientIP"`
}

func main() {
	// Configure structured logger with JSON output
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Serve static files with logging
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", logRequest(fs))

	// API endpoint for IP addresses
	http.HandleFunc("/api/ip", handleIP)

	slog.Info("Server starting", "port", 8080)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

// logRequest wraps an http.Handler to log incoming requests
func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)
		handler.ServeHTTP(w, r)
	})
}

func handleIP(w http.ResponseWriter, r *http.Request) {
	slog.Info("Request",
		"method", r.Method,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
	)

	response := IPResponse{
		ServerIP: getServerIP(),
		ClientIP: getClientIP(r),
	}

	slog.Info("Serving IPs",
		"server_ip", response.ServerIP,
		"client_ip", response.ClientIP,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getServerIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "Unknown"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "Unknown"
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
