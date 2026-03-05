const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_URL}/api/v1${path}`, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `API error: ${res.status}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

import type {
  Analysis,
  Card,
  Dashboard,
  ExecuteAnalysisResponse,
} from "@/types";

export const api = {
  // Analyses
  listAnalyses: () => request<Analysis[]>("/analyses"),
  getAnalysis: (id: string) => request<Analysis>(`/analyses/${id}`),
  createAnalysis: (data: { name: string; description: string; owner: string }) =>
    request<Analysis>("/analyses", { method: "POST", body: JSON.stringify(data) }),
  updateAnalysis: (id: string, data: Partial<Analysis>) =>
    request<Analysis>(`/analyses/${id}`, { method: "PATCH", body: JSON.stringify(data) }),
  deleteAnalysis: (id: string) =>
    request<void>(`/analyses/${id}`, { method: "DELETE" }),

  // Cards
  listCards: (analysisId: string) =>
    request<Card[]>(`/analyses/${analysisId}/cards`),
  createCard: (analysisId: string, data: Partial<Card>) =>
    request<Card>(`/analyses/${analysisId}/cards`, { method: "POST", body: JSON.stringify(data) }),
  updateCard: (analysisId: string, cardId: string, data: Partial<Card>) =>
    request<Card>(`/analyses/${analysisId}/cards/${cardId}`, { method: "PATCH", body: JSON.stringify(data) }),
  deleteCard: (analysisId: string, cardId: string) =>
    request<void>(`/analyses/${analysisId}/cards/${cardId}`, { method: "DELETE" }),

  // Execute
  executeAnalysis: (analysisId: string) =>
    request<ExecuteAnalysisResponse>(`/analyses/${analysisId}/execute`, { method: "POST" }),

  // Dashboards
  listDashboards: () => request<Dashboard[]>("/dashboards"),
  getDashboard: (id: string) => request<Dashboard>(`/dashboards/${id}`),
  createDashboard: (data: Partial<Dashboard>) =>
    request<Dashboard>("/dashboards", { method: "POST", body: JSON.stringify(data) }),
  updateDashboard: (id: string, data: Partial<Dashboard>) =>
    request<Dashboard>(`/dashboards/${id}`, { method: "PATCH", body: JSON.stringify(data) }),
  deleteDashboard: (id: string) =>
    request<void>(`/dashboards/${id}`, { method: "DELETE" }),
};
