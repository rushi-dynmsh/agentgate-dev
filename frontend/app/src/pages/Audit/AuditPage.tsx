import { useState } from "react";
import { ComingSoonPage } from "../../components/ui/ComingSoonPage";
import { StatusDot } from "../../components/ui/StatusDot";
import type { Decision, ReasonCode } from "@contract/models/decision";

interface AuditFixtureRow {
  time: string;
  agentId: string;
  onBehalfOf?: string;
  toolName: string;
  decision: Decision;
  reason: ReasonCode;
  policyVersion: string;
  executionId: string;
  rowHash: string;
  prevHash: string;
}

const FIXTURE_ROWS: AuditFixtureRow[] = [
  {
    time: "10:24:01",
    agentId: "agent-support",
    onBehalfOf: "user-jdoe",
    toolName: "customer.read",
    decision: "ALLOW",
    reason: "policy_allow",
    policyVersion: "v1",
    executionId: "req-83920",
    rowHash: "9f2a...c31e",
    prevHash: "7b10...44aa",
  },
  {
    time: "10:24:12",
    agentId: "agent-support",
    onBehalfOf: "user-support",
    toolName: "customer.delete",
    decision: "DENY",
    reason: "policy_deny",
    policyVersion: "v1",
    executionId: "req-83921",
    rowHash: "3c88...0021",
    prevHash: "9f2a...c31e",
  },
  {
    time: "10:23:55",
    agentId: "agent-finance",
    onBehalfOf: "user-mike",
    toolName: "customer.update",
    decision: "DENY",
    reason: "no_matching_policy",
    policyVersion: "v1",
    executionId: "req-83918",
    rowHash: "7b10...44aa",
    prevHash: "1a04...9e02",
  },
];

export function AuditPage() {
  const [selected, setSelected] = useState<AuditFixtureRow>(FIXTURE_ROWS[0]!);

  return (
    <ComingSoonPage
      title="Audit log"
      subtitle="Every authorization decision, allowed or denied."
      note="Not backed yet. Durable audit persistence (Gate G5) is real on the backend — hash-chained, tamper-evident — but there is no REST endpoint to read it back yet. This screen is fixture data shaped like the real record, not a live feed."
    >
      <div style={{ display: "grid", gridTemplateColumns: "1.6fr 1fr", gap: 24 }}>
        <table className="ag-table">
          <thead>
            <tr>
              <th>Time</th>
              <th>Identity</th>
              <th>Tool</th>
              <th>Decision</th>
            </tr>
          </thead>
          <tbody>
            {FIXTURE_ROWS.map((row) => (
              <tr
                key={row.executionId}
                onClick={() => setSelected(row)}
                style={{
                  cursor: "pointer",
                  background: selected.executionId === row.executionId ? "var(--ag-bg-canvas)" : undefined,
                }}
              >
                <td className="ag-mono">{row.time}</td>
                <td className="ag-mono">{row.agentId}</td>
                <td className="ag-mono">{row.toolName}</td>
                <td>
                  <StatusDot tone={row.decision === "ALLOW" ? "allow" : "deny"}>{row.decision}</StatusDot>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        <div className="ag-card">
          <div className="ag-card-header">
            <h2 className="ag-card-title">Event detail</h2>
            <StatusDot tone={selected.decision === "ALLOW" ? "allow" : "deny"}>{selected.decision}</StatusDot>
          </div>
          <div className="ag-card-body">
            <dl style={{ margin: 0, fontSize: 13 }}>
              {[
                ["Time", selected.time],
                ["Agent", selected.agentId],
                ["Acting for", selected.onBehalfOf ?? "—"],
                ["Tool", selected.toolName],
                ["Reason", selected.reason],
                ["Policy version", selected.policyVersion],
                ["Execution ID", selected.executionId],
                ["Row hash", selected.rowHash],
                ["Prev hash", selected.prevHash],
              ].map(([label, value]) => (
                <div key={label} style={{ display: "flex", justifyContent: "space-between", padding: "8px 0", borderBottom: "1px solid var(--ag-border-soft)" }}>
                  <dt style={{ color: "var(--ag-text-secondary)" }}>{label}</dt>
                  <dd className="ag-mono" style={{ margin: 0, textAlign: "right", color: "var(--ag-text-primary)" }}>{value}</dd>
                </div>
              ))}
            </dl>
          </div>
        </div>
      </div>
    </ComingSoonPage>
  );
}
