import { ChevronLeft, ChevronRight, Download, Search } from "lucide-react";
import React, { useEffect, useState } from "react";
import { api } from "../../lib/api";
import { FilterParams, GameSummary } from "../../types";

interface GameExplorerViewProps {
  filters: FilterParams;
  initialParams?: Record<string, any>;
  onSelectGame: (appId: number) => void;
}

export const GameExplorerView: React.FC<GameExplorerViewProps> = ({
  filters,
  initialParams,
  onSelectGame,
}) => {
  const [searchTerm, setSearchTerm] = useState(initialParams?.search || "");
  const [games, setGames] = useState<GameSummary[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [loading, setLoading] = useState(true);

  const fetchGames = () => {
    setLoading(true);
    const combinedFilters = {
      ...filters,
      search: searchTerm || undefined,
    };
    api
      .getGames(combinedFilters, 25, page * 25)
      .then((res) => {
        setGames(res.games || []);
        setTotal(res.total || 0);
      })
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchGames();
  }, [filters, page]);

  useEffect(() => {
    const timer = setTimeout(() => {
      setPage(0);
      fetchGames();
    }, 300);
    return () => clearTimeout(timer);
  }, [searchTerm]);

  const exportUrl = api.getExportGamesUrl({
    ...filters,
    search: searchTerm || undefined,
  });

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-5">
      {/* Search and Action Bar */}
      <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
        <div className="relative w-full md:w-96">
          <Search className="w-3.5 h-3.5 absolute left-3 top-2.5 text-slate-400" />
          <input
            type="text"
            placeholder="Search by game title or App ID..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-8 pr-3 py-1.5 rounded-md border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs focus:ring-1 focus:ring-indigo-500"
          />
        </div>

        <div className="flex items-center gap-3">
          <span className="text-xs font-mono text-slate-500">
            {total.toLocaleString()} games match filters
          </span>

          <a
            href={exportUrl}
            download
            className="flex items-center gap-1.5 px-3.5 py-1.5 rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-700 text-xs font-medium text-slate-700 dark:text-slate-200 shadow-sm transition-colors"
          >
            <Download className="w-3.5 h-3.5 text-slate-500" />
            <span>Export Catalog CSV</span>
          </a>
        </div>
      </div>

      {/* Main Table */}
      <div className="rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-xs text-left">
            <thead className="bg-slate-50 dark:bg-slate-950 text-slate-500 border-b border-slate-200 dark:border-slate-800 font-mono text-[11px]">
              <tr>
                <th className="py-2.5 px-4 w-24">App ID</th>
                <th className="py-2.5 px-4">Title</th>
                <th className="py-2.5 px-3">Release</th>
                <th className="py-2.5 px-3">Est. Owners</th>
                <th className="py-2.5 px-3">Price</th>
                <th className="py-2.5 px-3">Genres</th>
                <th className="py-2.5 px-3 text-center">Text / Audio</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-slate-800/80">
              {loading ? (
                <tr>
                  <td
                    colSpan={7}
                    className="py-12 text-center text-slate-400 font-mono"
                  >
                    Loading catalog records...
                  </td>
                </tr>
              ) : games.length === 0 ? (
                <tr>
                  <td colSpan={7} className="py-12 text-center text-slate-400">
                    No games match the current search or filter parameters.
                  </td>
                </tr>
              ) : (
                games.map((g) => (
                  <tr
                    key={g.appId}
                    onClick={() => onSelectGame(g.appId)}
                    className="hover:bg-slate-50 dark:hover:bg-slate-800/40 cursor-pointer transition-colors"
                  >
                    <td className="py-2 px-4 font-mono text-slate-500">
                      {g.appId}
                    </td>
                    <td className="py-2 px-4">
                      <div className="font-medium text-slate-900 dark:text-white flex items-center gap-1.5">
                        <span className="truncate max-w-sm">{g.name}</span>
                        {g.isPlaytest && (
                          <span className="px-1.5 py-0.2 rounded bg-amber-100 dark:bg-amber-950/60 text-amber-800 dark:text-amber-300 text-[9px] font-semibold shrink-0">
                            Playtest
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="py-2 px-3 font-mono text-slate-500">
                      {g.releaseDate || "—"}
                    </td>
                    <td className="py-2 px-3 font-mono text-slate-600 dark:text-slate-400">
                      {g.estimatedOwners}
                    </td>
                    <td className="py-2 px-3 font-mono">
                      {g.isFree ? (
                        <span className="text-emerald-600 dark:text-emerald-400 font-semibold">
                          Free
                        </span>
                      ) : (
                        `$${g.price.toFixed(2)}`
                      )}
                    </td>
                    <td className="py-2 px-3">
                      <div className="flex flex-wrap gap-1 max-w-xs">
                        {g.genres?.slice(0, 3).map((gn) => (
                          <span
                            key={gn}
                            className="px-1.5 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 text-[10px]"
                          >
                            {gn}
                          </span>
                        ))}
                      </div>
                    </td>
                    <td className="py-2 px-3 text-center font-mono">
                      <span className="font-semibold text-indigo-600 dark:text-indigo-400">
                        {g.textLanguages?.length || 0}
                      </span>
                      <span className="text-slate-300 dark:text-slate-600 mx-1">
                        /
                      </span>
                      <span className="font-semibold text-emerald-600 dark:text-emerald-400">
                        {g.audioLanguages?.length || 0}
                      </span>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination Footer */}
        <div className="px-5 py-3 border-t border-slate-200 dark:border-slate-800 bg-slate-50/60 dark:bg-slate-950/60 flex items-center justify-between text-xs">
          <span className="font-mono text-slate-500">
            Showing {page * 25 + 1}–{Math.min(total, (page + 1) * 25)} of{" "}
            {total.toLocaleString()} games
          </span>

          <div className="flex items-center gap-2">
            <button
              disabled={page === 0 || loading}
              onClick={() => setPage(page - 1)}
              className="p-1 rounded border border-slate-200 dark:border-slate-700 disabled:opacity-30 hover:bg-slate-50 dark:hover:bg-slate-800"
            >
              <ChevronLeft className="w-3.5 h-3.5" />
            </button>
            <span className="font-mono text-slate-600 dark:text-slate-400 px-1">
              Page {page + 1} of {Math.max(1, Math.ceil(total / 25))}
            </span>
            <button
              disabled={(page + 1) * 25 >= total || loading}
              onClick={() => setPage(page + 1)}
              className="p-1 rounded border border-slate-200 dark:border-slate-700 disabled:opacity-30 hover:bg-slate-50 dark:hover:bg-slate-800"
            >
              <ChevronRight className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
