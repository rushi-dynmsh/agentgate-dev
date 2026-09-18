import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useGovernance } from "../../state/GovernanceProvider";
import { useToast } from "../../state/ToastProvider";
import { toValidationDisplayView } from "@contract/view/policy-lifecycle-view";
import type { ValidateResponse } from "@contract/models/governance";

type PolicyRole = "reader" | "admin" | "payer" | "all";
type PolicyEffect = "allow" | "deny";
type Risk = "read" | "write" | "destructive";

interface PolicyRule {
  id: number;
  effect: PolicyEffect;
  role: PolicyRole;
  risks: Risk[];
  amountLimit: string;
}

const DEFAULT_RULES: PolicyRule[] = [
  { id: 1, effect: "allow", role: "reader", risks: ["read"], amountLimit: "" },
  { id: 2, effect: "allow", role: "admin", risks: ["read", "write"], amountLimit: "" },
];

function cedarRule(rule: PolicyRule): string {
  const principal = rule.role === "all" ? "principal" : `principal in AgentGate::Role::"${rule.role}"`;
  const effect = rule.effect === "allow" ? "permit" : "forbid";
  const riskCondition = rule.risks.length === 1
    ? `resource.risk == "${rule.risks[0]}"`
    : rule.risks.map((risk) => `resource.risk == "${risk}"`).join(" || ");
  const amountCondition = rule.role === "payer" && rule.amountLimit.trim()
    ? ` &&\n  context has amount &&\n  context.amount <= ${rule.amountLimit.trim()}`
    : "";

  return `${effect}(
  ${principal},
  action == AgentGate::Action::"InvokeTool",
  resource
) when {
  ${riskCondition}${amountCondition}
};`;
}

function generateCedar(rules: PolicyRule[]): string {
  return rules.filter((rule) => rule.risks.length > 0).map(cedarRule).join("\n\n") + "\n";
}

function ruleSummary(rule: PolicyRule): string {
  const effect = rule.effect === "allow" ? "Allow" : "Deny";
  const role = rule.role === "all" ? "all roles" : rule.role;
  const risks = rule.risks.length > 0 ? rule.risks.join(", ") : "no tool risk selected";
  const limit = rule.role === "payer" && rule.amountLimit.trim()
    ? ` up to amount ${rule.amountLimit.trim()}`
    : "";
  return `${effect} ${role} to use ${risks} tools${limit}`;
}

