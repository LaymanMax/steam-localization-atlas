import {
  DatasetStatus,
  FilterOptions,
  FilterParams,
  GameSummary,
  GenreComparison,
  GenreDetail,
  LanguageDetail,
  LanguageRankItem,
  MatrixResponse,
  OverviewSummary,
  QualityIssueItem,
  QualitySummary,
} from "../types";

const API_BASE = "/api/v1";

function buildQuery(params: Record<string, any>): string {
  const parts: string[] = [];
  for (const [key, val] of Object.entries(params)) {
    if (val === undefined || val === null || val === "") continue;
    if (Array.isArray(val)) {
      if (val.length > 0)
        parts.push(`${key}=${encodeURIComponent(val.join(","))}`);
    } else {
      parts.push(`${key}=${encodeURIComponent(String(val))}`);
    }
  }
  return parts.length > 0 ? `?${parts.join("&")}` : "";
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
    ...options,
  });

  if (!res.ok) {
    let errMsg = `HTTP Error ${res.status}`;
    try {
      const errJson = await res.json();
      if (errJson?.error?.message) {
        errMsg = errJson.error.message;
      }
    } catch {
      // ignore JSON parse error on non-json error responses
    }
    throw new Error(errMsg);
  }

  return res.json();
}

export const api = {
  getHealth: () => request<{ status: string; hasData: boolean }>("/health"),

  getDatasetStatus: () => request<DatasetStatus>("/dataset/status"),

  getDatasetMetadata: () => request<Record<string, any>>("/dataset/metadata"),

  startImport: (source: string, file?: string) =>
    request<{ message: string; source: string }>("/dataset/import", {
      method: "POST",
      body: JSON.stringify({ source, file }),
    }),

  getFilterOptions: () => request<FilterOptions>("/filters"),

  getOverview: (filters: FilterParams) =>
    request<OverviewSummary>(`/overview${buildQuery(filters)}`),

  getMatrix: (
    metric: "text" | "audio" | "conversion",
    filters: FilterParams,
    search?: string,
    minSample?: number,
    limit?: number,
    sortBy?: string,
  ) =>
    request<MatrixResponse>(
      `/localization/matrix${buildQuery({
        metric,
        ...filters,
        search,
        minSample,
        limit,
        sortBy,
      })}`,
    ),

  getLanguages: (filters: FilterParams, limit = 50) =>
    request<{ languages: LanguageRankItem[]; totalEligible: number }>(
      `/languages${buildQuery({ ...filters, limit })}`,
    ),

  getLanguageDetail: (id: string, filters: FilterParams) =>
    request<LanguageDetail>(
      `/languages/${encodeURIComponent(id)}${buildQuery(filters)}`,
    ),

  getLanguageGames: (
    id: string,
    filters: FilterParams,
    limit = 25,
    offset = 0,
  ) =>
    request<{ games: any[]; total: number; limit: number; offset: number }>(
      `/languages/${encodeURIComponent(id)}/games${buildQuery({ ...filters, limit, offset })}`,
    ),

  getGenres: () => request<any[]>("/genres"),

  getGenreDetail: (id: string, filters: FilterParams) =>
    request<GenreDetail>(
      `/genres/${encodeURIComponent(id)}${buildQuery(filters)}`,
    ),

  compareGenres: (genreA: string, genreB: string, filters: FilterParams) =>
    request<GenreComparison>(
      `/genres/${encodeURIComponent(genreA)}/compare${buildQuery({
        against: genreB,
        ...filters,
      })}`,
    ),

  getAudioGaps: (filters: FilterParams, limit = 20) =>
    request<any[]>(`/localization/gaps${buildQuery({ ...filters, limit })}`),

  getOwnershipCoverage: (filters: FilterParams, language?: string) =>
    request<any[]>(
      `/ownership/coverage${buildQuery({ ...filters, language })}`,
    ),

  getGames: (filters: FilterParams, limit = 25, offset = 0) =>
    request<{
      games: GameSummary[];
      total: number;
      limit: number;
      offset: number;
    }>(`/games${buildQuery({ ...filters, limit, offset })}`),

  getGame: (appId: number) => request<GameSummary>(`/games/${appId}`),

  getQualitySummary: () => request<QualitySummary>("/quality/summary"),

  getQualityIssues: (type?: string, limit = 25, offset = 0) =>
    request<{ issues: QualityIssueItem[]; total: number }>(
      `/quality/issues${buildQuery({ type, limit, offset })}`,
    ),

  getExportMatrixUrl: (
    metric: string,
    filters: FilterParams,
    search?: string,
    minSample?: number,
    limit?: number,
    sortBy?: string,
  ) =>
    `${API_BASE}/exports/matrix${buildQuery({
      metric,
      ...filters,
      search,
      minSample,
      limit,
      sortBy,
    })}`,

  getExportGamesUrl: (filters: FilterParams) =>
    `${API_BASE}/exports/games${buildQuery(filters)}`,

  getExportLanguageUrl: (id: string, filters: FilterParams) =>
    `${API_BASE}/exports/language/${encodeURIComponent(id)}${buildQuery(filters)}`,

  getExportGenreUrl: (id: string, filters: FilterParams) =>
    `${API_BASE}/exports/genre/${encodeURIComponent(id)}${buildQuery(filters)}`,

  getExportGapsUrl: (filters: FilterParams) =>
    `${API_BASE}/exports/gaps${buildQuery(filters)}`,
};
