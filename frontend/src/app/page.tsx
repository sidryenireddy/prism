"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api } from "@/lib/api";
import type { Analysis } from "@/types";

export default function HomePage() {
  const [analyses, setAnalyses] = useState<Analysis[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");

  useEffect(() => {
    api
      .listAnalyses()
      .then(setAnalyses)
      .catch(() => setAnalyses([]))
      .finally(() => setLoading(false));
  }, []);

  const handleCreate = async () => {
    if (!name.trim()) return;
    const analysis = await api.createAnalysis({
      name,
      description,
      owner: "default",
    });
    setAnalyses((prev) => [analysis, ...prev]);
    setName("");
    setDescription("");
    setShowCreate(false);
  };

  return (
    <div className="p-8 max-w-4xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-semibold text-black">Analyses</h1>
        <button
          onClick={() => setShowCreate(true)}
          className="px-4 py-2 bg-red-600 text-white text-sm rounded hover:bg-red-700 transition-colors"
        >
          New Analysis
        </button>
      </div>

      {showCreate && (
        <div className="mb-6 p-4 border border-neutral-200 rounded-lg">
          <input
            type="text"
            placeholder="Analysis name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full px-3 py-2 border border-neutral-300 rounded mb-2 text-sm text-black"
          />
          <input
            type="text"
            placeholder="Description (optional)"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className="w-full px-3 py-2 border border-neutral-300 rounded mb-3 text-sm text-black"
          />
          <div className="flex gap-2">
            <button
              onClick={handleCreate}
              className="px-4 py-2 bg-red-600 text-white text-sm rounded hover:bg-red-700"
            >
              Create
            </button>
            <button
              onClick={() => setShowCreate(false)}
              className="px-4 py-2 bg-neutral-100 text-black text-sm rounded hover:bg-neutral-200"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      {loading ? (
        <p className="text-neutral-500 text-sm">Loading...</p>
      ) : analyses.length === 0 ? (
        <p className="text-neutral-500 text-sm">
          No analyses yet. Create one to get started.
        </p>
      ) : (
        <div className="space-y-2">
          {analyses.map((a) => (
            <Link
              key={a.id}
              href={`/analysis/${a.id}`}
              className="block p-4 border border-neutral-200 rounded-lg hover:border-red-300 transition-colors"
            >
              <div className="font-medium text-black">{a.name}</div>
              {a.description && (
                <div className="text-sm text-neutral-500 mt-1">
                  {a.description}
                </div>
              )}
              <div className="text-xs text-neutral-400 mt-2">
                {new Date(a.created_at).toLocaleDateString()}
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
