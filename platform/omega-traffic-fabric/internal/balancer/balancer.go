package balancer

import (
	"errors"
	"hash/fnv"
	"strings"
	"sync/atomic"
)

type Algorithm string

const (
	RoundRobin             Algorithm = "round-robin"
	LeastConn              Algorithm = "least-connections"
	PowerOfTwo             Algorithm = "p2c"
	Maglev                 Algorithm = "maglev"
	defaultMaglevTableSize           = 65537
)

type Backend struct {
	Addr    string
	healthy atomic.Bool
	active  atomic.Int64
}

func NewBackend(addr string) *Backend {
	b := &Backend{Addr: strings.TrimSpace(addr)}
	b.healthy.Store(true)
	return b
}

func (b *Backend) Healthy() bool     { return b.healthy.Load() }
func (b *Backend) SetHealthy(v bool) { b.healthy.Store(v) }
func (b *Backend) Active() int64     { return b.active.Load() }
func (b *Backend) Acquire()          { b.active.Add(1) }
func (b *Backend) Release()          { b.active.Add(-1) }

type Pool struct {
	backends []*Backend
	algo     Algorithm
	rr       atomic.Uint64
	entropy  atomic.Uint64
	maglev   []int
}

func NewPool(addrs []string, algo Algorithm) (*Pool, error) {
	if len(addrs) == 0 {
		return nil, errors.New("at least one backend is required")
	}
	p := &Pool{algo: algo}
	for _, addr := range addrs {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}
		p.backends = append(p.backends, NewBackend(addr))
	}
	if len(p.backends) == 0 {
		return nil, errors.New("no valid backend addresses provided")
	}
	switch algo {
	case RoundRobin, LeastConn, PowerOfTwo:
	case Maglev:
		p.maglev = buildMaglev(p.backends, defaultMaglevTableSize)
	default:
		return nil, errors.New("unsupported algorithm: " + string(algo))
	}
	return p, nil
}

func (p *Pool) Backends() []*Backend { return p.backends }

func (p *Pool) Pick(key string) *Backend {
	n := len(p.backends)
	if n == 0 {
		return nil
	}

	switch p.algo {
	case RoundRobin:
		start := int((p.rr.Add(1) - 1) % uint64(n))
		return p.nextHealthy(start)
	case LeastConn:
		var best *Backend
		var bestLoad int64
		for _, b := range p.backends {
			if !b.Healthy() {
				continue
			}
			load := b.Active()
			if best == nil || load < bestLoad {
				best, bestLoad = b, load
			}
		}
		return best
	case PowerOfTwo:
		x := p.entropy.Add(0x9e3779b97f4a7c15)
		aStart := int(mix64(x) % uint64(n))
		bStart := int(mix64(x^0xbf58476d1ce4e5b9) % uint64(n))
		a := p.nextHealthy(aStart)
		if a == nil {
			return nil
		}
		if n == 1 {
			return a
		}
		if bStart == aStart {
			bStart = (bStart + 1) % n
		}
		b := p.nextHealthy(bStart)
		if b == nil || b == a || a.Active() <= b.Active() {
			return a
		}
		return b
	case Maglev:
		if len(p.maglev) == 0 {
			return nil
		}
		start := int(hash64(key) % uint64(len(p.maglev)))
		for i := 0; i < len(p.maglev); i++ {
			idx := p.maglev[(start+i)%len(p.maglev)]
			if idx >= 0 && p.backends[idx].Healthy() {
				return p.backends[idx]
			}
		}
	}
	return nil
}

func (p *Pool) nextHealthy(start int) *Backend {
	n := len(p.backends)
	if n == 0 {
		return nil
	}
	if start < 0 {
		start = 0
	}
	start %= n
	for i := 0; i < n; i++ {
		b := p.backends[(start+i)%n]
		if b.Healthy() {
			return b
		}
	}
	return nil
}

func hash64(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}

func mix64(z uint64) uint64 {
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

func buildMaglev(backends []*Backend, m int) []int {
	if m < 2 {
		m = defaultMaglevTableSize
	}
	n := len(backends)
	offset := make([]int, n)
	skip := make([]int, n)
	next := make([]int, n)
	entry := make([]int, m)
	for i := range entry {
		entry[i] = -1
	}
	for i, b := range backends {
		h1 := hash64("offset:" + b.Addr)
		h2 := hash64("skip:" + b.Addr)
		offset[i] = int(h1 % uint64(m))
		skip[i] = int(h2%uint64(m-1)) + 1
		if gcd(skip[i], m) != 1 {
			for gcd(skip[i], m) != 1 {
				skip[i]++
				if skip[i] >= m {
					skip[i] = 1
				}
			}
		}
	}
	filled := 0
	for filled < m {
		for i := 0; i < n && filled < m; i++ {
			c := (offset[i] + next[i]*skip[i]) % m
			for entry[c] >= 0 {
				next[i]++
				c = (offset[i] + next[i]*skip[i]) % m
			}
			entry[c] = i
			next[i]++
			filled++
		}
	}
	return entry
}

func gcd(a, b int) int {
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
