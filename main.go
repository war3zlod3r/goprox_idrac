package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
)

// Helper function to dynamically get the preferred local IP
func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1" // Fallback to localhost if network is down
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func main() {
	localIP := getLocalIP()
	log.Printf("Starting HTTP-to-HTTPS proxy on %s:8888", localIP)

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			log.Printf("Intercepting client request: %s %s%s", req.Method, req.Host, req.URL.Path)

			req.URL.Scheme = "https"

			if req.URL.Host == "" {
				req.URL.Host = req.Host
			}

			if req.URL.Host == "downloads.dell.com:80" {
				req.URL.Host = "downloads.dell.com"
			}

			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		},
	}

	server := &http.Server{
		Addr: ":8888",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				log.Printf("Rejecting CONNECT tunnel request from %s. Please configure iDRAC to use HTTP for updates.", r.RemoteAddr)
				http.Error(w, "HTTPS CONNECT tunneling not supported. Use HTTP endpoint.", http.StatusMethodNotAllowed)
				return
			}
			
			proxy.ServeHTTP(w, r)
		}),
	}

	log.Fatal(server.ListenAndServe())
}
