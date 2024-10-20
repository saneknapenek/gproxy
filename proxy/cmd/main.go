package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/saneknapenek/clog"
)

func main() {

	log := clog.InitLogger(clog.EnvDev, nil)

	cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Error(err.Error())
		return
	}

	server := &http.Server{
		Addr: ":8888", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				handleTunneling(*log, cer, w, r)
			}
			// } else {
			// 	handleHTTP(w, r)
			// }
		}),
		// Disable HTTP/2.
		// TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error(err.Error())
		return
	}
}

func handleTunneling(log slog.Logger, cert tls.Certificate, w http.ResponseWriter, r *http.Request) {
	// dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	// if err != nil {
	//     http.Error(w, err.Error(), http.StatusServiceUnavailable)
	//     return
	// }

	w.WriteHeader(http.StatusOK)

	strHeaders, err := headerToString(r.Header)
	if err != nil {
		log.Error("error parsing headers:", err.Error())
	} else {
		log.Debug("", slog.String("Headers", strHeaders))
	}

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("error reading request body:", err.Error())
		return
	}
	defer r.Body.Close()

	log.Debug("response body:", string(requestBody))

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		log.Debug("Hijacking not supported", slog.String("host", r.Host))
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Debug("connection capture error", slog.String("error", err.Error()))
		return
	}

	defer clientConn.Close()

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}

	tlsConn := tls.Server(clientConn, tlsConfig)

	err = tlsConn.Handshake()
	if err != nil {
		http.Error(w, "Failed to complete TLS handshake: "+err.Error(), http.StatusInternalServerError)
		log.Error("Failed TLS handshake", slog.String("host", r.Host), slog.String("error", err.Error()))
		return
	}

	go transfer(clientConn, tlsConn)
	go transfer(tlsConn, clientConn)
}

func transfer(dst io.WriteCloser, src io.ReadCloser) {
	defer dst.Close()
	defer src.Close()
	io.Copy(dst, src) // Копируем данные
}

func headerToString(headers http.Header) (string, error) {
	var sb strings.Builder

	for key, values := range headers {
		for _, value := range values {
			_, err := sb.WriteString(fmt.Sprintf("%s: %s\n", key, value))
			if err != nil {
				return "", err
			}
		}
	}

	return sb.String(), nil
}

// func headerToSlogSlice(headers http.Header) ([]slog.Attr, error) {
// 	hCount := len(headers)

//     result := make([]slog.Attr, 0, hCount)

//     for key, values := range headers {
//         for _, value := range values {
//             // Добавляем заголовок в срез в формате slog.String
//             result = append(result, slog.String(key, value))
//         }
//     }

//     return result, nil
// }
