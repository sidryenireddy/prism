"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";
import { AnalysisCanvas } from "@/components/AnalysisCanvas";
import type { Analysis, Card, CardResult, CardType } from "@/types";

export default function AnalysisPage() {
  const params = useParams();
  const analysisId = params.id as string;

  const [analysis, setAnalysis] = useState<Analysis | null>(null);
  const [cards, setCards] = useState<Card[]>([]);
  const [results, setResults] = useState<Record<string, CardResult>>({});
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      api.getAnalysis(analysisId),
      api.listCards(analysisId),
    ])
      .then(([a, c]) => {
        setAnalysis(a);
        setCards(c);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [analysisId]);

  const handleAddCard = useCallback(
    async (cardType: CardType, x: number, y: number) => {
      const card = await api.createCard(analysisId, {
        card_type: cardType,
        label: cardType.replace(/_/g, " "),
        config: {},
        position_x: x,
        position_y: y,
        input_card_ids: [],
      });
      setCards((prev) => [...prev, card]);
    },
    [analysisId]
  );

  const handleUpdateCard = useCallback(
    async (cardId: string, data: Partial<Card>) => {
      const updated = await api.updateCard(analysisId, cardId, data);
      setCards((prev) => prev.map((c) => (c.id === cardId ? updated : c)));
    },
    [analysisId]
  );

  const handleDeleteCard = useCallback(
    async (cardId: string) => {
      await api.deleteCard(analysisId, cardId);
      setCards((prev) => prev.filter((c) => c.id !== cardId));
    },
    [analysisId]
  );

  const handleConnect = useCallback(
    async (sourceId: string, targetId: string) => {
      const targetCard = cards.find((c) => c.id === targetId);
      if (!targetCard) return;

      const newInputIds = [...targetCard.input_card_ids, sourceId];
      await api.updateCard(analysisId, targetId, {
        input_card_ids: newInputIds,
      });
      setCards((prev) =>
        prev.map((c) =>
          c.id === targetId ? { ...c, input_card_ids: newInputIds } : c
        )
      );
    },
    [analysisId, cards]
  );

  const handleExecute = useCallback(async () => {
    const response = await api.executeAnalysis(analysisId);
    const parsed: Record<string, CardResult> = {};
    for (const [id, raw] of Object.entries(response.results)) {
      parsed[id] = typeof raw === "string" ? JSON.parse(raw) : raw;
    }
    setResults(parsed);
  }, [analysisId]);

  if (loading) {
    return (
      <div className="p-8">
        <p className="text-neutral-500 text-sm">Loading...</p>
      </div>
    );
  }

  if (!analysis) {
    return (
      <div className="p-8">
        <p className="text-neutral-500 text-sm">Analysis not found.</p>
        <Link href="/" className="text-red-600 text-sm hover:underline mt-2 inline-block">
          Back to Analyses
        </Link>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      <div className="px-4 py-3 border-b border-neutral-200 flex items-center justify-between shrink-0">
        <div className="flex items-center gap-3">
          <Link
            href="/"
            className="text-neutral-400 hover:text-black text-sm transition-colors"
          >
            &lt; Back
          </Link>
          <h1 className="text-lg font-semibold text-black">{analysis.name}</h1>
          {analysis.description && (
            <span className="text-sm text-neutral-500">
              {analysis.description}
            </span>
          )}
        </div>
      </div>
      <div className="flex-1">
        <AnalysisCanvas
          analysisId={analysisId}
          cards={cards}
          results={results}
          onAddCard={handleAddCard}
          onUpdateCard={handleUpdateCard}
          onDeleteCard={handleDeleteCard}
          onConnect={handleConnect}
          onExecute={handleExecute}
        />
      </div>
    </div>
  );
}
