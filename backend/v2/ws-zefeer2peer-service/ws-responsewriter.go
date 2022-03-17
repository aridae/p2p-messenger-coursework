package wszefeer2peer

import (
	"bufio"
	"log"
	"net"
	"net/http"
)

//WSZefeerWriter - реализация интерфейса ResponseWriter
type WSZefeerWriter struct {
	conn net.Conn
}

func (w WSZefeerWriter) Write(b []byte) (int, error) {
	return w.conn.Write(b)
}

func (w WSZefeerWriter) Header() http.Header {
	return http.Header{}
}

func (w WSZefeerWriter) WriteHeader(statusCode int) {
	_, err := w.conn.Write([]byte("HTTP/1.1 200 OK"))
	if err != nil {
		log.Printf("WriteHeaderError: %v\n", err)
	}
}

func (w WSZefeerWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	reader := bufio.NewReader(w.conn)
	writer := bufio.NewWriter(w.conn)

	readWriter := bufio.NewReadWriter(reader, writer)
	return w.conn, readWriter, nil
}

func NewWSZefeerWriter(conn net.Conn) http.ResponseWriter {
	return &WSZefeerWriter{conn}
}
