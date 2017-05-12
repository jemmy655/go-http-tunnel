package integrationtest

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/websocket"
)

// EchoHTTP starts serving HTTP requests on listener l, it accepts connections,
// reads request body and writes is back in response.
func EchoHTTP(l net.Listener) {
	wsServer := &websocket.Server{Handler: func(ws *websocket.Conn) {
		io.Copy(io.MultiWriter(ws, os.Stdout), ws)
	}}

	http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isWebSocketConn(r) {
			wsServer.ServeHTTP(w, r)
			return
		}

		w.WriteHeader(http.StatusOK)
		if r.Body != nil {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}
			w.Write(body)
		}
	}))
}

func isWebSocketConn(r *http.Request) bool {
	return r.Method == "GET" && headerContains(r.Header["Connection"], "upgrade") &&
		headerContains(r.Header["Upgrade"], "websocket")
}

func headerContains(header []string, value string) bool {
	for _, h := range header {
		for _, v := range strings.Split(h, ",") {
			if strings.EqualFold(strings.TrimSpace(v), value) {
				return true
			}
		}
	}

	return false
}

// EchoTCP accepts connections and copies back received bytes.
func EchoTCP(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		go func() {
			io.Copy(conn, conn)
		}()
	}
}

// RandBytes creates a randomy initialised byte slice of length n.
func RandBytes(n int) []byte {
	b := make([]byte, n)
	read, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	if read != n {
		panic("Read did not fill whole slice")
	}
	return b
}

// TLSConfig returns valid http/2 tls configuration that can be used by both
// client and server.
func TLSConfig(cert tls.Certificate) *tls.Config {
	c := &tls.Config{
		Certificates:             []tls.Certificate{cert},
		ClientAuth:               tls.RequestClientCert,
		SessionTicketsDisabled:   true,
		InsecureSkipVerify:       true,
		MinVersion:               tls.VersionTLS12,
		CipherSuites:             []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
		PreferServerCipherSuites: true,
		NextProtos:               []string{"h2"},
	}
	c.BuildNameToCertificate()
	return c
}
