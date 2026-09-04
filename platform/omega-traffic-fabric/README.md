# PEFY ΩTRAFFIC FABRIC™ / ΩBALANCER™

Public-safe bootstrap implementation of the sovereign PEFY-GG traffic-management layer.

Owner Code: PGG-SOC-001  
IP Controller: PGG-IP-CTRL  
Asset Registry: PGG-ASSET-REG  
Technical Production: by PEFY-TECH  
Powered by: PEFY-GG (GROUP PEFY-CONSULTING GLOBAL)

## Current executable scope

This repository module provides a dependency-free Go bootstrap data plane for L4 TCP balancing with:

- round robin;
- least connections;
- power-of-two choices (P2C);
- Maglev-style consistent hashing;
- active TCP health checks;
- passive failure marking;
- graceful process shutdown/draining;
- Prometheus-compatible text metrics;
- health endpoint;
- deterministic configuration through flags/environment variables.

It is deliberately a **bootstrap implementation**, not a claim of hyperscale parity with mature Envoy/Cilium/HAProxy/Pingora deployments. The production ΩTF architecture is pluggable: optimized Rust/Pingora and eBPF/XDP data planes can be introduced behind the same policy and conformance model after qualification.

## Run

```bash
go test ./...
go run ./cmd/omegabalancer \
  -listen :8080 \
  -backends 127.0.0.1:8081,127.0.0.1:8082 \
  -algorithm p2c
```

Observability defaults to `:9090`:

- `/healthz`
- `/metrics`

## Safety invariants

1. No self-discovered dependency or Internet code is promoted directly to production.
2. All changes must pass source/license review, SBOM/provenance controls, tests, performance gates, canary deployment, and rollback readiness.
3. AI may recommend traffic policy but is never required in the synchronous packet/connection forwarding path.
4. Retry semantics for L7 adapters must respect idempotency and explicit retry budgets.
5. Experimental upstream capabilities remain disabled until qualified.
