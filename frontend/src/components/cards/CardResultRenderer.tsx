"use client";

import type { CardType, CardResult, ChartDataPoint } from "@/types";
import { VIZ_CARD_TYPES, NUMERIC_CARD_TYPES, TABLE_CARD_TYPES, PARAM_CARD_TYPES } from "@/types";
import { ChartRenderer } from "./ChartRenderer";
import { NumericDisplay } from "./NumericDisplay";
import { ObjectTableView } from "./ObjectTableView";
import { PivotTableView } from "./PivotTableView";
import { ParameterInput } from "./ParameterInput";

interface CardResultRendererProps {
  cardType: CardType;
  result?: CardResult;
  config: Record<string, unknown>;
  compact?: boolean;
  onParamChange?: (value: unknown) => void;
}

export function CardResultRenderer({ cardType, result, config, compact, onParamChange }: CardResultRendererProps) {
  if (PARAM_CARD_TYPES.includes(cardType)) {
    const data = result?.data as Record<string, unknown> | undefined;
    return (
      <ParameterInput
        cardType={cardType}
        config={config}
        value={data?.value}
        onChange={onParamChange || (() => {})}
      />
    );
  }

  if (!result) {
    return <div className="text-xs text-neutral-400 p-2">Not executed</div>;
  }

  if (result.error) {
    return <div className="text-xs text-red-600 p-2">{result.error}</div>;
  }

  const data = result.data as Record<string, unknown>;
  if (!data) return <div className="text-xs text-neutral-400 p-2">No data</div>;

  if (VIZ_CARD_TYPES.includes(cardType)) {
    const chartData = (data.chartData as ChartDataPoint[]) || [];
    return <ChartRenderer cardType={cardType} data={chartData} height={compact ? 150 : 250} />;
  }

  if (NUMERIC_CARD_TYPES.includes(cardType)) {
    return (
      <NumericDisplay
        value={(data.value as number) ?? 0}
        label={(data.label as string) || ""}
        type={(data.type as string) || cardType}
      />
    );
  }

  if (cardType === "object_table") {
    return (
      <ObjectTableView
        rows={(data.rows as Record<string, unknown>[]) || []}
        columns={(data.columns as string[]) || []}
        totalCount={(data.totalCount as number) || 0}
        compact={compact}
      />
    );
  }

  if (cardType === "pivot_table") {
    return (
      <PivotTableView
        pivotData={(data.pivotData as Record<string, Record<string, number>>) || {}}
        rowKeys={(data.rowKeys as string[]) || []}
        columnKeys={(data.columnKeys as string[]) || []}
      />
    );
  }

  if (cardType === "filter_object_set" || cardType === "search_around" ||
      cardType === "set_math_union" || cardType === "set_math_intersection" || cardType === "set_math_difference") {
    const objects = (data.objects as Record<string, unknown>[]) || [];
    const count = (data.totalCount as number) || objects.length;
    if (compact) {
      return (
        <div className="p-2">
          <div className="text-lg font-bold text-black">{count}</div>
          <div className="text-xs text-neutral-500">objects</div>
        </div>
      );
    }
    if (objects.length > 0) {
      const cols = Object.keys(objects[0]).filter((k) => k !== "objectTypeId" && k !== "objectTypeID" && k !== "indexedAt");
      return <ObjectTableView rows={objects} columns={cols} totalCount={count} compact />;
    }
    return <div className="text-xs text-neutral-400 p-2">{count} objects</div>;
  }

  if (cardType === "action_button") {
    return (
      <div className="p-3 flex justify-center">
        <button className="px-4 py-2 bg-red-600 text-white text-sm rounded hover:bg-red-700">
          Execute Action
        </button>
      </div>
    );
  }

  return (
    <pre className="text-[10px] text-neutral-600 p-2 overflow-auto max-h-32 font-mono">
      {JSON.stringify(data, null, 2)}
    </pre>
  );
}
