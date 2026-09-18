import { useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Search, X } from "lucide-react";
import { useGovernance } from "../../state/GovernanceProvider";

export function QuickSearch() {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const { state } = useGovernance();
  const navigate = useNavigate();
  const inputRef = useRef<HTMLInputElement>(null);

  const matches = query.trim()
    ? state.policies.filter(
        (p) =>
          p.version.toLowerCase().includes(query.toLowerCase()) ||
          p.description.toLowerCase().includes(query.toLowerCase())
      )
    : [];

  const openSearch = () => {
    setOpen(true);
    setTimeout(() => inputRef.current?.focus(), 0);
  };

  const closeSearch = () => {
    setOpen(false);
    setQuery("");
  };

  const goToPolicy = () => {
    navigate("/policies");
    closeSearch();
  };

  if (!open) {
    return (
      <button
        onClick={openSearch}
        aria-label="Search policies"
        style={{ background: "none", border: "none", cursor: "pointer", display: "flex", padding: 0 }}
      >
        <Search size={16} color="var(--ag-text-secondary)" />
      </button>
    );
  }

  return (
    <div style={{ position: "relative" }}>
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 6,
          border: "1px solid var(--ag-border-strong)",
          borderRadius: "var(--ag-radius-control)",
          padding: "5px 9px",
          width: 220,
        }}
      >
        <Search size={14} color="var(--ag-text-secondary)" />
        <input
          ref={inputRef}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && matches.length > 0) goToPolicy();
            if (e.key === "Escape") closeSearch();
          }}
          placeholder="Jump to a policy version…"
          style={{ border: "none", outline: "none", fontSize: 12.5, flexGrow: 1, fontFamily: "inherit" }}
        />
        <button
          onClick={closeSearch}
          aria-label="Close search"
          style={{ background: "none", border: "none", cursor: "pointer", display: "flex", padding: 0 }}
        >
          <X size={13} color="var(--ag-text-muted)" />
        </button>
      </div>

      {query.trim() && (
        <div
          style={{
            position: "absolute",
            top: "calc(100% + 6px)",
            right: 0,
            width: 260,
            background: "var(--ag-bg-surface)",
            border: "1px solid var(--ag-border)",
            borderRadius: 8,
            boxShadow: "0 4px 16px rgba(0,0,0,0.08)",
            overflow: "hidden",
            zIndex: 20,
          }}
        >
          {matches.length === 0 ? (
            <div style={{ padding: "12px 14px", fontSize: 12.5, color: "var(--ag-text-secondary)" }}>
              No matching policy.
            </div>
          ) : (
            matches.map((p) => (
              <div
                key={p.version}
                onClick={goToPolicy}
                style={{ padding: "10px 14px", fontSize: 12.5, cursor: "pointer" }}
                onMouseDown={(e) => e.preventDefault()}
              >
                <span className="ag-mono">{p.version}</span>
                <span style={{ color: "var(--ag-text-secondary)", marginLeft: 8 }}>{p.description}</span>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}
