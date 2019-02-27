package httptest

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeAddr string

func (a fakeAddr) Network() string { return string(a) }

func (a fakeAddr) String() string { return string(a) }

type fakeConnection struct {
	readBuff  bytes.Buffer
	writeBuff bytes.Buffer
}

func (c *fakeConnection) Read(b []byte) (n int, err error) {
	return c.readBuff.Read(b)
}

func (c *fakeConnection) Write(b []byte) (n int, err error) {
	return c.writeBuff.Write(b)
}

func (c *fakeConnection) Close() error                       { return nil }
func (c *fakeConnection) LocalAddr() net.Addr                { return fakeAddr("local-address") }
func (c *fakeConnection) RemoteAddr() net.Addr               { return fakeAddr("remote-address") }
func (c *fakeConnection) SetDeadline(t time.Time) error      { return nil }
func (c *fakeConnection) SetReadDeadline(t time.Time) error  { return nil }
func (c *fakeConnection) SetWriteDeadline(t time.Time) error { return nil }

type onetimeListener struct {
	conn net.Conn
}

func (l *onetimeListener) Accept() (net.Conn, error) {
	if l.conn == nil {
		return nil, io.EOF
	}
	conn := l.conn
	l.conn = nil
	return conn, nil
}

func (l *onetimeListener) Close() error   { return nil }
func (l *onetimeListener) Addr() net.Addr { return nil }

func TestConsumeRequest(t *testing.T) {
	c := new(fakeConnection)
	// Write two http requests into connection buffer.
	for i := 0; i < 2; i++ {
		c.readBuff.Write([]byte("GET / HTTP/1.1\r\n" +
			"Host: test\r\n" +
			"Content-Length: 3\r\n" +
			"\r\n" +
			"abc"))
	}

	reqch := make(chan *http.Request)

	l := &onetimeListener{
		conn: c,
	}
	servech := make(chan error)
	go func() {
		servech <- http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqch <- r
		}))
	}()

	for i := 0; i < 2; i++ {
		req := <-reqch
		if req == nil {
			t.Fatal("Got nil Request")
		}
		if req.Method != "GET" {
			t.Errorf("The %dth Request's Method, got %q; expected %q",
				i, req.Method, "GET")
		}
	}

	if err := <-servech; err != io.EOF {
		t.Errorf("Serve returned %v; expected EOF", err)
	}
}

func TestServerTimeout(t *testing.T) {
	tries := []time.Duration{250 * time.Millisecond, 500 * time.Millisecond, 1 * time.Second}
	for _, try := range tries {
		testServerTimeout(t, try)
	}
}

func testServerTimeout(t *testing.T, timeout time.Duration) {
	req := 0
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req++
		fmt.Fprintf(w, "req#%d", req)
	}))
	server.Config.ReadTimeout = timeout
	server.Config.WriteTimeout = timeout
	server.Start()
	defer server.Close()

	client := server.Client()
	url := server.URL

	resp, err := client.Get(url)
	if err != nil {
		t.Errorf("Get first response failed: %v", err)
	}
	get, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Read first response body failed: %v", err)
	}
	expected := "req#1"
	if string(get) != expected {
		t.Errorf("First response, get %s, expect %s", string(get), expected)
	}

	// Slow client should timeout.
	now := time.Now()
	conn, err := net.Dial(server.Listener.Addr().Network(), server.Listener.Addr().String())
	if err != nil {
		t.Errorf("Dial http server failed: %v", err)
	}
	buf := make([]byte, 1)
	n, err := conn.Read(buf)
	if n != 0 || err != io.EOF {
		t.Errorf("Read = %v, %v, wanted %d, %v", n, err, 0, io.EOF)
	}
	conn.Close()
	latency := time.Since(now)
	if latency < timeout {
		t.Errorf("Got EOF after %s, want >= %s", latency, timeout)
	}

	// Request after timeout, still get the right sequence number.
	resp, err = client.Get(url)
	if err != nil {
		t.Errorf("Get second response failed: %v", err)
	}
	get, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Read second response body failed: %v", err)
	}
	expected = "req#2"
	if string(get) != expected {
		t.Errorf("Second response, get %s, expect %s", string(get), expected)
	}
}

