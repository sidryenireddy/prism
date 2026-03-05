"use client";

import type { Card, CardResult } from "@/types";

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
  return (
    <div className="w-72 border-l border-neutral-200 bg-white overflow-y-auto shrink-0">
      <div className="p-3 border-b border-neutral-200 flex items-center justify-between">
        <h3 className="text-sm font-semibold text-black">Card Configuration</h3>
        <button
          onClick={onClose}
          className="text-neutral-400 hover:text-black text-sm"
        >
          x
        </button>
      </div>

      <div className="p-3 space-y-3">
        <div>
          <label className="block text-xs font-medium text-neutral-500 mb-1">
            Type
          </label>
          <div className="text-sm text-black">
            {card.card_type.replace(/_/g, " ")}
          </div>
        </div>

        <div>
          <label className="block text-xs font-medium text-neutral-500 mb-1">
            Label
          </label>
          <input
            type="text"
            value={card.label}
            onChange={(e) => onUpdate({ label: e.target.value })}
            className="w-full px-2 py-1.5 border border-neutral-300 rounded text-sm text-black"
            placeholder="Card label"
          />
        </div>

        <div>
          <label className="block text-xs font-medium text-neutral-500 mb-1">
            Configuration (JSON)
          </label>
          <textarea
            value={JSON.stringify(card.config, null, 2)}
            onChange={(e) => {
              try {
                const config = JSON.parse(e.target.value);
                onUpdate({ config });
              } catch {
                // invalid JSON, don't update
              }
            }}
            className="w-full px-2 py-1.5 border border-neutral-300 rounded text-xs text-black font-mono h-32 resize-y"
          />
        </div>

        {result && (
          <div>
            <label className="block text-xs font-medium text-neutral-500 mb-1">
              Result
            </label>
            {result.error ? (
              <div className="text-xs text-red-600 p-2 bg-red-50 rounded">
                {result.error}
              </div>
            ) : (
              <pre className="text-xs text-black p-2 bg-neutral-50 rounded overflow-auto max-h-48 font-mono">
                {JSON.stringify(result.data, null, 2)}
              </pre>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
