import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { AlertTriangle } from "lucide-react";
import { useGovernance } from "../../state/GovernanceProvider";
import { useToast } from "../../state/ToastProvider";

export function ActivatePolicyPage() {
  const { version } = useParams<{ version: string }>();
  const { state, activate } = useGovernance();
  const { showToast } = useToast();
  const navigate = useNavigate();

  const [confirmed, setConfirmed] = useState(false);
  const [reason, setReason] = useState("");
  const [activating, setActivating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleActivate = async () => {
    if (!version) return;
    setActivating(true);
    setError(null);
    try {
      await activate(version);
      showToast(`Policy ${version} activated`, "success");
      navigate("/policies");
    } catch (err) {
      const message = err instanceof Error ? err.message : "Activation failed.";
      setError(message);
      showToast(message, "error");
    } finally {
      setActivating(false);
    }
  };

  return (
    <div className="ag-page">
      <div className="ag-eyebrow">
        <Link to="/policies">Policies</Link> &nbsp;/&nbsp; <span className="ag-mono">{version}</span> &nbsp;/&nbsp; Activate
      </div>
      <div className="ag-page-header">
        <div>
          <h1 className="ag-page-title">Activate policy {version}</h1>
          <p className="ag-page-subtitle">Review and confirm activation.</p>
        </div>
      </div>

      <div style={{ maxWidth: 480 }}>
        <div style={{ display: "grid", gridTemplateColumns: "140px 1fr", rowGap: 10, fontSize: 13.5, marginBottom: 20 }}>
          <span style={{ color: "var(--ag-text-secondary)" }}>Currently active</span>
          <span className="ag-mono">{state.activeVersion ?? "none"}</span>
          <span style={{ color: "var(--ag-text-secondary)" }}>New version</span>
          <span className="ag-mono">{version}</span>
        </div>

        <div className="ag-alert ag-alert-warning">
          <AlertTriangle size={16} style={{ flexShrink: 0, marginTop: 2 }} />
          <div>
            This will replace the current active policy. This action immediately affects all
            authorization decisions in this workspace.
          </div>
        </div>

        {error && <div className="ag-alert ag-alert-error">{error}</div>}

        <div className="ag-field">
          <label htmlFor="activation-reason">Reason for activation</label>
          <textarea
            id="activation-reason"
            className="ag-input"
            style={{ minHeight: 70, fontFamily: "var(--ag-font-sans)" }}
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="Approved for production"
          />
        </div>

        <label style={{ display: "flex", gap: 9, fontSize: 13.5, alignItems: "flex-start", marginBottom: 24 }}>
          <input
            type="checkbox"
            checked={confirmed}
            onChange={(e) => setConfirmed(e.target.checked)}
            style={{ marginTop: 3 }}
          />
          I confirm I have reviewed the dry run results and understand the impact.
        </label>

        <div style={{ display: "flex", justifyContent: "flex-end", gap: 10 }}>
          <Link to="/policies" className="ag-btn ag-btn-ghost">Cancel</Link>
          <button className="ag-btn ag-btn-danger" onClick={handleActivate} disabled={!confirmed || activating}>
            {activating ? "Activating…" : "Activate policy"}
          </button>
        </div>
      </div>
    </div>
  );
}
