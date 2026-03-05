"use client";

import { useState } from "react";
import type { CardType } from "@/types";
import { CARD_CATEGORIES } from "@/types";

interface CardPaletteProps {
  onAddCard: (cardType: CardType, x: number, y: number) => void;
}

export function CardPalette({ onAddCard }: CardPaletteProps) {
  const [expanded, setExpanded] = useState<string | null>(null);

  return (
    <div className="w-52 border-r border-neutral-200 bg-neutral-50 overflow-y-auto shrink-0">
      <div className="p-3 border-b border-neutral-200">
        <h3 className="text-xs font-semibold text-neutral-500 uppercase tracking-wider">
          Card Types
        </h3>
      </div>
      {Object.entries(CARD_CATEGORIES).map(([category, types]) => (
        <div key={category} className="border-b border-neutral-100">
          <button
            onClick={() => setExpanded(expanded === category ? null : category)}
            className="w-full px-3 py-2 text-left text-sm font-medium text-black hover:bg-neutral-100 transition-colors flex justify-between items-center"
          >
            {category}
            <span className="text-neutral-400 text-xs">
              {expanded === category ? "-" : "+"}
            </span>
          </button>
          {expanded === category && (
            <div className="pb-1">
              {types.map((item) => (
                <button
                  key={item.type}
                  onClick={() => {
                    const x = 100 + Math.random() * 400;
                    const y = 100 + Math.random() * 300;
                    onAddCard(item.type, x, y);
                  }}
                  className="w-full px-4 py-1.5 text-left text-xs text-neutral-600 hover:bg-neutral-200 hover:text-black transition-colors"
                >
                  {item.label}
                </button>
              ))}
            </div>
          )}
        </div>
      ))}
    </div>
  );
}
