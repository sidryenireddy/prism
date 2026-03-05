"use client";

import { useState } from "react";

interface ObjectTableViewProps {
  rows: Record<string, unknown>[];
  columns: string[];
  totalCount: number;
  compact?: boolean;
}

export function ObjectTableView({ rows, columns, totalCount, compact }: ObjectTableViewProps) {
  const [sortCol, setSortCol] = useState<string | null>(null);
  const [sortDir, setSortDir] = useState<"asc" | "desc">("asc");
  const [filterCol, setFilterCol] = useState("");
  const [filterVal, setFilterVal] = useState("");

  let displayed = [...rows];

  if (filterCol && filterVal) {
    displayed = displayed.filter((r) =>
      String(r[filterCol] ?? "").toLowerCase().includes(filterVal.toLowerCase())
    );
  }

  if (sortCol) {
    displayed.sort((a, b) => {
      const av = String(a[sortCol] ?? "");
      const bv = String(b[sortCol] ?? "");
      const cmp = av.localeCompare(bv, undefined, { numeric: true });
      return sortDir === "asc" ? cmp : -cmp;
    });
  }

  const handleSort = (col: string) => {
    if (sortCol === col) {
      setSortDir(sortDir === "asc" ? "desc" : "asc");
    } else {
      setSortCol(col);
      setSortDir("asc");
    }
  };

  const visibleCols = columns.length > 0 ? columns : Object.keys(rows[0] || {});
  const fontSize = compact ? "text-[10px]" : "text-xs";

  return (
    <div className="overflow-auto">
      {!compact && (
        <div className="flex gap-2 p-2 border-b border-neutral-100">
          <select
            value={filterCol}
            onChange={(e) => setFilterCol(e.target.value)}
            className="text-xs border border-neutral-200 rounded px-1 py-0.5 text-black"
          >
            <option value="">Filter column...</option>
            {visibleCols.map((c) => (
              <option key={c} value={c}>{c}</option>
            ))}
          </select>
          {filterCol && (
            <input
              type="text"
              value={filterVal}
              onChange={(e) => setFilterVal(e.target.value)}
              placeholder="Filter value..."
              className="text-xs border border-neutral-200 rounded px-1 py-0.5 text-black flex-1"
            />
          )}
          <span className="text-xs text-neutral-400 self-center">{totalCount} rows</span>
        </div>
      )}
      <table className="w-full border-collapse">
        <thead>
          <tr>
            {visibleCols.map((col) => (
              <th
                key={col}
                onClick={() => handleSort(col)}
                className={`${fontSize} font-medium text-neutral-500 text-left px-2 py-1 border-b border-neutral-100 cursor-pointer hover:text-black whitespace-nowrap`}
              >
                {col}
                {sortCol === col && (sortDir === "asc" ? " ^" : " v")}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {displayed.slice(0, compact ? 5 : 50).map((row, i) => (
            <tr key={i} className="hover:bg-neutral-50">
              {visibleCols.map((col) => (
                <td key={col} className={`${fontSize} text-black px-2 py-1 border-b border-neutral-50 whitespace-nowrap`}>
                  {formatCellValue(row[col])}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function formatCellValue(v: unknown): string {
  if (v === null || v === undefined) return "";
  if (typeof v === "number") return v.toLocaleString(undefined, { maximumFractionDigits: 2 });
  return String(v);
}
