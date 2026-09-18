import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useGovernance } from "../../state/GovernanceProvider";
import { useToast } from "../../state/ToastProvider";

export function RollbackPolicyPage() {
  const { version } = useParams<{ version: string }>();
  const { state, rollback } = useGovernance();
  const { showToast } = useToast();
  const navigate = useNavigate();

  const [rollingBack, setRollingBack] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleRollback = async () => {
    if (!version) return;
    setRollingBack(true);
    setError(null);
    try {
      await rollback(version);
      showToast(`Rolled back to ${version}`, "success");
      navigate("/policies");
    } catch (err) {
      const message = err instanceof Error ? err.message : "Rollback failed.";
      setError(message);
      showToast(message, "error");
    } finally {
      setRollingBack(false);
    }
  };

  return (
    <div className="ag-page">
      <div className="ag-eyebrow">
        <Link to="/policies">Policies</Link> &nbsp;/&nbsp; Rollback
      </div>
      <div className="ag-page-header">
        <div>
          <h1 className="ag-page-title">Rollback policy</h1>
          <p className="ag-page-subtitle">Restore a previous policy version to active.</p>
        </div>
      </div>

      <div style={{ maxWidth: 480 }}>
        <div style={{ display: "grid", gridTemplateColumns: "140px 1fr", rowGap: 10, fontSize: 13.5, marginBottom: 20 }}>
          <span style={{ color: "var(--ag-text-secondary)" }}>Current active</span>
          <span className="ag-mono">{state.activeVersion ?? "none"}</span>
          <span style={{ color: "var(--ag-text-secondary)" }}>Rollback to</span>
          <span className="ag-mono">{version}</span>
        </div>

        <div className="ag-alert ag-alert-warning">
          This will change authorization behavior immediately for this workspace.
        </div>

        {error && <div className="ag-alert ag-alert-error">{error}</div>}

        <div style={{ display: "flex", justifyContent: "flex-end", gap: 10 }}>
          <Link to="/policies" className="ag-btn ag-btn-ghost">Cancel</Link>
          <button className="ag-btn ag-btn-danger" onClick={handleRollback} disabled={rollingBack}>
            {rollingBack ? "Rolling back…" : "Confirm rollback"}
          </button>
        </div>
      </div>
    </div>
  );
}
