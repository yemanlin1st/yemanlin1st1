package main

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"pefy.gg/omega-traffic-fabric/internal/balancer"
)

func TestSplitBackends(t *testing.T) {
	got := splitBackends(" a:1, ,b:2 ,c:3")
	want := []string{"a:1", "b:2", "c:3"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestDialSelectedBackendFailsOver(t *testing.T) {
	dead, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	deadAddr := dead.Addr().String()
	_ = dead.Close()

	live, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer live.Close()
	go func() {
		c, err := live.Accept()
		if err == nil {
			_ = c.Close()
		}
	}()

	pool, err := balancer.NewPool([]string{deadAddr, live.Addr().String()}, balancer.RoundRobin)
	if err != nil {
		t.Fatal(err)
	}
	first := pool.Pick("client")
	conn, selected := dialSelectedBackend(pool, first, "client", 250*time.Millisecond)
	if conn == nil || selected == nil {
		t.Fatal("expected fallback connection")
	}
	defer conn.Close()
	if selected.Addr != live.Addr().String() {
		t.Fatalf("selected %s want %s", selected.Addr, live.Addr().String())
	}
}

func TestHandleConnPreservesTCPHalfClose(t *testing.T) {
	backendListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer backendListener.Close()

	backendDone := make(chan error, 1)
	go func() {
		c, err := backendListener.Accept()
		if err != nil {
			backendDone <- err
			return
		}
		defer c.Close()
		body, err := io.ReadAll(c)
		if err != nil {
			backendDone <- err
			return
		}
		if string(body) != "request" {
			backendDone <- &testError{"unexpected backend request: " + string(body)}
			return
		}
		// Make truncation deterministic against an implementation that closes
		// both sides as soon as the client->backend copy direction reaches EOF.
		time.Sleep(25 * time.Millisecond)
		_, err = c.Write([]byte("response"))
		backendDone <- err
	}()

	front, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer front.Close()

	pool, err := balancer.NewPool([]string{backendListener.Addr().String()}, balancer.PowerOfTwo)
	if err != nil {
		t.Fatal(err)
	}
	m := &metrics{}
	proxyDone := make(chan struct{})
	go func() {
		c, err := front.Accept()
		if err == nil {
			handleConn(context.Background(), c, pool, m, time.Second)
		}
		close(proxyDone)
	}()

	clientRaw, err := net.Dial("tcp", front.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	client := clientRaw.(*net.TCPConn)
	if _, err := client.Write([]byte("request")); err != nil {
		t.Fatal(err)
	}
	if err := client.CloseWrite(); err != nil {
		t.Fatal(err)
	}
	response, err := io.ReadAll(client)
	_ = client.Close()
	if err != nil {
		t.Fatal(err)
	}
	if string(response) != "response" {
		t.Fatalf("got response %q", string(response))
	}
	if err := <-backendDone; err != nil {
		t.Fatal(err)
	}
	select {
	case <-proxyDone:
	case <-time.After(time.Second):
		t.Fatal("proxy did not terminate")
	}
}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
