export interface AgentGateConfig {
  apiUrl: string;
  oidcAuthority: string;
  oidcClientId: string;
  oidcScope: string;
  /**
   * The `iss` claim value the IdP actually stamps into tokens. Deliberately
   * separate from oidcAuthority: in a typical Docker deployment, the browser
   * must reach the IdP via a different, publicly-routable host/port than the
   * one backend services use (and that the IdP's own fixed hostname config
   * uses for the issuer string), so the two values can legitimately differ.
   * Falls back to oidcAuthority when unset (single-hostname deployments).
   */
  oidcIssuer: string;
}

declare global {
  interface Window {
    __AGENTGATE_CONFIG__?: Partial<AgentGateConfig>;
  }
}

/**
 * Runtime config from /config.js (window.__AGENTGATE_CONFIG__). Falls back to
 * the non-secret VITE_API_URL dev value for apiUrl only; OIDC values have no
 * fallback so an unconfigured deployment surfaces "SSO not configured".
 */
export function getConfig(): AgentGateConfig {
  const raw = (typeof window !== "undefined" ? window.__AGENTGATE_CONFIG__ : undefined) ?? {};
  const s = (v: unknown) => (typeof v === "string" ? v.trim() : "");
  return {
    apiUrl: s(raw.apiUrl) || (import.meta.env.VITE_API_URL as string | undefined) || "",
    oidcAuthority: s(raw.oidcAuthority),
    oidcClientId: s(raw.oidcClientId),
    oidcScope: s(raw.oidcScope) || "openid profile",
    oidcIssuer: s(raw.oidcIssuer) || s(raw.oidcAuthority),
  };
}
