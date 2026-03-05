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
import { CARD_CATEGORIES } from "@/types";
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
}

export function AnalysisCanvas({
  cards,
  results,
  onAddCard,
  onUpdateCard,
  onDeleteCard,
  onConnect,
  onExecute,
}: AnalysisCanvasProps) {
  const [selectedCard, setSelectedCard] = useState<Card | null>(null);

  const nodes: Node[] = useMemo(
    () =>
      cards.map((card) => ({
        id: card.id,
        type: "cardNode",
        position: { x: card.position_x, y: card.position_y },
        data: {
          card,
          result: results[card.id],
          onDelete: () => onDeleteCard(card.id),
          onSelect: () => setSelectedCard(card),
        },
      })),
    [cards, results, onDeleteCard]
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

  // Sync when cards/results change
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

  const nodeTypes: NodeTypes = useMemo(() => ({ cardNode: CardNode }), []);

  return (
    <div className="flex h-full">
      <CardPalette onAddCard={onAddCard} />
      <div className="flex-1 relative">
        <div className="absolute top-3 right-3 z-10">
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
        >
          <Background color="#e5e5e5" gap={20} />
          <Controls />
          <MiniMap
            nodeColor="#dc2626"
            maskColor="rgba(0,0,0,0.1)"
          />
        </ReactFlow>
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
