package delivery

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"proxy/internal/network"
	"proxy/internal/proxy/certs"
	"sync"
)

type MutexBuffer struct {
	buf []byte
	mu  sync.Mutex
}

func (proxy *ProxyHandler) handleConnect(writer http.ResponseWriter, req *network.HTTPRequest) error {
	config, err := certs.GetTLSConfig(req, proxy.ca)
	if err != nil {
		log.Println("error getting tls config", err)

		return err
	}

	netConn, err := handshake(writer, config.Cfg)
	if err != nil {
		log.Println("error getting handshake connection", err)

		return err
	}

	defer netConn.Close()

	if config.Conn == nil {
		config.Conn, err = tls.Dial("tcp", fmt.Sprintf("%s:%s", req.Host, req.Port), config.Cfg)
		if err != nil {
			log.Println("error while creating tls connection", err)

			return err
		}
	}
	defer config.Conn.Close()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	reqBuf := &MutexBuffer{
		buf: make([]byte, 0, 100*1024),
		mu:  sync.Mutex{},
	}
	respBuf := &MutexBuffer{
		buf: make([]byte, 0, 100*1024),
		mu:  sync.Mutex{},
	}

	go transfer(config.Conn, netConn, wg, respBuf)
	go transfer(netConn, config.Conn, wg, reqBuf)

	wg.Wait()

	reqBuf.mu.Lock()
	reqBuf.mu.Unlock()

	tlsReq, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(reqBuf.buf)))
	if err != nil {
		log.Println("error reading request", err)

		return err
	}

	respBuf.mu.Lock()
	tlsResp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(respBuf.buf)), tlsReq)
	respBuf.mu.Unlock()

	if err != nil {
		log.Println("error reading request", err)

		return err
	}

	parsedReq := network.NewHTTPRequest(tlsReq)
	parsedReq.Scheme = "HTTPS"
	parsedReq.Port = "443"

	id, err := proxy.storage.SaveRequest(parsedReq)
	if err != nil {
		log.Println("Something went wrong while saving request", err)

		return err
	}

	parsedResp := network.NewHTTPResponse(tlsResp)
	err = proxy.storage.SaveResponse(parsedResp, id)
	if err != nil {
		log.Println("Something went wrong while saving response", err)

		return err
	}

	return nil
}

func handshake(writer http.ResponseWriter, config *tls.Config) (net.Conn, error) {
	raw, _, err := writer.(http.Hijacker).Hijack()
	if err != nil {
		http.Error(writer, "no upstream", 503)

		return nil, err
	}

	if _, err = raw.Write([]byte("HTTP/1.1 200 Connection Established\n\n")); err != nil {
		raw.Close()

		return nil, err
	}

	conn := tls.Server(raw, config)

	err = conn.Handshake()
	if err != nil {
		conn.Close()
		raw.Close()

		return nil, err
	}

	return conn, nil
}

func transfer(reader io.Reader, writer io.Writer, wg *sync.WaitGroup, transferred *MutexBuffer) {
	defer wg.Done()

	buf := make([]byte, 10*1024)

	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			log.Println("error reading from connection:", err)

			return
		}

		err = nil

		if n > 0 {
			transferred.mu.Lock()
			transferred.buf = append(transferred.buf, buf[:n]...)

			transferred.mu.Unlock()

			_, err = writer.Write(buf[:n])
			if err != nil {
				log.Println("error writing to connection:", err)

				return
			}
		} else {
			return
		}
	}
}
