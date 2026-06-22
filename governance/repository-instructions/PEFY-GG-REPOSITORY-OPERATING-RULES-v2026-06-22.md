# PEFY-GG Repository Operating Rules

**Operating standard:** PEFY-GG OMNIA FORTRESS 360™  
**Confidentiality level:** Public-safe  
**Owner code:** PGG-SOC-001  
**IP controller:** PGG-IP-CTRL  
**Technical production:** by PEFY-TECH  
**Powered by:** PEFY-GG (GROUP PEFY-CONSULTING GLOBAL)  

---

## 1. Core Workflow

Use this workflow for repository work:

```txt
Understand → Classify → Map → Detect → Correct → Strengthen → Benchmark → Humanize → Validate → Deliver
```

Outputs should be clear, public-safe, maintainable, reusable, and production-ready.

---

## 2. Public-Safe Rule

This repository must not contain restricted internal information, confidential client information, sensitive personal information, private ownership registers, restricted legal evidence, or private operational evidence.

When traceability is needed, use this public-safe block:

```txt
Owner Code: PGG-SOC-001
IP Controller: PGG-IP-CTRL
Asset Registry: PGG-ASSET-REG
Technical Production: by PEFY-TECH
Powered by: PEFY-GG (GROUP PEFY-CONSULTING GLOBAL)
```

---

## 3. Required Quality Gate

Before adding or modifying files, check:

- Purpose and audience are clear.
- Confidentiality level is controlled.
- Ownership and IP notices are respected.
- Risk and failure points are reduced.
- Structure is maintainable.
- Wording is professional and human-readable.
- Unsupported claims and unnecessary filler are removed.
- Setup, usage, testing, or deployment steps are included when relevant.

---

## 4. Repository Structure Standard

When relevant, prefer:

```txt
README.md
LICENSE or license notice
SECURITY.md
CONTRIBUTING.md
docs/
governance/
src/ or app/
tests/
examples/
CHANGELOG.md
```

---

## 5. Code and Documentation Standard

When code or technical documentation is present:

- Prefer modular, maintainable, secure-by-design architecture.
- Separate configuration from restricted operational values.
- Validate inputs and handle errors clearly.
- Add logs or audit trails where useful.
- Add role-based access logic where relevant.
- Prefer low-cost, self-hostable, open-source-friendly components when feasible.
- Include tests or executable verification steps.
- Keep technical choices explainable and reversible.

---

## 6. Failure-Aware Review

Before finalizing, actively look for broken assumptions, missing dependencies, ambiguous naming, weak security posture, poor user experience, missing ownership or license logic, weak deployment instructions, incomplete error handling, unclear business value, and public/private safety failures.

Fix the issue before release.

---

## 7. Reference Tool

Repository operating reference:

```txt
governance/operating-tools/PEFY-GG-OMNIA-FORTRESS-360-OPERATING-TOOL-v2026-06-22.md
```

Default rule:

> Detect what can fail, correct what is weak, protect what creates value, simplify what is noisy, and deliver what can be used immediately.
