"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";
import type { Dashboard, Card, CardResult } from "@/types";
import { PARAM_CARD_TYPES } from "@/types";
import { CardResultRenderer } from "@/components/cards/CardResultRenderer";

export default function DashboardViewerPage() {
  const params = useParams();
  const dashboardId = params.id as string;

  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [cards, setCards] = useState<Card[]>([]);
  const [results, setResults] = useState<Record<string, CardResult>>({});
  const [loading, setLoading] = useState(true);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(0); // seconds, 0 = off
  const dashRef = useRef<HTMLDivElement>(null);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const executeAndSetResults = useCallback(async (analysisId: string) => {
    try {
      const response = await api.executeAnalysis(analysisId);
      const parsed: Record<string, CardResult> = {};
      for (const [id, raw] of Object.entries(response.results)) {
        parsed[id] = typeof raw === "string" ? JSON.parse(raw) : raw;
      }
      setResults(parsed);
    } catch {
      // execution might fail
    }
  }, []);

  useEffect(() => {
    api
      .getDashboard(dashboardId)
      .then(async (d) => {
        setDashboard(d);
        const allCards = await api.listCards(d.analysis_id);
        const dashCardIds = new Set(d.layout.cards.map((p) => p.card_id));
        setCards(allCards.filter((c) => dashCardIds.has(c.id)));
        await executeAndSetResults(d.analysis_id);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [dashboardId, executeAndSetResults]);

  // Auto-refresh
  useEffect(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
    if (autoRefresh > 0 && dashboard) {
      intervalRef.current = setInterval(() => {
        executeAndSetResults(dashboard.analysis_id);
      }, autoRefresh * 1000);
    }
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [autoRefresh, dashboard, executeAndSetResults]);

  const handleParamChange = useCallback(
    async (cardId: string, value: unknown) => {
      if (!dashboard) return;
      const card = cards.find((c) => c.id === cardId);
      if (!card) return;
      await api.updateCard(dashboard.analysis_id, cardId, {
        config: { ...card.config, value },
      });
      setCards((prev) =>
        prev.map((c) =>
          c.id === cardId ? { ...c, config: { ...c.config, value } } : c
        )
      );
      await executeAndSetResults(dashboard.analysis_id);
    },
    [dashboard, cards, executeAndSetResults]
  );

  const toggleFullscreen = () => {
    if (!isFullscreen) {
      dashRef.current?.requestFullscreen?.();
    } else {
      document.exitFullscreen?.();
    }
    setIsFullscreen(!isFullscreen);
  };

  const handleExport = async () => {
    if (!dashRef.current) return;
    // Use browser print as PDF export
    window.print();
  };

  if (loading) {
    return (
      <div className="p-8">
        <p className="text-neutral-500 text-sm">Loading...</p>
      </div>
    );
  }

  if (!dashboard) {
    return (
      <div className="p-8">
        <p className="text-neutral-500 text-sm">Dashboard not found.</p>
        <Link href="/dashboard" className="text-red-600 text-sm hover:underline mt-2 inline-block">
          Back to Dashboards
        </Link>
      </div>
    );
  }

  const placementMap = new Map(
    dashboard.layout.cards.map((p) => [p.card_id, p])
  );

  const gridCols = 12;
  const cellSize = 80;

  return (
    <div ref={dashRef} className="flex flex-col h-full bg-white">
      <div className="px-4 py-3 border-b border-neutral-200 flex items-center justify-between shrink-0 print:hidden">
        <div className="flex items-center gap-3">
          <Link
            href="/dashboard"
            className="text-neutral-400 hover:text-black text-sm transition-colors"
          >
            &lt; Back
          </Link>
          <h1 className="text-lg font-semibold text-black">{dashboard.name}</h1>
          {dashboard.published && (
            <span className="text-xs px-2 py-0.5 rounded bg-green-100 text-green-700">Published</span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <select
            value={autoRefresh}
            onChange={(e) => setAutoRefresh(parseInt(e.target.value))}
            className="text-xs border border-neutral-300 rounded px-2 py-1 text-black"
          >
            <option value={0}>Auto-refresh: Off</option>
            <option value={10}>10s</option>
            <option value={30}>30s</option>
            <option value={60}>1m</option>
            <option value={300}>5m</option>
          </select>
          <button
            onClick={toggleFullscreen}
            className="px-3 py-1.5 border border-neutral-300 text-sm text-black rounded hover:bg-neutral-50"
          >
            {isFullscreen ? "Exit Fullscreen" : "Fullscreen"}
          </button>
          <button
            onClick={handleExport}
            className="px-3 py-1.5 border border-neutral-300 text-sm text-black rounded hover:bg-neutral-50"
          >
            Export PDF
          </button>
          <button
            onClick={() => executeAndSetResults(dashboard.analysis_id)}
            className="px-3 py-1.5 bg-red-600 text-white text-sm rounded hover:bg-red-700"
          >
            Refresh
          </button>
        </div>
      </div>
      <div className="flex-1 overflow-auto p-6">
        <div
          className="relative mx-auto"
          style={{
            width: gridCols * cellSize,
            minHeight: 600,
          }}
        >
          {cards.map((card) => {
            const placement = placementMap.get(card.id);
            if (!placement) return null;
            const isParam = PARAM_CARD_TYPES.includes(card.card_type);
            return (
              <div
                key={card.id}
                className="absolute border border-neutral-200 rounded-lg bg-white shadow-sm overflow-hidden"
                style={{
                  left: placement.x * cellSize,
                  top: placement.y * cellSize,
                  width: placement.w * cellSize,
                  height: placement.h * cellSize,
                }}
              >
                <div className="px-3 py-1.5 border-b border-neutral-100 bg-neutral-50">
                  <div className="text-xs font-medium text-neutral-700 truncate">
                    {card.label || card.card_type.replace(/_/g, " ")}
                  </div>
                </div>
                <div className="overflow-auto" style={{ height: `calc(100% - 30px)` }}>
                  <CardResultRenderer
                    cardType={card.card_type}
                    result={results[card.id]}
                    config={card.config}
                    onParamChange={isParam ? (v) => handleParamChange(card.id, v) : undefined}
                  />
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
