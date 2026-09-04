# Kubernetes integration status

These manifests describe the target Gateway API contract for ΩTF. The current bootstrap binary is an L4 TCP data plane and **does not yet implement the GatewayClass controller reconciliation loop**.

Production use of these objects requires an ΩTF controller or a qualified adapter (for example Envoy Gateway, Cilium or another conformant implementation) that owns `pefy.gg/omega-traffic-fabric`.

Use the Gateway API Standard Channel by default. Experimental resources or fields must pass ΩEVOLVE qualification before use.
