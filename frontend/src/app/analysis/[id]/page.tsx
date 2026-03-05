"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";
import { AnalysisCanvas } from "@/components/AnalysisCanvas";
import type { Analysis, Card, CardResult, CardType } from "@/types";

export default function AnalysisPage() {
  const params = useParams();
  const router = useRouter();
  const analysisId = params.id as string;

  const [analysis, setAnalysis] = useState<Analysis | null>(null);
  const [cards, setCards] = useState<Card[]>([]);
  const [results, setResults] = useState<Record<string, CardResult>>({});
  const [loading, setLoading] = useState(true);
  const [showDashboardModal, setShowDashboardModal] = useState(false);
  const [dashboardName, setDashboardName] = useState("");
  const [selectedCardIds, setSelectedCardIds] = useState<Set<string>>(new Set());

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

  const handleAiGenerate = useCallback(
    async (prompt: string) => {
      const response = await api.aiGenerate({
        analysis_id: analysisId,
        prompt,
        cards,
      });
      if (response.cards) {
        setCards((prev) => [...prev, ...response.cards]);
      }
    },
    [analysisId, cards]
  );

  const handleCreateDashboard = async () => {
    if (!dashboardName.trim() || selectedCardIds.size === 0) return;
    const placements = Array.from(selectedCardIds).map((cardId, i) => ({
      card_id: cardId,
      x: (i % 3) * 4,
      y: Math.floor(i / 3) * 4,
      w: 4,
      h: 4,
    }));
    const dashboard = await api.createDashboard({
      analysis_id: analysisId,
      name: dashboardName,
      published: true,
      layout: { cards: placements },
    });
    setShowDashboardModal(false);
    router.push(`/dashboard/${dashboard.id}`);
  };

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
        <button
          onClick={() => {
            setDashboardName(analysis.name + " Dashboard");
            setSelectedCardIds(new Set(cards.map((c) => c.id)));
            setShowDashboardModal(true);
          }}
          className="px-3 py-1.5 border border-neutral-300 text-sm text-black rounded hover:bg-neutral-50 transition-colors"
        >
          Create Dashboard
        </button>
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
          onAiGenerate={handleAiGenerate}
        />
      </div>

      {/* Dashboard creation modal */}
      {showDashboardModal && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl w-[480px] max-h-[80vh] overflow-y-auto">
            <div className="p-4 border-b border-neutral-200 flex justify-between items-center">
              <h2 className="text-lg font-semibold text-black">Create Dashboard</h2>
              <button onClick={() => setShowDashboardModal(false)} className="text-neutral-400 hover:text-black">x</button>
            </div>
            <div className="p-4 space-y-4">
              <div>
                <label className="block text-sm font-medium text-neutral-700 mb-1">Name</label>
                <input
                  type="text"
                  value={dashboardName}
                  onChange={(e) => setDashboardName(e.target.value)}
                  className="w-full px-3 py-2 border border-neutral-300 rounded text-sm text-black"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-neutral-700 mb-2">Select Cards</label>
                <div className="space-y-1 max-h-60 overflow-y-auto">
                  {cards.map((card) => (
                    <label key={card.id} className="flex items-center gap-2 py-1 px-2 rounded hover:bg-neutral-50 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={selectedCardIds.has(card.id)}
                        onChange={(e) => {
                          const next = new Set(selectedCardIds);
                          if (e.target.checked) next.add(card.id);
                          else next.delete(card.id);
                          setSelectedCardIds(next);
                        }}
                        className="accent-red-600"
                      />
                      <span className="text-sm text-black">{card.label || card.card_type.replace(/_/g, " ")}</span>
                      <span className="text-xs text-neutral-400 ml-auto">{card.card_type.replace(/_/g, " ")}</span>
                    </label>
                  ))}
                </div>
              </div>
            </div>
            <div className="p-4 border-t border-neutral-200 flex justify-end gap-2">
              <button
                onClick={() => setShowDashboardModal(false)}
                className="px-4 py-2 text-sm text-neutral-600 hover:text-black"
              >
                Cancel
              </button>
              <button
                onClick={handleCreateDashboard}
                className="px-4 py-2 bg-red-600 text-white text-sm rounded hover:bg-red-700"
              >
                Publish Dashboard
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
