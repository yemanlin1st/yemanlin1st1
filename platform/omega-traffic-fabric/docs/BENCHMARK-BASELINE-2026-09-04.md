# ΩTF Technology Benchmark Baseline — 2026-09-04

This document is a public-safe architecture benchmark, not an instruction to copy third-party implementation code into the sovereign production core.

## Current reference signals

| Technology | Relevant capability | ΩTF decision |
|---|---|---|
| Cilium | eBPF distributed load balancing, socket-level east-west routing, XDP north-south acceleration, DSR, Maglev | Use as the primary architectural reference for qualified Linux/eBPF fast-path adapters. |
| Envoy | Weighted least-request/P2C, Ring Hash, Maglev, rich L7 routing and locality policies | Retain P2C and consistent-hash semantics; treat locality/hash combinations as explicit compatibility cases rather than blindly composing policies. |
| Kubernetes Gateway API | Stable/Standard portable GatewayClass, Gateway, HTTPRoute and ReferenceGrant contract; current documented Standard bundle v1.6.1 | Use the Standard Channel as the default Kubernetes northbound API. Experimental resources remain gated. |
| Pingora | Rust framework for programmable network services; current release line 0.8.x; mTLS and security hardening are evolving | Candidate Rust L7 adapter/framework only after dependency/advisory qualification. Never auto-promote a new upstream release. |
| HAProxy | Mature TCP/HTTP/QUIC balancing; 3.4 adds dynamic backends, memory/performance improvements and OpenTelemetry | Maintain a qualified adapter/reference path and absorb proven operational patterns without coupling the sovereign core. |
| NGINX | Round-robin, least-connections, hashing and passive health checks in OSS; active health checking in Plus | Retain as compatibility/edge adapter and baseline operational comparison. |

## Security evidence affecting update policy

Recent ecosystem evidence demonstrates why ΩEVOLVE must separate discovery from production promotion:

- Pingora 0.8.0 patched a critical HTTP request-smuggling issue affecting <=0.7.0; 0.8.1 later added additional HTTP/2 memory-exhaustion hardening and dependency advisory updates.
- An August 2026 Pingora dependency issue records a RustSec unsoundness advisory in an `lru` version used by the workspace, illustrating that a project can be current while still carrying dependency findings requiring assessment.
- HAProxy published fixes in February 2026 for two high-severity QUIC denial-of-service vulnerabilities.

Therefore all ΩTF production upgrades require provenance, dependency/advisory review, reproducible tests, benchmark regression gates, canary observation and rollback readiness. TUF-style update trust and SLSA-compatible build provenance are part of the target release-control model.

## Local bootstrap verification

Test host: Linux amd64, AMD EPYC 9V74 80-Core Processor.

Verification commands:

```text
go test ./...
go vet ./...
go test -race ./...
go test -bench='P2C|Maglev' -benchmem ./internal/balancer
```

Results after hot-path redesign, TCP half-close/failover fixes, overload shedding and observability hardening:

| Test | Result |
|---|---:|
| Unit + integration tests | PASS |
| `go vet` | PASS |
| Go race detector | PASS |
| P2C selector — 4 backends | ~16.03 ns/op, 0 B/op, 0 allocs/op |
| P2C selector — 4,096 backends | ~12.69 ns/op, 0 B/op, 0 allocs/op |
| Maglev lookup | ~8.75 ns/op, 0 B/op, 0 allocs/op |

The P2C design uses direct randomized probing: O(1) typical selection for healthy pools with bounded O(N) probing only when searching across unhealthy nodes. Least-connections remains intentionally O(N). Maglev uses a prebuilt lookup table.

These numbers measure only in-process backend selection on this host. They are not packets-per-second, requests-per-second, latency SLO, or production throughput claims.

## Qualification gaps before production-core release

The public bootstrap remains intentionally below production ΩTF scope. The sovereign private implementation still requires controlled milestones for TLS/mTLS, HTTP/2 and HTTP/3/QUIC L7 handling, signed dynamic configuration, eBPF/XDP, DSR, BGP/Anycast, Kubernetes reconciliation, fuzzing, chaos/failure tests, multi-zone/region evacuation, SLO-driven canarying, supply-chain evidence and reproducible end-to-end performance testing.
