"use client";

import type { CardType, CardResult, ChartDataPoint, TimeSeriesData } from "@/types";
import { VIZ_CARD_TYPES, NUMERIC_CARD_TYPES, PARAM_CARD_TYPES, TIME_SERIES_CARD_TYPES } from "@/types";
import { ChartRenderer } from "./ChartRenderer";
import { NumericDisplay } from "./NumericDisplay";
import { ObjectTableView } from "./ObjectTableView";
import { PivotTableView } from "./PivotTableView";
import { ParameterInput } from "./ParameterInput";
import { TimeSeriesRenderer } from "./TimeSeriesRenderer";
import { OverlayChartRenderer } from "./OverlayChartRenderer";

interface CardResultRendererProps {
  cardType: CardType;
  result?: CardResult;
  config: Record<string, unknown>;
  compact?: boolean;
  onParamChange?: (value: unknown) => void;
  onExecuteAction?: () => void;
  onSaveDataset?: () => void;
}

export function CardResultRenderer({ cardType, result, config, compact, onParamChange, onExecuteAction, onSaveDataset }: CardResultRendererProps) {
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

  // Time series cards
  if (TIME_SERIES_CARD_TYPES.includes(cardType)) {
    const series = (data.series as TimeSeriesData[]) || [];
    const grouped = data.grouped as boolean;
    return <TimeSeriesRenderer series={series} height={compact ? 150 : 250} grouped={grouped} />;
  }

  // Visualization cards
  if (VIZ_CARD_TYPES.includes(cardType)) {
    const chartData = (data.chartData as ChartDataPoint[]) || [];
    return <ChartRenderer cardType={cardType} data={chartData} height={compact ? 150 : 250} />;
  }

  // Overlay chart
  if (cardType === "overlay_chart") {
    const layers = (data.layers as { cardId: string; type: string; chartData?: { name: string; value: number }[]; series?: TimeSeriesData[] }[]) || [];
    return <OverlayChartRenderer layers={layers} height={compact ? 150 : 250} />;
  }

  // Numeric cards
  if (NUMERIC_CARD_TYPES.includes(cardType)) {
    return (
      <NumericDisplay
        value={(data.value as number) ?? 0}
        label={(data.label as string) || ""}
        type={(data.type as string) || cardType}
      />
    );
  }

  // Formula card
  if (cardType === "formula") {
    if (data.mode === "per_row") {
      return (
        <ObjectTableView
          rows={(data.rows as Record<string, unknown>[]) || []}
          columns={(data.columns as string[]) || []}
          totalCount={(data.totalCount as number) || 0}
          compact={compact}
        />
      );
    }
    return (
      <NumericDisplay
        value={(data.value as number) ?? 0}
        label={(data.expression as string) || (data.label as string) || ""}
        type="formula"
      />
    );
  }

  // Object table
  if (cardType === "object_table") {
    return (
      <div>
        <ObjectTableView
          rows={(data.rows as Record<string, unknown>[]) || []}
          columns={(data.columns as string[]) || []}
          totalCount={(data.totalCount as number) || 0}
          compact={compact}
        />
        {!compact && onSaveDataset && (
          <div className="p-2 border-t border-neutral-100">
            <button onClick={onSaveDataset} className="text-xs text-red-600 hover:underline">
              Save as Dataset
            </button>
          </div>
        )}
      </div>
    );
  }

  // Pivot table
  if (cardType === "pivot_table") {
    return (
      <PivotTableView
        pivotData={(data.pivotData as Record<string, Record<string, number>>) || {}}
        rowKeys={(data.rowKeys as string[]) || []}
        columnKeys={(data.columnKeys as string[]) || []}
      />
    );
  }

  // Object set cards
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
    return (
      <div>
        {objects.length > 0 ? (
          <ObjectTableView
            rows={objects}
            columns={Object.keys(objects[0]).filter((k) => k !== "objectTypeId" && k !== "objectTypeID" && k !== "indexedAt")}
            totalCount={count}
            compact
          />
        ) : (
          <div className="text-xs text-neutral-400 p-2">{count} objects</div>
        )}
        {onSaveDataset && (
          <div className="p-2 border-t border-neutral-100">
            <button onClick={onSaveDataset} className="text-xs text-red-600 hover:underline">
              Save as Dataset
            </button>
          </div>
        )}
      </div>
    );
  }

  // Action button
  if (cardType === "action_button") {
    const status = data.status as string;
    return (
      <div className="p-3 space-y-2">
        <div className="flex justify-center">
          <button
            onClick={onExecuteAction}
            disabled={status === "executed"}
            className={`px-4 py-2 text-white text-sm rounded transition-colors ${
              status === "executed"
                ? "bg-green-600"
                : status === "error"
                ? "bg-red-800 hover:bg-red-900"
                : "bg-red-600 hover:bg-red-700"
            }`}
          >
            {status === "executed" ? "Executed" : status === "error" ? "Retry Action" : "Execute Action"}
          </button>
        </div>
        {status === "error" && data.error ? (
          <div className="text-xs text-red-600 text-center">{String(data.error)}</div>
        ) : null}
        {data.actionTypeId ? (
          <div className="text-xs text-neutral-400 text-center">{String(data.actionTypeId)}</div>
        ) : null}
      </div>
    );
  }

  return (
    <pre className="text-[10px] text-neutral-600 p-2 overflow-auto max-h-32 font-mono">
      {JSON.stringify(data, null, 2)}
    </pre>
  );
}
