---
sidebar_position: 1
---

# Introduction

AgentGate is a checkpoint in the only path between an AI agent and its tools. Every tool call an agent makes passes through it, and AgentGate decides, before the call reaches the tool, whether it is allowed.

## What it does

- **Authorizes every call.** Policies written in Cedar decide which roles may use tools of which risk level. Anything unknown or unclear is denied.
- **Keeps backend credentials away from agents.** Agents present a short-lived token. For allowed calls AgentGate mints a fresh 60-second credential for the backend.
- **Records every decision.** Each decision goes into a tamper-evident hash chain that you can verify independently.
- **Looks like a normal MCP server.** Agents and tools such as MCP Inspector connect to the gateway URL (`/mcp`) as they would to any MCP server.

## Where to go next

| I want to... | Read |
|---|---|
| Run it locally and see a call allowed and blocked | [Quickstart](/docs/quickstart) |
| Understand identity, tools, policies and decisions | [Core Concepts](/docs/concepts) |
| Put AgentGate in front of my own gateway | [Integrate with agentgateway](/docs/integrate-agentgateway) |
| Know what is and is not covered today | [Security and Limitations](/docs/security-and-limitations) |
| Work on AgentGate itself | [Contributor Guide](/docs/developer/getting-started/overview) |
