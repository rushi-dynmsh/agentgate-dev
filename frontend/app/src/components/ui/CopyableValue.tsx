import { useState } from "react";
import { Copy, Check } from "lucide-react";

export function CopyableValue({ value }: { value: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      // clipboard unavailable — silently no-op, nothing to recover from
    }
  };

  return (
    <span
      className="ag-mono ag-copyable-value"
      onClick={handleCopy}
      title={copied ? "Copied" : `Copy ${value}`}
    >
      <span className="ag-copyable-text">{value}</span>
      {copied ? (
        <Check size={12} color="var(--ag-allow)" />
      ) : (
        <Copy size={12} style={{ opacity: 0.45 }} />
      )}
    </span>
  );
}
