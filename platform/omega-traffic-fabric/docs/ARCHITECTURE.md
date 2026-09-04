# ΩTF Architecture Baseline

## Layers

1. **ΩGLOBAL DIRECTOR** — DNS, Geo, latency, Anycast/BGP and region steering.
2. **ΩFASTPATH** — eBPF/XDP L3/L4 acceleration where kernel/platform support and qualification exist.
3. **ΩBALANCER** — portable L4/L7 balancing engines selected by deployment profile.
4. **ΩGATEWAY CONTROLLER** — Kubernetes Gateway API Standard Channel integration.
5. **ΩSERVICE DIRECTOR** — east-west service discovery, locality and resilience policy.
6. **ΩOBSERVE** — OpenTelemetry/Prometheus-compatible telemetry.
7. **ΩEVOLVE** — research, benchmark, qualification and controlled upgrade loop.

## Data-plane profiles

- Nano: portable process, minimal telemetry.
- Edge: portable process + optional kernel acceleration and local autonomy.
- Enterprise: redundant L4/L7 instances, mTLS, Gateway API, centralized observability.
- Hyperscale: Anycast/BGP + XDP/eBPF + DSR + consistent hashing + multi-region control plane.

## Control-plane rule

Desired state is declarative and signed. The data plane must continue forwarding with its last-known-good configuration when the control plane, AI subsystem, or external connectivity is unavailable.

## Algorithm policy

- P2C / least-request: default for heterogeneous request loads.
- Least-connections: long-lived TCP/session workloads.
- Maglev: affinity and minimal disruption requirements.
- Round-robin: homogeneous/simple pools and deterministic fallback.
- Locality/zone aware: control-plane feature layered only where semantics are compatible.

## API Gateway separation

ΩTF owns traffic distribution, availability and network/service steering. API gateways own API authentication, quota, transformation, lifecycle and consumer policy. They may be deployed together but are not the same component.
