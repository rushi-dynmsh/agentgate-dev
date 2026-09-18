import type { ReactNode } from "react";
import type { LucideIcon } from "lucide-react";

export function EmptyState({
  icon: Icon,
  title,
  subtitle,
  action,
}: {
  icon: LucideIcon;
  title: string;
  subtitle: string;
  action?: ReactNode;
}) {
  return (
    <div style={{ padding: "56px 0", textAlign: "center" }}>
      <div
        style={{
          width: 40,
          height: 40,
          borderRadius: "50%",
          background: "var(--ag-neutral-soft)",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          margin: "0 auto 16px",
        }}
      >
        <Icon size={18} color="var(--ag-text-secondary)" />
      </div>
      <div style={{ fontSize: 14.5, fontWeight: 500, marginBottom: 6 }}>{title}</div>
      <div style={{ fontSize: 13.5, color: "var(--ag-text-secondary)", marginBottom: action ? 20 : 0 }}>
        {subtitle}
      </div>
      {action}
    </div>
  );
}
