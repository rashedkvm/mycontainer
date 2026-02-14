package main

import (
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
)

const (
	DefaultPort = "8080"
)

type Config struct {
	Port        string
	ContainerID string
}

func loadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	containerID := os.Getenv("CONTAINER_ID")
	if containerID == "" {
		containerID = "N/A"
	}

	return Config{
		Port:        port,
		ContainerID: containerID,
	}
}

type IPResponse struct {
	ServerIP string `json:"serverIP"`
	ClientIP string `json:"clientIP"`
}

type ConfigResponse struct {
	ContainerID string `json:"containerID"`
}

func main() {
	// Configure structured logger with JSON output
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load configuration
	config := loadConfig()

	// Serve static files with logging
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", logRequest(fs))

	// API endpoint for IP addresses
	http.HandleFunc("/api/ip", handleIP)

	// API endpoint for config
	http.HandleFunc("/api/config", handleConfigWithConfig(config))

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	slog.Info("Server starting", "port", config.Port)
	if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
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
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode JSON response", "error", err)
	}
}

func getServerIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "Unknown"
	}

	var ipv6Addr string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String() // Prefer IPv4
			}
			if ipv6Addr == "" && ipnet.IP.To16() != nil {
				ipv6Addr = ipnet.IP.String() // Store first IPv6 as fallback
			}
		}
	}
	if ipv6Addr != "" {
		return ipv6Addr
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

func handleConfigWithConfig(config Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		slog.Info("Request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)

		response := ConfigResponse{
			ContainerID: config.ContainerID,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("Failed to encode JSON response", "error", err)
		}
	}
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	slog.Info("Request",
		"method", r.Method,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
	)

	containerID := os.Getenv("CONTAINER_ID")
	if containerID == "" {
		containerID = "N/A"
	}
	response := ConfigResponse{
		ContainerID: containerID,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode JSON response", "error", err)
	}
}