func TestOnlyWriteTimeout(t *testing.T) {
	var (
		mu   sync.Mutex
		conn net.Conn
	)
	afterTimeout := make(chan error)
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buff := make([]byte, 512<<10)
		_, err := w.Write(buff)
		if err != nil {
			t.Errorf("Handler: write failed: %v", err)
		}
		mu.Lock()
		defer mu.Unlock()
		if conn == nil {
			t.Errorf("Handler: connection is nil")
		}
		err = conn.SetWriteDeadline(time.Now().Add(-30 * time.Second))
		if err != nil {
			t.Errorf("Handler: set connection deadline failed: %v", err)
		}
		_, err = w.Write(buff)
		if err == nil {
			t.Errorf("Handler: write after deadline setting failed: %v", err)
		}
		afterTimeout <- err
	}))
	server.Listener = &trackLastConnectionListener{server.Listener, &mu, &conn}
	server.Start()
	defer server.Close()

	client := server.Client()

	errch := make(chan error)
	go func() {
		defer close(errch)
		resp, err := client.Get(server.URL)
		if err != nil {
			errch <- err
		}
		_, err = io.Copy(ioutil.Discard, resp.Body)
		if err != nil {
			errch <- err
		}
		resp.Body.Close()
	}()

	select {
	case err := <-errch:
		if err == nil {
			t.Errorf("Getting response should failed")
		}
	case <-time.After(5 * time.Second):
		t.Errorf("Timeout waiting for getting response")
	}
	if err := <-afterTimeout; err == nil {
		t.Errorf("Should get error after setting timeout")
	}
}

type trackLastConnectionListener struct {
	net.Listener

	mu   *sync.Mutex
	conn *net.Conn
}

func (t *trackLastConnectionListener) Accept() (net.Conn, error) {
	conn, err := t.Listener.Accept()
	if err != nil {
		return nil, err
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	*t.conn = conn

	return conn, nil
}

func testTCPConnectionClose(t *testing.T, req string, handler http.Handler) {
	server := httptest.NewServer(handler)
	defer server.Close()

	conn, err := net.Dial(server.Listener.Addr().Network(), server.Listener.Addr().String())
	if err != nil {
		t.Errorf("Dial http server failed: %v", err)
	}
	defer conn.Close()

	n, err := conn.Write([]byte(req))
	if err != nil || n != len(req) {
		t.Errorf("Write request failed: %v", err)
	}

	r := bufio.NewReader(conn)
	resp, err := http.ReadResponse(r, &http.Request{Method: "GET"})
	if err != nil {
		t.Errorf("Read response failed: %v", err)
	}

	donech := make(chan struct{})
	go func() {
		select {
		case <-time.After(5 * time.Second):
			t.Errorf("Timeout when waiting for reading finished")
		case <-donech:
		}
	}()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Read from response body failed: %v", err)
	}

	donech <- struct{}{}

	if !resp.Close {
		t.Errorf("Connection should be closed")
	}
}

func TestHTTP10Close(t *testing.T) {
	testTCPConnectionClose(t, "GET / HTTP/1.0\r\n\r\n", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do nothing.
	}))
}

func TestClientCanClose(t *testing.T) {
	testTCPConnectionClose(t, "GET / HTTP/1.1\r\nConnection: close\r\n\r\n", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do nothing.
	}))
}

func TestHandlerCanSetConnectionClose11(t *testing.T) {
	testTCPConnectionClose(t, "GET / HTTP/1.1\r\n\r\n", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "close")
	}))
}

func TestHandlerCanSetConnectionClose10(t *testing.T) {
	testTCPConnectionClose(t, "GET / HTTP/1.0\r\nConnection: keep-alive\r\n\r\n", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "close")
	}))
}

