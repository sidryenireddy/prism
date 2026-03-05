"use client";

interface FormulaEditorProps {
  expression: string;
  onChange: (expr: string) => void;
  mode: string;
  onModeChange: (mode: string) => void;
}

const FORMULA_HELP = [
  "SUM(field)", "COUNT()", "AVG(field)", "MIN(field)", "MAX(field)",
  "IF(cond, then, else)", "DAYS_BETWEEN(d1, d2)", "MONTH(date)", "YEAR(date)",
  "+  -  *  /  =  !=  <  >  <=  >=",
];

export function FormulaEditor({ expression, onChange, mode, onModeChange }: FormulaEditorProps) {
  return (
    <div className="p-2 space-y-2">
      <div className="flex gap-1">
        <button
          onClick={() => onModeChange("aggregate")}
          className={`px-2 py-0.5 text-xs rounded ${mode === "aggregate" ? "bg-red-600 text-white" : "bg-neutral-100 text-neutral-600"}`}
        >
          Aggregate
        </button>
        <button
          onClick={() => onModeChange("per_row")}
          className={`px-2 py-0.5 text-xs rounded ${mode === "per_row" ? "bg-red-600 text-white" : "bg-neutral-100 text-neutral-600"}`}
        >
          Per Row
        </button>
      </div>
      <textarea
        value={expression}
        onChange={(e) => onChange(e.target.value)}
        className="w-full h-20 px-2 py-1.5 border border-neutral-300 rounded text-xs text-black font-mono resize-y bg-neutral-50"
        placeholder="e.g. SUM(amount) / COUNT()"
        spellCheck={false}
      />
      <div className="text-[10px] text-neutral-400 leading-tight">
        {FORMULA_HELP.map((h, i) => (
          <span key={i} className="mr-2">{h}</span>
        ))}
      </div>
    </div>
  );
}
