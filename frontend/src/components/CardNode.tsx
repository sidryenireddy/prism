"use client";

import { memo } from "react";
import { Handle, Position, type NodeProps } from "@xyflow/react";
import type { Card, CardResult } from "@/types";
import { CardResultRenderer } from "./cards/CardResultRenderer";
import { PARAM_CARD_TYPES } from "@/types";

interface CardNodeData {
  card: Card;
  result?: CardResult;
  onDelete: () => void;
  onSelect: () => void;
  onParamChange?: (value: unknown) => void;
}

function CardNodeComponent({ data }: NodeProps) {
  const { card, result, onDelete, onSelect, onParamChange } = data as unknown as CardNodeData;

  const categoryColor = getCategoryColor(card.card_type);
  const isParam = PARAM_CARD_TYPES.includes(card.card_type);
  const hasResult = !!result && !result.error && result.data;

  return (
    <div
      className={`bg-white border border-neutral-200 rounded-lg shadow-sm cursor-pointer hover:border-red-300 transition-colors ${hasResult || isParam ? "min-w-[220px] max-w-[320px]" : "min-w-[180px]"}`}
      onClick={onSelect}
    >
      <Handle type="target" position={Position.Top} className="!bg-red-600 !w-2 !h-2" />
      <div className={`px-3 py-1.5 text-xs font-medium border-b border-neutral-100 rounded-t-lg ${categoryColor}`}>
        {card.card_type.replace(/_/g, " ")}
      </div>
      <div className="px-3 py-1.5">
        <div className="text-sm font-medium text-black truncate">
          {card.label || "Untitled"}
        </div>
      </div>
      {(hasResult || isParam) && (
        <div className="border-t border-neutral-100" onClick={(e) => e.stopPropagation()}>
          <CardResultRenderer
            cardType={card.card_type}
            result={result}
            config={card.config}
            compact
            onParamChange={onParamChange}
          />
        </div>
      )}
      {result?.error && (
        <div className="px-3 py-1 text-xs text-red-600 border-t border-neutral-100 truncate">
          {result.error}
        </div>
      )}
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
  if (["bar_chart", "line_chart", "pie_chart", "scatter_plot", "heat_grid"].includes(cardType))
    return "bg-neutral-100 text-neutral-700";
  if (cardType.includes("table"))
    return "bg-neutral-50 text-neutral-600";
  if (["count", "sum", "average", "min", "max"].includes(cardType))
    return "bg-red-50 text-red-600";
  if (cardType.startsWith("param_"))
    return "bg-neutral-50 text-neutral-500";
  if (cardType === "action_button")
    return "bg-red-100 text-red-700";
  return "bg-neutral-50 text-neutral-600";
}

export const CardNode = memo(CardNodeComponent);
