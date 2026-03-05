"use client";

import { useState } from "react";
import type { Card, CardResult } from "@/types";
import { VIZ_CARD_TYPES, NUMERIC_CARD_TYPES, PARAM_CARD_TYPES } from "@/types";
import { CardResultRenderer } from "./cards/CardResultRenderer";
import { api } from "@/lib/api";

interface CardConfigPanelProps {
  card: Card;
  result?: CardResult;
  onUpdate: (data: Partial<Card>) => void;
  onClose: () => void;
}

export function CardConfigPanel({
  card,
  result,
  onUpdate,
  onClose,
}: CardConfigPanelProps) {
  const [aiPrompt, setAiPrompt] = useState("");
  const [aiLoading, setAiLoading] = useState(false);
  const [rawJson, setRawJson] = useState(JSON.stringify(card.config, null, 2));

  const updateConfig = (patch: Record<string, unknown>) => {
    const newConfig = { ...card.config, ...patch };
    setRawJson(JSON.stringify(newConfig, null, 2));
    onUpdate({ config: newConfig });
  };

  const handleAiConfigure = async () => {
    if (!aiPrompt.trim()) return;
    setAiLoading(true);
    try {
      const updated = await api.aiConfigure(card.id, aiPrompt);
      onUpdate({ config: updated.config });
      setRawJson(JSON.stringify(updated.config, null, 2));
      setAiPrompt("");
    } catch {
      // silently fail
    } finally {
      setAiLoading(false);
    }
  };

  return (
    <div className="w-80 border-l border-neutral-200 bg-white overflow-y-auto shrink-0">
      <div className="p-3 border-b border-neutral-200 flex items-center justify-between">
        <h3 className="text-sm font-semibold text-black">Configure</h3>
        <button
          onClick={onClose}
          className="text-neutral-400 hover:text-black text-sm"
        >
          x
        </button>
      </div>

      <div className="p-3 space-y-3">
        <div>
          <label className="block text-xs font-medium text-neutral-500 mb-1">Type</label>
          <div className="text-sm text-black">{card.card_type.replace(/_/g, " ")}</div>
        </div>

        <div>
          <label className="block text-xs font-medium text-neutral-500 mb-1">Label</label>
          <input
            type="text"
            value={card.label}
            onChange={(e) => onUpdate({ label: e.target.value })}
            className="w-full px-2 py-1.5 border border-neutral-300 rounded text-sm text-black"
            placeholder="Card label"
          />
        </div>

        {/* Type-specific config */}
        <TypeSpecificConfig card={card} updateConfig={updateConfig} />

        {/* Raw JSON config */}
        <div>
          <label className="block text-xs font-medium text-neutral-500 mb-1">Config JSON</label>
          <textarea
            value={rawJson}
            onChange={(e) => {
              setRawJson(e.target.value);
              try {
                const config = JSON.parse(e.target.value);
                onUpdate({ config });
              } catch {
                // invalid JSON
              }
            }}
            className="w-full px-2 py-1.5 border border-neutral-300 rounded text-xs text-black font-mono h-24 resize-y"
          />
        </div>

        {/* AI Configure */}
        <div>
          <label className="block text-xs font-medium text-neutral-500 mb-1">AI Configure</label>
          <div className="flex gap-1">
            <input
              type="text"
              value={aiPrompt}
              onChange={(e) => setAiPrompt(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleAiConfigure()}
              placeholder="Describe changes..."
              className="flex-1 px-2 py-1.5 border border-neutral-300 rounded text-xs text-black"
            />
            <button
              onClick={handleAiConfigure}
              disabled={aiLoading}
              className="px-2 py-1.5 bg-red-600 text-white text-xs rounded hover:bg-red-700 disabled:opacity-50"
            >
              {aiLoading ? "..." : "AI"}
            </button>
          </div>
        </div>

        {/* Result */}
        {result && (
          <div>
            <label className="block text-xs font-medium text-neutral-500 mb-1">Result</label>
            <div className="border border-neutral-200 rounded overflow-hidden">
              <CardResultRenderer
                cardType={card.card_type}
                result={result}
                config={card.config}
              />
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function TypeSpecificConfig({
  card,
  updateConfig,
}: {
  card: Card;
  updateConfig: (patch: Record<string, unknown>) => void;
}) {
  const config = card.config;

  if (card.card_type === "filter_object_set") {
    return (
      <div className="space-y-2">
        <div>
          <label className="block text-xs font-medium text-neutral-500 mb-1">Object Type</label>
          <input
            type="text"
            value={(config.objectTypeId as string) || ""}
            onChange={(e) => updateConfig({ objectTypeId: e.target.value })}
            className="w-full px-2 py-1 border border-neutral-300 rounded text-xs text-black"
            placeholder="e.g. customers"
          />
        </div>
        <div>
          <label className="block text-xs font-medium text-neutral-500 mb-1">Query</label>
          <input
            type="text"
            value={(config.query as string) || ""}
            onChange={(e) => updateConfig({ query: e.target.value })}
            className="w-full px-2 py-1 border border-neutral-300 rounded text-xs text-black"
            placeholder="Search query"
          />
        </div>
      </div>
    );
  }

  if (VIZ_CARD_TYPES.includes(card.card_type)) {
    return (
      <div className="space-y-2">
        <div>
          <label className="block text-xs font-medium text-neutral-500 mb-1">Group By</label>
          <input
            type="text"
            value={(config.groupBy as string) || ""}
            onChange={(e) => updateConfig({ groupBy: e.target.value })}
            className="w-full px-2 py-1 border border-neutral-300 rounded text-xs text-black"
            placeholder="e.g. region"
          />
        </div>
        <div>
          <label className="block text-xs font-medium text-neutral-500 mb-1">Metric</label>
          <select
            value={(config.metric as string) || "count"}
            onChange={(e) => updateConfig({ metric: e.target.value })}
            className="w-full px-2 py-1 border border-neutral-300 rounded text-xs text-black"
          >
            <option value="count">Count</option>
            <option value="sum">Sum</option>
            <option value="average">Average</option>
          </select>
        </div>
        <div>
          <label className="block text-xs font-medium text-neutral-500 mb-1">Value Field</label>
          <input
            type="text"
            value={(config.valueField as string) || ""}
            onChange={(e) => updateConfig({ valueField: e.target.value })}
            className="w-full px-2 py-1 border border-neutral-300 rounded text-xs text-black"
            placeholder="e.g. amount"
          />
        </div>
      </div>
    );
  }

  if (NUMERIC_CARD_TYPES.includes(card.card_type)) {
    return (
      <div className="space-y-2">
        <div>
          <label className="block text-xs font-medium text-neutral-500 mb-1">Object Type</label>
          <input
            type="text"
            value={(config.objectTypeId as string) || ""}
            onChange={(e) => updateConfig({ objectTypeId: e.target.value })}
            className="w-full px-2 py-1 border border-neutral-300 rounded text-xs text-black"
            placeholder="e.g. orders"
          />
        </div>
        <div>
          <label className="block text-xs font-medium text-neutral-500 mb-1">Field</label>
          <input
            type="text"
            value={(config.field as string) || ""}
            onChange={(e) => updateConfig({ field: e.target.value })}
            className="w-full px-2 py-1 border border-neutral-300 rounded text-xs text-black"
            placeholder="e.g. amount"
          />
        </div>
      </div>
    );
  }

  if (card.card_type === "pivot_table") {
    return (
      <div className="space-y-2">
        <div>
          <label className="block text-xs font-medium text-neutral-500 mb-1">Row Field</label>
          <input
            type="text"
            value={(config.rowField as string) || ""}
            onChange={(e) => updateConfig({ rowField: e.target.value })}
            className="w-full px-2 py-1 border border-neutral-300 rounded text-xs text-black"
            placeholder="e.g. region"
          />
        </div>
        <div>
          <label className="block text-xs font-medium text-neutral-500 mb-1">Column Field</label>
          <input
            type="text"
            value={(config.columnField as string) || ""}
            onChange={(e) => updateConfig({ columnField: e.target.value })}
            className="w-full px-2 py-1 border border-neutral-300 rounded text-xs text-black"
            placeholder="e.g. status"
          />
        </div>
        <div>
          <label className="block text-xs font-medium text-neutral-500 mb-1">Value Field</label>
          <input
            type="text"
            value={(config.valueField as string) || ""}
            onChange={(e) => updateConfig({ valueField: e.target.value })}
            className="w-full px-2 py-1 border border-neutral-300 rounded text-xs text-black"
            placeholder="e.g. amount"
          />
        </div>
      </div>
    );
  }

  if (PARAM_CARD_TYPES.includes(card.card_type)) {
    return (
      <div className="space-y-2">
        <div>
          <label className="block text-xs font-medium text-neutral-500 mb-1">Label</label>
          <input
            type="text"
            value={(config.label as string) || ""}
            onChange={(e) => updateConfig({ label: e.target.value })}
            className="w-full px-2 py-1 border border-neutral-300 rounded text-xs text-black"
            placeholder="Parameter label"
          />
        </div>
        {card.card_type === "param_object_selection" && (
          <div>
            <label className="block text-xs font-medium text-neutral-500 mb-1">Options (comma-separated)</label>
            <input
              type="text"
              value={((config.options as string[]) || []).join(", ")}
              onChange={(e) => updateConfig({ options: e.target.value.split(",").map((s) => s.trim()).filter(Boolean) })}
              className="w-full px-2 py-1 border border-neutral-300 rounded text-xs text-black"
              placeholder="Option A, Option B"
            />
          </div>
        )}
        {(card.card_type === "param_numeric" || card.card_type === "param_string") && (
          <div>
            <label className="block text-xs font-medium text-neutral-500 mb-1">Default Value</label>
            <input
              type="text"
              value={String(config.defaultValue ?? "")}
              onChange={(e) => updateConfig({ defaultValue: card.card_type === "param_numeric" ? parseFloat(e.target.value) || 0 : e.target.value })}
              className="w-full px-2 py-1 border border-neutral-300 rounded text-xs text-black"
            />
          </div>
        )}
      </div>
    );
  }

  return null;
}
