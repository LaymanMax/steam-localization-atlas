import { AlertTriangle, ChevronRight, Download, Search } from "lucide-react";
import React, { useEffect, useState } from "react";
import { PageId } from "../../components/Sidebar";
import { api } from "../../lib/api";
import { FilterParams, MatrixCell, MatrixResponse } from "../../types";

interface MatrixViewProps {
  filters: FilterParams;
  onNavigate: (page: PageId, extraParams?: Record<string, any>) => void;
}

export const MatrixView: React.FC<MatrixViewProps> = ({
  filters,
  onNavigate,
}) => {
  const [metric, setMetric] = useState<"text" | "audio" | "conversion">("text");
  const [search, setSearch] = useState("");
  const [minSample, setMinSample] = useState(30);
  const [limit, setLimit] = useState(35);
  const [sortBy, setSortBy] = useState<"overall" | "alphabetical" | "gap">(
    "overall",
  );

  const [data, setData] = useState<MatrixResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Active hovered or focused cell for detailed tooltip/inspection
  const [hoveredCell, setHoveredCell] = useState<{
    langName: string;
    langId: string;
    genreName: string;
    genreId: string;
    cell: MatrixCell;
  } | null>(null);

  const fetchMatrix = async () => {
    setLoading(true);
    setError(null);
    try {
      const resp = await api.getMatrix(
        metric,
        filters,
        search,
        minSample,
        limit,
        sortBy,
      );
      setData(resp);
    } catch (err: any) {
      setError(err?.message || "Failed to load localization matrix");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMatrix();
  }, [metric, filters, minSample, limit, sortBy]);

  // Debounced search
  useEffect(() => {
    const timer = setTimeout(() => {
      fetchMatrix();
    }, 300);
    return () => clearTimeout(timer);
  }, [search]);

  const handleCellClick = (_langId: string, genreId: string) => {
    onNavigate("games", {
      genres: [genreId],
      search: "",
    });
  };

  // Color intensity calculation for heatmap
  const getCellBg = (pct: number, isLowSample: boolean) => {
    if (isLowSample) {
      return "bg-amber-50/50 dark:bg-amber-950/20 text-slate-500 border border-dashed border-amber-300 dark:border-amber-800/60";
    }
    if (pct === 0) return "bg-slate-50 dark:bg-slate-900/40 text-slate-400";
    if (pct < 10)
      return "bg-indigo-50/60 dark:bg-indigo-950/30 text-indigo-900 dark:text-indigo-200";
    if (pct < 25)
      return "bg-indigo-100 dark:bg-indigo-950/70 text-indigo-900 dark:text-indigo-100 font-medium";
    if (pct < 50)
      return "bg-indigo-200 dark:bg-indigo-900/60 text-indigo-950 dark:text-white font-medium";
    if (pct < 75)
      return "bg-indigo-400 dark:bg-indigo-700 text-white font-semibold";
    return "bg-indigo-600 dark:bg-indigo-600 text-white font-bold";
  };

  const exportUrl = api.getExportMatrixUrl(
    metric,
    filters,
    search,
    minSample,
    limit,
    sortBy,
  );

  return (
    <div className="p-6 max-w-[1600px] mx-auto space-y-5">
      {/* Editorial Methodology & Definitions Header */}
      <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <span className="px-2 py-0.5 rounded bg-indigo-100 dark:bg-indigo-950 text-indigo-700 dark:text-indigo-300 font-mono text-[11px] font-semibold uppercase tracking-wider">
              Flagship Instrument
            </span>
            <h3 className="text-sm font-semibold text-slate-900 dark:text-white">
              Localization Coverage Heatmap
            </h3>
          </div>
          <p className="text-xs text-slate-500 dark:text-slate-400 max-w-3xl leading-relaxed">
            Every cell displays the observed representation percentage of
            language (L) within genre (G). Denominators are strictly
            genre-specific and exclude records without usable language metadata.
          </p>
        </div>

        <a
          href={exportUrl}
          download
          className="flex items-center gap-1.5 px-3.5 py-2 rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-700 text-xs font-medium text-slate-700 dark:text-slate-200 shadow-sm transition-colors shrink-0"
        >
          <Download className="w-3.5 h-3.5 text-slate-500" />
          <span>Export View to CSV</span>
        </a>
      </div>

      {/* Interactive Controls Bar */}
      <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm space-y-4">
        {/* Metric Selector Tabs */}
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-slate-100 dark:border-slate-800 pb-3">
          <div className="flex items-center gap-1">
            <span className="text-xs font-semibold text-slate-600 dark:text-slate-400 mr-2">
              Metric:
            </span>
            <div className="flex rounded-md border border-slate-200 dark:border-slate-700 p-0.5 bg-slate-50 dark:bg-slate-950 text-xs font-medium">
              <button
                onClick={() => setMetric("text")}
                className={`px-3 py-1.5 rounded transition-colors ${
                  metric === "text"
                    ? "bg-indigo-600 text-white font-semibold shadow-sm"
                    : "text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
                }`}
              >
                Listed Text/Interface Coverage
              </button>
              <button
                onClick={() => setMetric("audio")}
                className={`px-3 py-1.5 rounded transition-colors ${
                  metric === "audio"
                    ? "bg-indigo-600 text-white font-semibold shadow-sm"
                    : "text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
                }`}
              >
                Listed Full-Audio Coverage
              </button>
              <button
                onClick={() => setMetric("conversion")}
                className={`px-3 py-1.5 rounded transition-colors ${
                  metric === "conversion"
                    ? "bg-indigo-600 text-white font-semibold shadow-sm"
                    : "text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
                }`}
              >
                Audio Conversion Conditional on Text
              </button>
            </div>
          </div>

          <div className="text-[11px] font-mono text-slate-500">
            {metric === "text" &&
              "Formula: (Distinct games in G listing L for text) / (Eligible games in G) × 100"}
            {metric === "audio" &&
              "Formula: (Distinct games in G listing L for audio) / (Eligible games in G) × 100"}
            {metric === "conversion" &&
              "Formula: (Games in G with audio L) / (Games in G with text L) × 100"}
          </div>
        </div>

        {/* Secondary Filters & Controls */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-3 text-xs">
          {/* Language Search */}
          <div className="relative">
            <Search className="w-3.5 h-3.5 absolute left-3 top-2.5 text-slate-400" />
            <input
              type="text"
              placeholder="Search languages..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-8 pr-3 py-1.5 rounded border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs focus:ring-1 focus:ring-indigo-500"
            />
          </div>

          {/* Sort By */}
          <div className="flex items-center gap-2">
            <span className="text-slate-500 shrink-0">Sort:</span>
            <select
              value={sortBy}
              onChange={(e: any) => setSortBy(e.target.value)}
              className="w-full px-2.5 py-1.5 rounded border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs"
            >
              <option value="overall">Highest Overall Coverage</option>
              <option value="alphabetical">Alphabetical (A–Z)</option>
              <option value="gap">Largest Text-to-Audio Gap</option>
            </select>
          </div>

          {/* Display Limit */}
          <div className="flex items-center gap-2">
            <span className="text-slate-500 shrink-0">Show:</span>
            <select
              value={limit}
              onChange={(e) => setLimit(parseInt(e.target.value, 10))}
              className="w-full px-2.5 py-1.5 rounded border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs"
            >
              <option value="15">Top 15 Languages</option>
              <option value="25">Top 25 Languages</option>
              <option value="35">Top 35 Languages</option>
              <option value="50">Top 50 Languages</option>
              <option value="75">Top 75 Languages</option>
            </select>
          </div>

          {/* Sample Size Safeguard */}
          <div className="flex items-center gap-2">
            <span className="text-slate-500 shrink-0">Min Sample:</span>
            <select
              value={minSample}
              onChange={(e) => setMinSample(parseInt(e.target.value, 10))}
              className="w-full px-2.5 py-1.5 rounded border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs"
            >
              <option value="10">10 Games (Minimal)</option>
              <option value="30">30 Games (Default)</option>
              <option value="100">100 Games (Strict)</option>
              <option value="250">250 Games (Robust)</option>
            </select>
          </div>
        </div>
      </div>

      {/* Heatmap Legend */}
      <div className="flex flex-wrap items-center justify-between gap-3 px-1 text-xs text-slate-500">
        <div className="flex items-center gap-1.5">
          <span className="font-medium mr-1 text-slate-700 dark:text-slate-300">
            Coverage Scale:
          </span>
          <span className="px-1.5 py-0.5 text-[10px] rounded bg-slate-100 dark:bg-slate-800 text-slate-600">
            0%
          </span>
          <span className="px-1.5 py-0.5 text-[10px] rounded bg-indigo-50 dark:bg-indigo-950/30 text-indigo-900 dark:text-indigo-200">
            1–10%
          </span>
          <span className="px-1.5 py-0.5 text-[10px] rounded bg-indigo-100 dark:bg-indigo-950/70 text-indigo-900 dark:text-indigo-100">
            10–25%
          </span>
          <span className="px-1.5 py-0.5 text-[10px] rounded bg-indigo-200 dark:bg-indigo-900/60 text-indigo-950 dark:text-white">
            25–50%
          </span>
          <span className="px-1.5 py-0.5 text-[10px] rounded bg-indigo-400 text-white">
            50–75%
          </span>
          <span className="px-1.5 py-0.5 text-[10px] rounded bg-indigo-600 text-white font-bold">
            75–100%
          </span>
        </div>

        <div className="flex items-center gap-3">
          <span className="flex items-center gap-1 text-[11px] text-amber-600 dark:text-amber-400">
            <span className="w-2.5 h-2.5 rounded-sm bg-amber-50 dark:bg-amber-950/40 border border-dashed border-amber-400" />
            Low Sample (&lt;{minSample} games)
          </span>
          <span className="text-[11px] text-slate-400">
            Click any cell to inspect filtered catalog
          </span>
        </div>
      </div>

      {/* Main Heatmap Table Container */}
      <div className="rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm overflow-hidden">
        {loading ? (
          <div className="p-16 text-center text-slate-400 font-mono text-xs animate-pulse">
            Calculating coverage matrix over{" "}
            {data?.totalEligible?.toLocaleString() || 115000} distinct game
            records...
          </div>
        ) : error ? (
          <div className="p-8 text-center text-rose-600 text-xs">{error}</div>
        ) : !data || data.languages.length === 0 ? (
          <div className="p-12 text-center text-slate-500 text-xs">
            No matching languages or genres found for the current filter
            criteria.
          </div>
        ) : (
          <div className="overflow-x-auto max-h-[70vh]">
            <table className="w-full text-xs border-collapse">
              {/* Sticky Table Header */}
              <thead className="sticky top-0 z-10 bg-slate-50 dark:bg-slate-950 border-b border-slate-200 dark:border-slate-800 shadow-sm font-mono">
                <tr>
                  <th className="sticky left-0 z-20 bg-slate-50 dark:bg-slate-950 text-left py-2.5 px-3.5 font-semibold text-slate-800 dark:text-slate-200 w-44 min-w-[176px] border-r border-slate-200 dark:border-slate-800">
                    Language
                  </th>
                  <th className="text-right py-2.5 px-2.5 font-semibold text-slate-600 dark:text-slate-400 w-24 min-w-[96px] border-r border-slate-200 dark:border-slate-800">
                    Overall %
                  </th>
                  {data.genres.map((g) => (
                    <th
                      key={g.id}
                      className="py-2.5 px-2 text-center font-semibold text-slate-800 dark:text-slate-200 min-w-[100px] border-r border-slate-100 dark:border-slate-800/60"
                    >
                      <div
                        className="truncate font-sans font-medium text-[11px]"
                        title={g.name}
                      >
                        {g.name}
                      </div>
                      <div className="text-[10px] text-slate-400 font-normal">
                        n={g.eligibleGames.toLocaleString()}
                      </div>
                    </th>
                  ))}
                </tr>
              </thead>

              {/* Table Body */}
              <tbody className="divide-y divide-slate-100 dark:divide-slate-800/80">
                {data.languages.map((l) => (
                  <tr
                    key={l.id}
                    className="hover:bg-slate-50/70 dark:hover:bg-slate-800/40 transition-colors"
                  >
                    {/* Sticky Language Label Column */}
                    <td
                      onClick={() =>
                        onNavigate("languages", { languageId: l.id })
                      }
                      className="sticky left-0 z-10 bg-white dark:bg-slate-900 group-hover:bg-slate-50 dark:group-hover:bg-slate-800/70 py-2 px-3.5 font-medium text-slate-800 dark:text-slate-200 border-r border-slate-200 dark:border-slate-800 truncate cursor-pointer hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors"
                      title={l.name}
                    >
                      <div className="flex items-center justify-between">
                        <span className="truncate">{l.name}</span>
                        <ChevronRight className="w-3 h-3 text-slate-300 dark:text-slate-600 opacity-0 group-hover:opacity-100 transition-opacity" />
                      </div>
                    </td>

                    {/* Overall Coverage Column */}
                    <td className="text-right py-2 px-2.5 font-mono font-semibold text-slate-900 dark:text-white border-r border-slate-200 dark:border-slate-800 bg-slate-50/40 dark:bg-slate-950/20">
                      {l.overallCoverage.toFixed(1)}%
                    </td>

                    {/* Genre Cells */}
                    {data.genres.map((g) => {
                      const cell = l.cells[g.id] || {
                        numerator: 0,
                        denominator: g.eligibleGames,
                        percentage: 0,
                        isLowSample: false,
                      };

                      const bgClass = getCellBg(
                        cell.percentage,
                        cell.isLowSample,
                      );

                      return (
                        <td
                          key={g.id}
                          tabIndex={0}
                          onClick={() => handleCellClick(l.id, g.id)}
                          onMouseEnter={() =>
                            setHoveredCell({
                              langName: l.name,
                              langId: l.id,
                              genreName: g.name,
                              genreId: g.id,
                              cell,
                            })
                          }
                          onFocus={() =>
                            setHoveredCell({
                              langName: l.name,
                              langId: l.id,
                              genreName: g.name,
                              genreId: g.id,
                              cell,
                            })
                          }
                          className={`py-2 px-1 text-center font-mono text-[11px] cursor-pointer transition-all border-r border-slate-100 dark:border-slate-800/40 hover:ring-2 hover:ring-indigo-500 hover:z-10 relative ${bgClass}`}
                        >
                          <div className="flex items-center justify-center gap-0.5">
                            <span>
                              {cell.percentage > 0
                                ? `${cell.percentage.toFixed(1)}%`
                                : "—"}
                            </span>
                            {cell.isLowSample && (
                              <AlertTriangle className="w-2.5 h-2.5 text-amber-500 shrink-0" />
                            )}
                          </div>
                        </td>
                      );
                    })}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Persistent Inspection Card for Hovered/Focused Cell */}
      <div className="p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm text-xs flex items-center justify-between">
        {hoveredCell ? (
          <div className="flex flex-wrap items-center gap-4">
            <div className="flex items-center gap-1.5">
              <span className="font-semibold text-slate-900 dark:text-white">
                {hoveredCell.langName}
              </span>
              <span className="text-slate-400">in</span>
              <span className="font-semibold text-slate-900 dark:text-white">
                {hoveredCell.genreName}
              </span>
            </div>

            <div className="flex items-center gap-3 font-mono text-[11px]">
              <span className="px-2 py-0.5 rounded bg-indigo-50 dark:bg-indigo-950/50 text-indigo-700 dark:text-indigo-300 font-bold">
                Coverage: {hoveredCell.cell.percentage.toFixed(1)}%
              </span>
              <span className="text-slate-600 dark:text-slate-400">
                Sample: {hoveredCell.cell.numerator.toLocaleString()} supported
                / {hoveredCell.cell.denominator.toLocaleString()} eligible games
              </span>
              {hoveredCell.cell.isLowSample && (
                <span className="text-amber-600 dark:text-amber-400 flex items-center gap-1 font-semibold">
                  <AlertTriangle className="w-3 h-3" />
                  Small sample (&lt;{minSample})
                </span>
              )}
            </div>
          </div>
        ) : (
          <span className="text-slate-400 italic">
            Hover over or tab into any cell to inspect exact numerators,
            denominators, and sample counts.
          </span>
        )}

        <div className="text-[11px] text-slate-400 font-mono">
          Distinct Game Aggregation
        </div>
      </div>
    </div>
  );
};
