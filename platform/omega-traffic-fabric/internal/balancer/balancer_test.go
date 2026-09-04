package balancer

import (
	"fmt"
	"testing"
)

func TestRoundRobin(t *testing.T) {
	p, err := NewPool([]string{"a:1", "b:2", "c:3"}, RoundRobin)
	if err != nil {
		t.Fatal(err)
	}
	got := []string{p.Pick("").Addr, p.Pick("").Addr, p.Pick("").Addr, p.Pick("").Addr}
	want := []string{"a:1", "b:2", "c:3", "a:1"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pick %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestLeastConnections(t *testing.T) {
	p, _ := NewPool([]string{"a:1", "b:2"}, LeastConn)
	p.backends[0].Acquire()
	p.backends[0].Acquire()
	defer p.backends[0].Release()
	defer p.backends[0].Release()
	if got := p.Pick("").Addr; got != "b:2" {
		t.Fatalf("got %q", got)
	}
}

func TestUnhealthyExcluded(t *testing.T) {
	p, _ := NewPool([]string{"a:1", "b:2"}, PowerOfTwo)
	p.backends[0].SetHealthy(false)
	for i := 0; i < 10; i++ {
		if got := p.Pick("x").Addr; got != "b:2" {
			t.Fatalf("got %q", got)
		}
	}
}

func TestMaglevDeterministic(t *testing.T) {
	p, _ := NewPool([]string{"a:1", "b:2", "c:3"}, Maglev)
	first := p.Pick("client-42").Addr
	for i := 0; i < 100; i++ {
		if got := p.Pick("client-42").Addr; got != first {
			t.Fatalf("got %q want %q", got, first)
		}
	}
}

func BenchmarkP2C(b *testing.B) {
	p, _ := NewPool([]string{"a:1", "b:2", "c:3", "d:4"}, PowerOfTwo)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Pick("bench")
	}
}

func BenchmarkMaglev(b *testing.B) {
	p, _ := NewPool([]string{"a:1", "b:2", "c:3", "d:4"}, Maglev)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Pick("bench")
	}
}

func BenchmarkP2CLargePool(b *testing.B) {
	addrs := make([]string, 4096)
	for i := range addrs {
		addrs[i] = fmt.Sprintf("10.0.%d.%d:8080", (i/254)%254, (i%254)+1)
	}
	p, _ := NewPool(addrs, PowerOfTwo)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Pick("bench")
	}
}
