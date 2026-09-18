import { ComingSoonPage } from "../../components/ui/ComingSoonPage";
import { StatusDot } from "../../components/ui/StatusDot";
import type { IdentityMappingStatus } from "@contract/models/identity-governance";

const FIXTURE_IDENTITIES: {
  agentId: string;
  onBehalfOf?: string;
  roles: string[];
  status: IdentityMappingStatus;
}[] = [
  { agentId: "agent-support", onBehalfOf: "user-123", roles: ["support-agents"], status: "ok" },
  { agentId: "agent-finance", onBehalfOf: "user-456", roles: ["finance"], status: "ok" },
  { agentId: "agent-admin", roles: [], status: "missing_roles" },
];

export function IdentitiesPage() {
  return (
    <ComingSoonPage
      title="Identities"
      subtitle="Agent and on-behalf-of identity mapping status."
      note="Preview data. G2's identity claim-mapping model is real and tested, but there is no GovernanceClient method to fetch live mappings yet."
    >
      <table className="ag-table">
        <thead>
          <tr>
            <th>Agent</th>
            <th>Acting for</th>
            <th>Roles</th>
            <th>Mapping status</th>
          </tr>
        </thead>
        <tbody>
          {FIXTURE_IDENTITIES.map((row) => (
            <tr key={row.agentId}>
              <td className="ag-mono" style={{ color: "var(--ag-text-primary)" }}>{row.agentId}</td>
              <td className="ag-mono">{row.onBehalfOf ?? "—"}</td>
              <td>{row.roles.length > 0 ? row.roles.join(", ") : "—"}</td>
              <td>
                <StatusDot tone={row.status === "ok" ? "allow" : "deny"}>{row.status}</StatusDot>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </ComingSoonPage>
  );
}