export function CreatePolicyPage() {
  const { validateDraft, submitCandidate } = useGovernance();
  const { showToast } = useToast();
  const navigate = useNavigate();

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [rules, setRules] = useState<PolicyRule[]>(DEFAULT_RULES);
  const [advancedMode, setAdvancedMode] = useState(false);
  const [content, setContent] = useState(() => generateCedar(DEFAULT_RULES));
  const [validation, setValidation] = useState<ValidateResponse | null>(null);
  const [validating, setValidating] = useState(false);
  const [saving, setSaving] = useState(false);

  const updateRules = (nextRules: PolicyRule[]) => {
    setRules(nextRules);
    setContent(generateCedar(nextRules));
    setValidation(null);
  };

  const updateRule = (id: number, patch: Partial<PolicyRule>) => {
    updateRules(rules.map((rule) => (rule.id === id ? { ...rule, ...patch } : rule)));
  };

  const toggleRisk = (rule: PolicyRule, risk: Risk) => {
    const risks = rule.risks.includes(risk)
      ? rule.risks.filter((item) => item !== risk)
      : [...rule.risks, risk];
    updateRule(rule.id, { risks });
  };

  const addRule = () => {
    const nextId = Math.max(0, ...rules.map((rule) => rule.id)) + 1;
    updateRules([...rules, { id: nextId, effect: "allow", role: "reader", risks: ["read"], amountLimit: "" }]);
  };

  const removeRule = (id: number) => updateRules(rules.filter((rule) => rule.id !== id));

  const toggleAdvancedMode = () => {
    if (advancedMode) {
      setContent(generateCedar(rules));
      setValidation(null);
    }
    setAdvancedMode(!advancedMode);
  };

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

      <div className="ag-policy-builder-layout">
        <div className="ag-policy-builder-panel">
          <h2 className="ag-form-section-title">
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
              style={{ minHeight: 100, fontFamily: "var(--ag-font-sans)" }}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Controls access to customer data tools."
            />
          </div>
        </div>

        <div className="ag-policy-builder-panel">
          <div className="ag-builder-heading">
            <div>
              <h2 className="ag-form-section-title">Policy rules</h2>
              <p className="ag-builder-help">Describe access with clear controls. AgentGate generates the Cedar source for you.</p>
            </div>
            <button type="button" className="ag-btn ag-btn-secondary ag-btn-sm" onClick={addRule}>+ Add rule</button>
          </div>

          {!advancedMode && (
            <div className="ag-rule-list">
              {rules.map((rule, index) => (
                <div className="ag-rule-card" key={rule.id}>
                  <div className="ag-rule-card-header">
                    <div>
                      <span className="ag-rule-number">Rule {index + 1}</span>
                      <p className="ag-rule-summary">{ruleSummary(rule)}</p>
                    </div>
                    <button type="button" className="ag-rule-remove" onClick={() => removeRule(rule.id)} disabled={rules.length === 1}>
                      Remove
                    </button>
                  </div>
                  <div className="ag-rule-controls">
                    <label className="ag-rule-field">
                      <span>Decision</span>
                      <select className="ag-input" value={rule.effect} onChange={(event) => updateRule(rule.id, { effect: event.target.value as PolicyEffect })}>
                        <option value="allow">Allow</option>
                        <option value="deny">Deny</option>
                      </select>
                    </label>
                    <label className="ag-rule-field">
                      <span>Role</span>
                      <select className="ag-input" value={rule.role} onChange={(event) => updateRule(rule.id, { role: event.target.value as PolicyRole })}>
                        <option value="reader">Reader</option>
                        <option value="admin">Admin</option>
                        <option value="payer">Payer</option>
                        <option value="all">All roles</option>
                      </select>
                    </label>
                    {rule.role === "payer" && rule.effect === "allow" && (
                      <label className="ag-rule-field ag-rule-limit">
                        <span>Amount limit</span>
                        <input className="ag-input" type="number" min="0" value={rule.amountLimit} placeholder="1000" onChange={(event) => updateRule(rule.id, { amountLimit: event.target.value })} />
                      </label>
                    )}
                  </div>
                  <div className="ag-rule-field">
                    <span>Tool risk</span>
                    <div className="ag-risk-options">
                      {(["read", "write", "destructive"] as Risk[]).map((risk) => (
                        <label className="ag-risk-option" key={risk}>
                          <input type="checkbox" checked={rule.risks.includes(risk)} onChange={() => toggleRisk(rule, risk)} />
                          {risk}
                        </label>
                      ))}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}

          <div className="ag-manual-policy-callout">
            <div>
              <strong>{advancedMode ? "Manual Cedar policy" : "Need a custom rule?"}</strong>
              <p>
                {advancedMode
                  ? "Write any Cedar policy supported by AgentGate. The backend remains the final validator."
                  : "Use the manual editor when your rule needs a custom role, tool condition, or argument."}
              </p>
            </div>
            <button type="button" className="ag-btn ag-btn-secondary ag-btn-sm" onClick={toggleAdvancedMode}>
              {advancedMode ? "Back to rule builder" : "Write manually"}
            </button>
          </div>

          {advancedMode && (
            <div className="ag-manual-editor">
              <div className="ag-generated-policy-heading">Cedar source <span>Manual editor</span></div>
              <textarea
                className="ag-textarea"
                value={content}
                onChange={(e) => {
                  setContent(e.target.value);
                  setValidation(null);
                }}
                spellCheck={false}
              />
            </div>
          )}

          {!advancedMode && (
            <div className="ag-generated-policy">
              <div className="ag-generated-policy-heading">Generated Cedar <span>Read-only preview</span></div>
              <pre>{content}</pre>
            </div>
          )}

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
