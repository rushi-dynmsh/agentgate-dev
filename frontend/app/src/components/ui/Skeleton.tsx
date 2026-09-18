export function TableSkeleton({ rows = 3, columns = 4 }: { rows?: number; columns?: number }) {
  return (
    <table className="ag-table">
      <tbody>
        {Array.from({ length: rows }).map((_, r) => (
          <tr key={r}>
            {Array.from({ length: columns }).map((_, c) => (
              <td key={c}>
                <div
                  style={{
                    height: 12,
                    borderRadius: 4,
                    background: "var(--ag-neutral-soft)",
                    width: c === 0 ? "50%" : c === columns - 1 ? "30%" : "80%",
                    animation: "ag-pulse 1.4s ease-in-out infinite",
                  }}
                />
              </td>
            ))}
          </tr>
        ))}
      </tbody>
      <style>{`
        @keyframes ag-pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.4; }
        }
      `}</style>
    </table>
  );
}