func TestContentLengthZero(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	versions := []string{"HTTP/1.0", "HTTP/1.1"}
	for _, v := range versions {
		conn, err := net.Dial(server.Listener.Addr().Network(), server.Listener.Addr().String())
		if err != nil {
			t.Errorf("Dial failed: %v", err)
		}

		fmt.Fprintf(conn, "GET / %v\r\nConnection: keep-alive\r\nHost: foo\r\n\r\n", v)

		r := bufio.NewReader(conn)
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Errorf("Construct new request failed: %v", err)
		}
		resp, err := http.ReadResponse(r, req)
		if err != nil {
			t.Errorf("Read response failed: %v", err)
		}

		if len(resp.TransferEncoding) > 0 {
			t.Errorf("The TransferEncoding should be 0")
		}

		if resp.ContentLength != 0 {
			t.Errorf("The content length should be 0 otherthan %d", resp.ContentLength)
		}

		conn.Close()
	}
}

func TestCloseNotifier(t *testing.T) {
	godie := make(chan struct{})
	closec := make(chan bool)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		godie <- struct{}{}
		n, ok := w.(http.CloseNotifier)
		if !ok {
			t.Errorf("Asset ResponseWriter to CloseNotifier failed")
		}
		c := n.CloseNotify()
		<-c
		close(closec)
	}))
	defer server.Close()

	conn, err := net.Dial(server.Listener.Addr().Network(), server.Listener.Addr().String())
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: foo\r\n\r\n")
	go func() {
		<-godie
		conn.Close()
	}()

	select {
	case <-closec:
		break
	case <-time.After(5 * time.Second):
		t.Errorf("Timeout waiting for close notifier channel")
	}
}

func TestServerHijackGetBackgroundBytes(t *testing.T) {
	done := make(chan struct{})
	size := 8 << 10
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(done)

		// HTTP server wait for the extra data after GET request.
		select {
		case <-w.(http.CloseNotifier).CloseNotify():
		case <-time.After(5 * time.Second):
			t.Errorf("Timeout waiting for connection being closed")
		}

		conn, b, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Errorf("Hijack failed: %v", err)
		}

		buf, err := ioutil.ReadAll(b)
		if err != nil {
			t.Errorf("Read hijack buffer failed: %v", err)
		}

		if len(buf) != size {
			t.Errorf("Hijack buffer size, get %v, expect %v", len(buf), size)
		}

		flag := true
		for _, b := range buf {
			if b != 'x' {
				flag = false
			}
		}
		if flag == false {
			t.Errorf("Include wrong bytes in hijack buffer")
		}

		conn.Close()
	}))
	defer server.Close()

	conn, err := net.Dial(server.Listener.Addr().Network(), server.Listener.Addr().String())
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	defer conn.Close()
	fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: foo\r\n\r\n%v", strings.Repeat("x", size))

	err = conn.(*net.TCPConn).CloseWrite()
	if err != nil {
		t.Errorf("CloseWrite failed: %v", err)
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Errorf("Timeout waiting for request handling finished")
	}
}

func TestWithClientTrace(t *testing.T) {
	var buf bytes.Buffer
	f := func(b byte) func(string, string) {
		return func(network, addr string) {
			buf.Write([]byte{b})
		}
	}

	ctx := context.Background()

	old := &httptrace.ClientTrace{
		ConnectStart: f('O'),
	}
	ctx = httptrace.WithClientTrace(ctx, old)
	trace := &httptrace.ClientTrace{
		ConnectStart: f('N'),
	}
	ctx = httptrace.WithClientTrace(ctx, trace)

	trace = httptrace.ContextClientTrace(ctx)
	trace.ConnectStart("network", "addr")

	if get, want := buf.String(), "NO"; get != want {
		t.Errorf("Expect \"NO\", get %v", get)
	}
}
