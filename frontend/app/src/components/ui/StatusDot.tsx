import type { ReactNode } from "react";

export type StatusTone = "allow" | "deny" | "candidate" | "neutral";

export function StatusDot({ tone, children }: { tone: StatusTone; children: ReactNode }) {
  return (
    <span className={`ag-status ag-status-${tone}`}>
      <span className="ag-status-dot" />
      {children}
    </span>
  );
}
