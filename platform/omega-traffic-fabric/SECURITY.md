# Security baseline

## Bootstrap restrictions

The current ΩBALANCER bootstrap is an L4 TCP reference data plane. It does not provide TLS termination, mTLS, WAF, authentication, administrative APIs, eBPF/XDP acceleration or a production-grade dynamic control plane.

Do not expose its metrics endpoint to untrusted networks. Bind observability to a management interface or protect it through an approved security control.

## Required production controls

Before production qualification, add or integrate:

- signed configuration and least-privilege runtime identity;
- secret/certificate management and mTLS where applicable;
- network policy and GARKAEL Ω security enforcement;
- SBOM, provenance, artifact signing and dependency vulnerability review;
- fuzzing and protocol-abuse tests;
- rate/connection limits, overload protection and DDoS controls;
- policy-gated canary rollout and automatic rollback;
- immutable audit evidence for configuration and release changes.

## ΩEVOLVE boundary

Discovery and benchmarking may be automated. Production binaries and traffic policy must not be changed solely because an external project published a new release or because an AI agent recommended a change. Promotion requires the gates defined in `policies/EVOLVE.md`.
