"use client";

interface PivotTableViewProps {
  pivotData: Record<string, Record<string, number>>;
  rowKeys: string[];
  columnKeys: string[];
}

export function PivotTableView({ pivotData, rowKeys, columnKeys }: PivotTableViewProps) {
  if (!rowKeys?.length || !columnKeys?.length) {
    return <div className="text-xs text-neutral-400 p-2">Configure row and column fields</div>;
  }

  return (
    <div className="overflow-auto">
      <table className="w-full border-collapse text-xs">
        <thead>
          <tr>
            <th className="px-2 py-1 text-left font-medium text-neutral-500 border-b border-neutral-200" />
            {columnKeys.map((ck) => (
              <th key={ck} className="px-2 py-1 text-right font-medium text-neutral-500 border-b border-neutral-200 whitespace-nowrap">
                {ck}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rowKeys.map((rk) => (
            <tr key={rk} className="hover:bg-neutral-50">
              <td className="px-2 py-1 font-medium text-black border-b border-neutral-50 whitespace-nowrap">{rk}</td>
              {columnKeys.map((ck) => (
                <td key={ck} className="px-2 py-1 text-right text-black border-b border-neutral-50 tabular-nums">
                  {pivotData[rk]?.[ck]?.toLocaleString(undefined, { maximumFractionDigits: 1 }) ?? "—"}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
