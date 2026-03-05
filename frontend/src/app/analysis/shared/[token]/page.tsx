"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { api } from "@/lib/api";
import type { Analysis } from "@/types";

export default function SharedAnalysisPage() {
  const params = useParams();
  const router = useRouter();
  const token = params.token as string;

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    api
      .getAnalysisByShareToken(token)
      .then((analysis: Analysis) => {
        router.replace(`/analysis/${analysis.id}`);
      })
      .catch(() => {
        setError("Analysis not found or share link is invalid.");
        setLoading(false);
      });
  }, [token, router]);

  if (loading && !error) {
    return (
      <div className="p-8">
        <p className="text-neutral-500 text-sm">Loading shared analysis...</p>
      </div>
    );
  }

  return (
    <div className="p-8">
      <p className="text-red-600 text-sm">{error}</p>
    </div>
  );
}
