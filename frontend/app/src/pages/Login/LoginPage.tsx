import { useState, type FormEvent } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { ShieldCheck } from "lucide-react";
import { useAuth } from "../../state/AuthProvider";

export function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const [token, setToken] = useState("");
  const [error, setError] = useState<string | null>(null);

  const from = (location.state as { from?: Location })?.from?.pathname ?? "/";

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    const result = login(token);
    if (result.ok) {
      navigate(from, { replace: true });
    } else {
      setError(result.error ?? "Sign-in failed.");
    }
  };

  return (
    <div
      style={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: "var(--ag-bg-canvas)",
      }}
    >
      <form
        onSubmit={handleSubmit}
        style={{
          width: 360,
          background: "var(--ag-bg-surface)",
          border: "1px solid var(--ag-border)",
          borderRadius: 10,
          padding: "32px 28px",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: 9, fontWeight: 600, fontSize: 15, marginBottom: 4 }}>
          <ShieldCheck size={20} strokeWidth={2} />
          AgentGate
        </div>
        <p style={{ margin: "0 0 24px", fontSize: 13, color: "var(--ag-text-secondary)" }}>
          Sign in with the admin token to manage policies.
        </p>

        <div className="ag-field">
          <label htmlFor="admin-token">Admin token</label>
          <input
            id="admin-token"
            type="password"
            className="ag-input"
            value={token}
            onChange={(e) => {
              setToken(e.target.value);
              setError(null);
            }}
            placeholder="Paste your admin token"
            autoFocus
          />
        </div>

        {error && <div className="ag-alert ag-alert-error">{error}</div>}

        <button type="submit" className="ag-btn ag-btn-primary" style={{ width: "100%", justifyContent: "center" }}>
          Sign in
        </button>

        <p style={{ margin: "18px 0 0", fontSize: 11.5, color: "var(--ag-text-muted)", lineHeight: 1.5 }}>
          This checks against <code>VITE_ADMIN_TOKEN</code> when it's set. In local/demo mode with
          no token configured, any non-empty value signs you in.
        </p>
      </form>
    </div>
  );
}
