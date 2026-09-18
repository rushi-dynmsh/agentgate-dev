import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useGovernance } from "../../state/GovernanceProvider";
import { useToast } from "../../state/ToastProvider";
import { toValidationDisplayView } from "@contract/view/policy-lifecycle-view";
import type { ValidateResponse } from "@contract/models/governance";

const TEMPLATE = `permit(
  principal in AgentGate::Role::"reader",
  action == AgentGate::Action::"InvokeTool",
  resource
) when {
  resource.risk == "read"
};

permit(
  principal in AgentGate::Role::"admin",
  action == AgentGate::Action::"InvokeTool",
  resource
) when {
  resource.risk == "read" || resource.risk == "write"
};

permit(
  principal in AgentGate::Role::"payer",
  action == AgentGate::Action::"InvokeTool",
  resource
) when {
  resource.risk == "write" &&
  context has amount &&
  context.amount <= 1000
};

forbid(
  principal,
  action == AgentGate::Action::"InvokeTool",
  resource
) when {
  resource.risk == "destructive"
};
`;

export function CreatePolicyPage() {
  const { validateDraft, submitCandidate } = useGovernance();
  const { showToast } = useToast();
  const navigate = useNavigate();

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [content, setContent] = useState(TEMPLATE);
  const [validation, setValidation] = useState<ValidateResponse | null>(null);
  const [validating, setValidating] = useState(false);
  const [saving, setSaving] = useState(false);

  const handleValidate = async () => {
    setValidating(true);
    try {
      setValidation(await validateDraft(content));
    } finally {
      setValidating(false);
    }
  };

  const handleSave = async () => {
    if (!validation?.valid) return;
    setSaving(true);
    try {
      const record = await submitCandidate(content, description || name || "New candidate policy");
      showToast(`Candidate ${record.version} saved`, "success");
      navigate(`/policies/${encodeURIComponent(record.version)}/dry-run`);
    } finally {
      setSaving(false);
    }
  };

  const view = validation ? toValidationDisplayView(validation) : null;

  return (
    <div className="ag-page">
      <div className="ag-eyebrow">
        <Link to="/policies">Policies</Link> &nbsp;/&nbsp; Create candidate
      </div>
      <div className="ag-page-header">
        <div>
          <h1 className="ag-page-title">Create candidate policy</h1>
          <p className="ag-page-subtitle">Write and validate your Cedar policy before it goes anywhere near production.</p>
        </div>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "300px 1fr", gap: 24, alignItems: "start" }}>
        <div>
          <h2 style={{ fontSize: 13, fontWeight: 500, margin: "0 0 12px", color: "var(--ag-text-secondary)" }}>
            Policy details
          </h2>
          <div className="ag-field">
            <label htmlFor="policy-name">Name</label>
            <input
              id="policy-name"
              className="ag-input"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Customer Data Access"
            />
          </div>
          <div className="ag-field">
            <label htmlFor="policy-desc">Description</label>
            <textarea
              id="policy-desc"
              className="ag-input"
              style={{ minHeight: 80, fontFamily: "var(--ag-font-sans)" }}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Controls access to customer data tools."
            />
          </div>
        </div>

        <div>
          <h2 style={{ fontSize: 13, fontWeight: 500, margin: "0 0 12px", color: "var(--ag-text-secondary)" }}>
            Cedar policy
          </h2>
          <textarea
            className="ag-textarea"
            style={{ marginBottom: 14 }}
            value={content}
            onChange={(e) => {
              setContent(e.target.value);
              setValidation(null);
            }}
            spellCheck={false}
          />

          {view && (
            <div className={`ag-alert ${view.tone === "valid" ? "ag-alert-success" : "ag-alert-error"}`}>
              <div>
                <strong>{view.headline}</strong>
                {view.versionLabel && (
                  <div className="ag-mono" style={{ marginTop: 4 }}>version: {view.versionLabel}</div>
                )}
                {view.errorMessages && (
                  <ul style={{ margin: "6px 0 0", paddingLeft: 18 }}>
                    {view.errorMessages.map((e, i) => (
                      <li key={i}>{e}</li>
                    ))}
                  </ul>
                )}
              </div>
            </div>
          )}

          <div style={{ display: "flex", gap: 10, justifyContent: "flex-end" }}>
            <Link to="/policies" className="ag-btn ag-btn-ghost">Cancel</Link>
            <button className="ag-btn ag-btn-secondary" onClick={handleValidate} disabled={validating}>
              {validating ? "Validating…" : "Validate policy"}
            </button>
            <button
              className="ag-btn ag-btn-primary"
              onClick={handleSave}
              disabled={!view || view.tone !== "valid" || saving}
            >
              {saving ? "Saving…" : "Save candidate"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
