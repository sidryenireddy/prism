"use client";

import type { CardType } from "@/types";

interface ParameterInputProps {
  cardType: CardType;
  config: Record<string, unknown>;
  value: unknown;
  onChange: (value: unknown) => void;
}

export function ParameterInput({ cardType, config, value, onChange }: ParameterInputProps) {
  const label = (config.label as string) || cardType.replace("param_", "");

  switch (cardType) {
    case "param_object_selection":
      return (
        <div className="p-2">
          <label className="text-xs text-neutral-500 block mb-1">{label}</label>
          <select
            value={String(value ?? "")}
            onChange={(e) => onChange(e.target.value)}
            className="w-full text-xs border border-neutral-200 rounded px-2 py-1 text-black"
          >
            <option value="">Select...</option>
            {((config.options as string[]) || []).map((opt) => (
              <option key={opt} value={opt}>{opt}</option>
            ))}
          </select>
        </div>
      );
    case "param_date_range":
      return (
        <div className="p-2">
          <label className="text-xs text-neutral-500 block mb-1">{label}</label>
          <div className="flex gap-1">
            <input
              type="date"
              className="text-xs border border-neutral-200 rounded px-1 py-0.5 text-black flex-1"
              onChange={(e) => onChange({ ...(value as Record<string, string> || {}), start: e.target.value })}
            />
            <input
              type="date"
              className="text-xs border border-neutral-200 rounded px-1 py-0.5 text-black flex-1"
              onChange={(e) => onChange({ ...(value as Record<string, string> || {}), end: e.target.value })}
            />
          </div>
        </div>
      );
    case "param_numeric":
      return (
        <div className="p-2">
          <label className="text-xs text-neutral-500 block mb-1">{label}</label>
          <input
            type="number"
            value={String(value ?? config.defaultValue ?? "")}
            onChange={(e) => onChange(parseFloat(e.target.value))}
            className="w-full text-xs border border-neutral-200 rounded px-2 py-1 text-black"
          />
        </div>
      );
    case "param_string":
      return (
        <div className="p-2">
          <label className="text-xs text-neutral-500 block mb-1">{label}</label>
          <input
            type="text"
            value={String(value ?? config.defaultValue ?? "")}
            onChange={(e) => onChange(e.target.value)}
            className="w-full text-xs border border-neutral-200 rounded px-2 py-1 text-black"
          />
        </div>
      );
    case "param_boolean":
      return (
        <div className="p-2 flex items-center gap-2">
          <label className="text-xs text-neutral-500">{label}</label>
          <button
            onClick={() => onChange(!value)}
            className={`w-8 h-4 rounded-full transition-colors ${value ? "bg-red-600" : "bg-neutral-300"}`}
          >
            <div className={`w-3 h-3 bg-white rounded-full transition-transform ${value ? "translate-x-4" : "translate-x-0.5"}`} />
          </button>
        </div>
      );
    default:
      return null;
  }
}
