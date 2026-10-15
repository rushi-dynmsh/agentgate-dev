import { useState } from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "../../state/AuthProvider";
import { usePageTitle } from "../../lib/usePageTitle";

export function LoginPage() {
  const { user, loading, ssoConfigured, login } = useAuth();
  const [error, setError] = useState<string | null>(null);
  usePageTitle("Login");

  if (!loading && user) {
    return <Navigate to="/" replace />;
  }

  const handleSignIn = () => {
    login().catch(() => setError("Could not start sign-in. Please try again."));
  };

  return (
    <div
      style={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: "radial-gradient(700px 360px at 50% -40px, var(--ag-accent-soft), transparent 70%), var(--ag-bg-canvas)",
      }}
    >
      <div
        style={{
          width: 380,
          background: "var(--ag-bg-surface)",
          border: "1px solid var(--ag-border)",
          borderRadius: 14,
          padding: "36px 32px",
          boxShadow: "0 12px 40px rgba(67, 56, 202, 0.10)",
        }}
      >
        {/* Brand mark + wordmark */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 10,
            marginBottom: 12,
          }}
        >
          <img
            src="/favicon.svg"
            alt="AgentGate"
            width={36}
            height={36}
            style={{ display: "block", flexShrink: 0 }}
          />
          <span style={{ fontWeight: 600, fontSize: 20, letterSpacing: "-0.01em", color: "var(--ag-text-primary)" }}>
            AgentGate
          </span>
        </div>

        {/* Tagline */}
        <p style={{ margin: "0 0 6px", fontSize: 13, color: "var(--ag-text-secondary)", lineHeight: 1.5 }}>
          A checkpoint in the only path between an AI agent and its tools.
        </p>
        <p style={{ margin: "0 0 24px", fontSize: 13, color: "var(--ag-text-muted)" }}>
          Sign in with your organization account to manage policies.
        </p>

        {!ssoConfigured ? (
          <div className="ag-alert ag-alert-error">
            <div>
              <strong>SSO not configured.</strong> Set <code>oidcAuthority</code> and <code>oidcClientId</code> in{" "}
              <code>/config.js</code>.
            </div>
          </div>
        ) : (
          <>
            {error && <div className="ag-alert ag-alert-error">{error}</div>}
            <button
              type="button"
              className="ag-btn ag-btn-primary"
              style={{ width: "100%", justifyContent: "center" }}
              onClick={handleSignIn}
            >
              Sign in with SSO
            </button>
          </>
        )}
      </div>
    </div>
  );
}
