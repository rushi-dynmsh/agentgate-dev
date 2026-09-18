# AgentGate — a plain-language guide to the screen we've built

*No technical background needed.*

### What AgentGate actually does

Companies are letting AI agents take real actions on real systems — reading customer records, updating databases, sending emails. AgentGate sits in the middle of every one of those actions and asks one question before anything happens: **is this agent allowed to do this, right now?** This is a tour of the screen a human uses to set those rules and watch what the agents are doing.

**01** An AI agent tries to do something — say, "delete this customer record."
**02** AgentGate checks the request against the rules a human wrote.
**03** It's allowed, or it's blocked — decided instantly, before the action happens.
**04** Either way, it's written down — so someone can review it later.

### On this page
1. [Home base](#01-home-base)
2. [The rulebook](#02-the-rulebook)
3. [Writing a new rule](#03-writing-a-new-rule)
4. [Undoing a change](#04-undoing-a-change)
5. [What agents can touch](#05-what-agents-can-touch)
6. [Who's been active](#06-whos-been-active)
7. [The paper trail](#07-the-paper-trail)
8. [Settings](#08-settings)
9. [A few honest notes](#09-a-few-honest-notes)

---

## 01 · Home base

The first thing you see after opening the app: a quick summary of the workspace — how many policies exist, which version is currently live, how many drafts are waiting to be tested, and what the active policy actually says.

> *Recent policies* and *Active policy* on this screen are real — they reflect whatever is genuinely in the system right now, not sample numbers.

---

## 02 · The rulebook

This is the actual rulebook: every policy that's ever been written, and which one is in force right now. Every policy has three possible lives — a draft still being tested (**Candidate**), the one version currently enforced (**Active**), and older versions kept around in case you need to undo a change (**Historical**).

Each row shows the version, its status (with a small colored dot — green for active, amber for candidate), a plain-language description of what it does, and when it last changed.

---

## 03 · Writing a new rule

Changing the rulebook is the single most powerful thing a person can do here — it can open up or shut down what every AI agent in the company can do. So the screen walks whoever's doing it through four checkpoints before anything goes live.

**1. Write** — describe the rule and write the actual policy.
**2. Check it's valid** — before anything else happens, the system confirms the policy is well-formed and tells you exactly what's wrong if it isn't.
**3. Test it safely** — a "dry run" runs the candidate against a set of realistic sample requests and shows exactly which decisions would change, without switching anything on yet.
**4. Turn it on** — only after reviewing the test results does a human confirm and activate it, with a required reason for the record.

> *Why the extra steps?* A rule change is a production change — get it wrong and you could lock everyone out, or open a door that should've stayed shut. Testing it against real scenarios first, and keeping every old version on hand to instantly undo a change, is the whole point of this flow.

> All four of these steps are fully real — every validation result and every dry-run comparison comes from the system's actual policy logic, not a canned response.

---

## 04 · Undoing a change

Every historical version stays on hand. If a policy needs to be undone, this screen shows exactly what's currently active and what you're rolling back to, with a plain warning that the change takes effect immediately.

---

## 05 · What agents can touch

Every tool an AI agent could call — reading a record, updating one, deleting one — is classified here by how risky it is: **read**, **write**, or **destructive**. Anything not yet classified, or whose schema has quietly changed since it was registered, is flagged rather than trusted by default.

> *Honest note:* this screen shows realistic example data. The underlying classification system is real and already built on the backend, but there's no live connection wiring it into this screen yet — that's the next piece to land, not a design gap.

---

## 06 · Who's been active

Not a list of accounts you create — the system doesn't manage logins for anyone. It's a record of which AI agents, and which humans they were acting on behalf of, have been mapped into the system.

> *Honest note:* same situation as the tools screen — real underlying data model, sample data shown here until it's wired up live.

---

## 07 · The paper trail

Every decision — allowed or denied — is meant to leave a permanent record: who, what, when, and why, plus a tamper-evident chain so a record can't be quietly altered after the fact. This is what a compliance review, or an investigation after something goes wrong, would look at. Clicking any row shows the full detail for that one decision.

> *Honest note:* this one's a bit further out than the others. The tamper-evident record-keeping itself is real and already built on the backend — but there isn't yet a way for this screen to actually read those records back. What you see here is realistic sample data standing in for that connection.

---

## 08 · Settings

Deliberately small and read-only for now: which policy engine is running, how the gateway connects, how the database is configured. This isn't meant to be a control panel non-technical staff adjust day to day — it's a status view.

> *Honest note:* there's also no full sign-in system yet. Right now the whole app is protected by a single shared access key rather than individual logins — a proper system (the kind that plugs into however your company already signs people into email or chat) is a deliberate next step, not an oversight.

---

## 09 · A few honest notes

- **The rulebook and everything in "Writing a new rule" are fully real.** Every validation result and dry-run comparison you see came from the system's actual logic — nothing there is a canned response.
- **Tools, identities, and the audit trail show realistic example data for now.** The systems that would feed them for real are either already built and just not connected yet (tools, identities), or waiting on one more piece before they even can be (audit log).
- **There's no individual sign-in yet, and no "try it yourself" simulator screen yet either.** Both are known next steps, not things anyone forgot.
- The short version: this is the checkpoint every AI agent's action is meant to pass through, the rulebook a human controls, and — soon — the record of everything that happened. What's above is what that looks like today: most of the core rule-making loop working for real, with a few screens still waiting on their data source to come online.
