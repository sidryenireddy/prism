"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api } from "@/lib/api";
import type { Dashboard } from "@/types";

export default function DashboardsPage() {
  const [dashboards, setDashboards] = useState<Dashboard[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .listDashboards()
      .then(setDashboards)
      .catch(() => setDashboards([]))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="p-8 max-w-4xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-semibold text-black">Dashboards</h1>
      </div>

      {loading ? (
        <p className="text-neutral-500 text-sm">Loading...</p>
      ) : dashboards.length === 0 ? (
        <p className="text-neutral-500 text-sm">
          No dashboards yet. Publish a dashboard from an analysis to see it here.
        </p>
      ) : (
        <div className="space-y-2">
          {dashboards.map((d) => (
            <Link
              key={d.id}
              href={`/dashboard/${d.id}`}
              className="block p-4 border border-neutral-200 rounded-lg hover:border-red-300 transition-colors"
            >
              <div className="flex items-center justify-between">
                <div className="font-medium text-black">{d.name}</div>
                <span
                  className={`text-xs px-2 py-0.5 rounded ${
                    d.published
                      ? "bg-green-100 text-green-700"
                      : "bg-neutral-100 text-neutral-500"
                  }`}
                >
                  {d.published ? "Published" : "Draft"}
                </span>
              </div>
              <div className="text-xs text-neutral-400 mt-2">
                {new Date(d.created_at).toLocaleDateString()}
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
