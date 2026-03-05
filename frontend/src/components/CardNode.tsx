"use client";

import { memo } from "react";
import { Handle, Position, type NodeProps } from "@xyflow/react";
import type { Card, CardResult } from "@/types";

interface CardNodeData {
  card: Card;
  result?: CardResult;
  onDelete: () => void;
  onSelect: () => void;
}

function CardNodeComponent({ data }: NodeProps) {
  const { card, result, onDelete, onSelect } = data as unknown as CardNodeData;

  const categoryColor = getCategoryColor(card.card_type);

  return (
    <div
      className="bg-white border border-neutral-200 rounded-lg shadow-sm min-w-[180px] cursor-pointer hover:border-red-300 transition-colors"
      onClick={onSelect}
    >
      <Handle type="target" position={Position.Top} className="!bg-red-600 !w-2 !h-2" />
      <div className={`px-3 py-1.5 text-xs font-medium border-b border-neutral-100 rounded-t-lg ${categoryColor}`}>
        {card.card_type.replace(/_/g, " ")}
      </div>
      <div className="px-3 py-2">
        <div className="text-sm font-medium text-black truncate">
          {card.label || "Untitled"}
        </div>
        {result && (
          <div className="mt-1 text-xs text-neutral-500">
            {result.error ? (
              <span className="text-red-600">{result.error}</span>
            ) : (
              <span className="text-green-700">Result ready</span>
            )}
          </div>
        )}
      </div>
      <div className="px-3 py-1 border-t border-neutral-100 flex justify-end">
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete();
          }}
          className="text-xs text-neutral-400 hover:text-red-600 transition-colors"
        >
          Remove
        </button>
      </div>
      <Handle type="source" position={Position.Bottom} className="!bg-red-600 !w-2 !h-2" />
    </div>
  );
}

function getCategoryColor(cardType: string): string {
  if (cardType.includes("object") || cardType.includes("search") || cardType.includes("set_math"))
    return "bg-red-50 text-red-700";
  if (["bar_chart", "line_chart", "pie_chart", "scatter_plot", "heat_grid", "map"].includes(cardType))
    return "bg-neutral-100 text-neutral-700";
  if (cardType.includes("table"))
    return "bg-neutral-50 text-neutral-600";
  if (["count", "sum", "average", "min", "max"].includes(cardType))
    return "bg-red-50 text-red-600";
  if (cardType.includes("time_series") || cardType.includes("rolling") || cardType.includes("formula"))
    return "bg-neutral-100 text-neutral-600";
  if (cardType.startsWith("param_"))
    return "bg-neutral-50 text-neutral-500";
  if (cardType === "action_button")
    return "bg-red-100 text-red-700";
  return "bg-neutral-50 text-neutral-600";
}

export const CardNode = memo(CardNodeComponent);
