import type { ReactNode } from "react";

export function ComingSoonPage({
  title,
  subtitle,
  note,
  children,
}: {
  title: string;
  subtitle: string;
  note: string;
  children?: ReactNode;
}) {
  return (
    <div className="ag-page">
      <div className="ag-page-header">
        <div>
          <h1 className="ag-page-title">{title}</h1>
          <p className="ag-page-subtitle">{subtitle}</p>
        </div>
      </div>

      <div className="ag-alert ag-alert-warning">{note}</div>

      {children}
    </div>
  );
}
