# AgentGate Demo Scenarios & Execution Guide

Welcome to the end-to-end execution guide for AgentGate. This repository contains everything needed to demonstrate the project from scratch, even if you have never inspected the underlying code.

---

## What is AgentGate? (The 30-Second Elevator Pitch)
> "When AI agents invoke tools, traditional firewalls and IAM cannot see what the agent is doing. AgentGate sits as a zero-trust checkpoint directly between the AI agent and the tools. It inspects every request in-flight, enforces Cedar security policies, mints short-lived credentials so the agent never holds raw backend keys, and records every decision into a tamper-evident audit hash chain."

---

## A note on exact wording in these scenarios

Each scenario shows example terminal output and UI text to help you know what to expect. Two things are guaranteed and verified against the real code: the HTTP status code (e.g. `403`), and that the Audit page shows a new row with the right decision (ALLOW/DENY) and tool name. The exact JSON field names and English wording in the printed examples have not all been individually re-verified against the running server and may differ slightly from what you actually see. If the on-screen result differs in wording but matches on status code and ALLOW/DENY, trust the screen, not this document, and keep going — don't stop to debug a wording mismatch live.

---

## System Components Quick Map

When the demo stack runs, here is what is running on your machine:

| Component | Port / URL | Description |
| :--- | :--- | :--- |
| **Admin Console** | `https://localhost:8444` | The web UI for security operators (Audit logs, Tool catalog, Policies). |
| **Documentation & Landing** | `http://localhost:3000` | The public landing page with interactive attack scenario player. |
| **Agent API Gateway** | `https://localhost:3443/mcp` | The entrypoint for agents making tool calls (Envoy / agentgateway). |
| **AgentGate Control Plane** | `https://localhost:9443` | The core Go engine evaluating Cedar policies and minting tokens. |
| **Keycloak (IdP)** | `https://localhost:8443` | Enterprise Identity Provider issuing agent and operator credentials. |
| **PostgreSQL** | `localhost:5432` (internal) | Storage for policies, tool schemas, and cryptographic audit logs. |
| **Support Desk MCP** | `localhost:9200` (internal) | Mock backend tools server (`list_tickets`, `delete_ticket`, etc.). |
| **Demo Agent Worker** | (Background container) | Continuously sends simulated agent traffic so the dashboard is live. |

---

**Shell assumption:** every command in these scenarios is Windows PowerShell (the curl calls use PowerShell-specific quoting). If you're on macOS/Linux, the same calls work with `curl` directly, but you'll need to re-quote the JSON body with single quotes instead of the escaped-double-quote style shown here.

## Before you start

Do the setup, start-up and health checks in **[`docs/DEMO_START_TO_FINISH.md`](../DEMO_START_TO_FINISH.md)** (Parts 1 to 3): Docker, the one-time hosts-file fix for the admin console login, `demo-mode.ps1 -Reset`, the documentation site, and the quick health check. It is the single place for setup and troubleshooting, so this guide doesn't repeat it. Come back here once the health check passes.

For what to *say*, use [`docs/DEMO_SCRIPT.md`](../DEMO_SCRIPT.md). The scenarios below hold the exact commands, expected results and explanations.

---

## The 5 Demo Scenarios

Scenarios 1 to 3 are driven by hand from MCP Inspector plus the admin console, and each call is logged. Run them in order or jump to any one. The automatic traffic client (`prod-demo-agent`) is optional and is covered in `DEMO_START_TO_FINISH.md` Part 5.

1. **[Scenario 1: A Real MCP Client (MCP Inspector) — Allowed, Blocked, Invented Tool](./01-real-client-inspector.md)**
   *The main proof. The official MCP Inspector tool connects to the gateway, lists tools, runs an allowed call, gets blocked on a forbidden one, and a made-up tool is refused. Includes a curl fallback.*

2. **[Scenario 2: Live Tool Risk Reclassification (Dynamic Governance)](./02-live-tool-reclassification.md)**
   *Changing a tool's risk level live in the console and watching the very same call flip from allowed to blocked, without restarting anything.*

3. **[Scenario 3: Policy Lifecycle (Dry-Run & Instant Rollback)](./03-policy-dryrun-and-rollback.md)**
   *Building a policy change with no-code controls (no Cedar typing), testing its impact with a dry run against 5 representative sample requests, and rolling back.*

4. **[Scenario 4: Cryptographic Audit Hash Verification](./04-cryptographic-audit-verify.md)**
   *Running the CLI verification tool to prove the audit log hasn't been tampered with.*

5. **[Scenario 5: Public Portal & Offline Attack Player](./05-offline-and-attack-player.md)**
   *The public portal, offline local search, and the interactive attack-scenario stepper.*
