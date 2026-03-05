"use client";

interface NumericDisplayProps {
  value: number;
  label: string;
  type: string;
}

export function NumericDisplay({ value, label, type }: NumericDisplayProps) {
  const formatted = typeof value === "number"
    ? value % 1 === 0 ? value.toLocaleString() : value.toLocaleString(undefined, { maximumFractionDigits: 2 })
    : "—";

  return (
    <div className="flex flex-col items-center justify-center p-4">
      <div className="text-3xl font-bold text-black tabular-nums">{formatted}</div>
      <div className="text-xs text-neutral-500 mt-1">{label || type}</div>
    </div>
  );
}
