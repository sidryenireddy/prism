"use client";

import { useCallback, useMemo, useState } from "react";
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  addEdge,
  useNodesState,
  useEdgesState,
  type Connection,
  type Node,
  type Edge,
  type NodeTypes,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";

import type { Card, CardType, CardResult } from "@/types";
import { CardNode } from "./CardNode";
import { CardPalette } from "./CardPalette";
import { CardConfigPanel } from "./CardConfigPanel";

interface AnalysisCanvasProps {
  analysisId: string;
  cards: Card[];
  results: Record<string, CardResult>;
  onAddCard: (cardType: CardType, x: number, y: number) => void;
  onUpdateCard: (cardId: string, data: Partial<Card>) => void;
  onDeleteCard: (cardId: string) => void;
  onConnect: (sourceId: string, targetId: string) => void;
  onExecute: () => void;
  onAiGenerate?: (prompt: string) => Promise<void>;
}

export function AnalysisCanvas({
  cards,
  results,
  onAddCard,
  onUpdateCard,
  onDeleteCard,
  onConnect,
  onExecute,
  onAiGenerate,
}: AnalysisCanvasProps) {
  const [selectedCard, setSelectedCard] = useState<Card | null>(null);
  const [graphMode, setGraphMode] = useState(false);
  const [aiPrompt, setAiPrompt] = useState("");
  const [aiLoading, setAiLoading] = useState(false);

  // Auto-layout positions for graph mode
  const graphPositions = useMemo(() => {
    if (!graphMode) return null;
    const positions: Record<string, { x: number; y: number }> = {};
    const adjList: Record<string, string[]> = {};
    const inDegree: Record<string, number> = {};

    cards.forEach((c) => {
      adjList[c.id] = [];
      inDegree[c.id] = 0;
    });
    cards.forEach((c) => {
      c.input_card_ids.forEach((inputId) => {
        if (adjList[inputId]) {
          adjList[inputId].push(c.id);
        }
        inDegree[c.id] = (inDegree[c.id] || 0) + 1;
      });
    });

    // Topological layers
    const layers: string[][] = [];
    const visited = new Set<string>();
    let queue = cards.filter((c) => inDegree[c.id] === 0).map((c) => c.id);

    while (queue.length > 0) {
      layers.push([...queue]);
      queue.forEach((id) => visited.add(id));
      const next: string[] = [];
      queue.forEach((id) => {
        (adjList[id] || []).forEach((child) => {
          inDegree[child]--;
          if (inDegree[child] === 0 && !visited.has(child)) {
            next.push(child);
          }
        });
      });
      queue = next;
    }

    // Position: center each layer horizontally
    layers.forEach((layer, layerIdx) => {
      const totalWidth = layer.length * 260;
      const startX = -totalWidth / 2 + 130;
      layer.forEach((id, i) => {
        positions[id] = { x: startX + i * 260, y: layerIdx * 180 };
      });
    });

    return positions;
  }, [graphMode, cards]);

  const nodes: Node[] = useMemo(
    () =>
      cards.map((card) => ({
        id: card.id,
        type: "cardNode",
        position: graphPositions
          ? graphPositions[card.id] || { x: 0, y: 0 }
          : { x: card.position_x, y: card.position_y },
        data: {
          card,
          result: results[card.id],
          onDelete: () => onDeleteCard(card.id),
          onSelect: () => setSelectedCard(card),
          onParamChange: (value: unknown) => {
            onUpdateCard(card.id, {
              config: { ...card.config, value },
            });
          },
        },
      })),
    [cards, results, onDeleteCard, onUpdateCard, graphPositions]
  );

  const edges: Edge[] = useMemo(
    () =>
      cards.flatMap((card) =>
        card.input_card_ids.map((inputId) => ({
          id: `${inputId}-${card.id}`,
          source: inputId,
          target: card.id,
          animated: true,
          style: { stroke: "#dc2626" },
        }))
      ),
    [cards]
  );

  const [flowNodes, setNodes, onNodesChange] = useNodesState(nodes);
  const [flowEdges, setEdges, onEdgesChange] = useEdgesState(edges);

  useMemo(() => {
    setNodes(nodes);
    setEdges(edges);
  }, [nodes, edges, setNodes, setEdges]);

  const handleConnect = useCallback(
    (connection: Connection) => {
      if (connection.source && connection.target) {
        onConnect(connection.source, connection.target);
        setEdges((eds) => addEdge({ ...connection, animated: true, style: { stroke: "#dc2626" } }, eds));
      }
    },
    [onConnect, setEdges]
  );

  const handleNodeDragStop = useCallback(
    (_: React.MouseEvent, node: Node) => {
      onUpdateCard(node.id, {
        position_x: node.position.x,
        position_y: node.position.y,
      });
    },
    [onUpdateCard]
  );

  const handleAiGenerate = async () => {
    if (!aiPrompt.trim() || !onAiGenerate) return;
    setAiLoading(true);
    try {
      await onAiGenerate(aiPrompt);
      setAiPrompt("");
    } finally {
      setAiLoading(false);
    }
  };

  const nodeTypes: NodeTypes = useMemo(() => ({ cardNode: CardNode }), []);

  return (
    <div className="flex h-full">
      <CardPalette onAddCard={onAddCard} />
      <div className="flex-1 relative">
        {/* Top bar */}
        <div className="absolute top-3 right-3 z-10 flex gap-2">
          <button
            onClick={() => setGraphMode(!graphMode)}
            className={`px-3 py-2 text-sm rounded transition-colors ${
              graphMode
                ? "bg-neutral-800 text-white"
                : "bg-white text-black border border-neutral-300 hover:bg-neutral-50"
            }`}
          >
            {graphMode ? "Canvas" : "Graph"}
          </button>
          <button
            onClick={onExecute}
            className="px-4 py-2 bg-red-600 text-white text-sm rounded hover:bg-red-700 transition-colors"
          >
            Execute
          </button>
        </div>
        <ReactFlow
          nodes={flowNodes}
          edges={flowEdges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={handleConnect}
          onNodeDragStop={handleNodeDragStop}
          nodeTypes={nodeTypes}
          fitView
          className="bg-white"
          nodesDraggable={!graphMode}
        >
          <Background color="#e5e5e5" gap={20} />
          <Controls />
          <MiniMap
            nodeColor="#dc2626"
            maskColor="rgba(0,0,0,0.1)"
          />
        </ReactFlow>

        {/* AI Generate bar */}
        <div className="absolute bottom-4 left-1/2 -translate-x-1/2 z-10 w-[500px] max-w-[90%]">
          <div className="flex bg-white border border-neutral-300 rounded-lg shadow-lg overflow-hidden">
            <input
              type="text"
              value={aiPrompt}
              onChange={(e) => setAiPrompt(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleAiGenerate()}
              placeholder="Describe an analysis to generate cards..."
              className="flex-1 px-4 py-2.5 text-sm text-black outline-none"
            />
            <button
              onClick={handleAiGenerate}
              disabled={aiLoading || !onAiGenerate}
              className="px-4 py-2.5 bg-red-600 text-white text-sm font-medium hover:bg-red-700 disabled:opacity-50 transition-colors"
            >
              {aiLoading ? "Generating..." : "Generate"}
            </button>
          </div>
        </div>
      </div>
      {selectedCard && (
        <CardConfigPanel
          card={selectedCard}
          result={results[selectedCard.id]}
          onUpdate={(data) => {
            onUpdateCard(selectedCard.id, data);
          }}
          onClose={() => setSelectedCard(null)}
        />
      )}
    </div>
  );
}
