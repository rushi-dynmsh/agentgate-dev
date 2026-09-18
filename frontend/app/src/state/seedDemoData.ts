import type { GovernanceClient } from "@contract/api/governanceClient";

const ACTIVE_POLICY = `// Customer Data Access — v1
permit (
  principal in Group::"support-agents",
  action == Action::"customer.read",
  resource
);

forbid (
  principal,
  action == Action::"customer.delete",
  resource
);
`;

const CANDIDATE_POLICY = `// Customer Data Access — v2 (candidate)
permit (
  principal in Group::"support-agents",
  action == Action::"customer.read",
  resource
);

permit (
  principal in Group::"support-agents",
  action == Action::"customer.update",
  resource
);

forbid (
  principal,
  action == Action::"customer.delete",
  resource
);
`;

let seeded = false;

/**
 * One-time demo seed so the Policies list isn't empty on first load.
 * MockGovernanceClient starts with zero policies per workspace by design
 * (it's a deterministic test double, not a UI-specific fixture set) — this
 * just gives the UI something real to render against on a fresh app boot.
 */
export async function seedDemoPolicies(client: GovernanceClient, workspaceId: string): Promise<void> {
  if (seeded) return;
  seeded = true;

  const existing = await client.listPolicies(workspaceId);
  if (existing.length > 0) return;

  const active = await client.createCandidate(
    workspaceId,
    ACTIVE_POLICY,
    "Controls access to customer data tools."
  );
  await client.activatePolicy(workspaceId, active.version);

  await client.createCandidate(
    workspaceId,
    CANDIDATE_POLICY,
    "Adds update permission for support agents."
  );
}
