import {
  Calendar,
  ChevronLeft,
  ChevronRight,
  Download,
  Languages,
} from "lucide-react";
import React, { useEffect, useState } from "react";
import { api } from "../../lib/api";
import { FilterParams, LanguageDetail, LanguageRankItem } from "../../types";

interface LanguageExplorerViewProps {
  filters: FilterParams;
  initialLanguageId?: string;
  onSelectGame: (appId: number) => void;
}

export const LanguageExplorerView: React.FC<LanguageExplorerViewProps> = ({
  filters,
  initialLanguageId,
  onSelectGame,
}) => {
  const [languages, setLanguages] = useState<LanguageRankItem[]>([]);
  const [selectedLangId, setSelectedLangId] = useState<string>(
    initialLanguageId || "simplified_chinese",
  );
  const [detail, setDetail] = useState<LanguageDetail | null>(null);
  const [games, setGames] = useState<any[]>([]);
  const [gamesTotal, setGamesTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [loading, setLoading] = useState(true);
  const [gamesLoading, setGamesLoading] = useState(false);

  // Load language list on mount or filter change
  useEffect(() => {
    api.getLanguages(filters, 100).then((res) => {
      setLanguages(res.languages || []);
      if (
        res.languages?.length &&
        !initialLanguageId &&
        !res.languages.some((l) => l.id === selectedLangId)
      ) {
        setSelectedLangId(res.languages[0].id);
      }
    });
  }, [filters]);

  useEffect(() => {
    if (initialLanguageId) {
      setSelectedLangId(initialLanguageId);
    }
  }, [initialLanguageId]);

  // Load language details when selectedLangId changes
  useEffect(() => {
    if (!selectedLangId) return;
    setLoading(true);
    api
      .getLanguageDetail(selectedLangId, filters)
      .then((data) => setDetail(data))
      .finally(() => setLoading(false));

    // Reset pagination
    setPage(0);
  }, [selectedLangId, filters]);

  // Load games for language
  useEffect(() => {
    if (!selectedLangId) return;
    setGamesLoading(true);
    api
      .getLanguageGames(selectedLangId, filters, 15, page * 15)
      .then((res) => {
        setGames(res.games || []);
        setGamesTotal(res.total || 0);
      })
      .finally(() => setGamesLoading(false));
  }, [selectedLangId, filters, page]);

  const exportUrl = selectedLangId
    ? api.getExportLanguageUrl(selectedLangId, filters)
    : "#";

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      {/* Selector and Actions Bar */}
      <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
        <div className="flex items-center gap-3 w-full md:w-auto">
          <Languages className="w-5 h-5 text-indigo-600 shrink-0" />
          <div className="flex-1 md:w-72">
            <label className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block mb-1">
              Select Language to Investigate:
            </label>
            <select
              value={selectedLangId}
              onChange={(e) => setSelectedLangId(e.target.value)}
              className="w-full px-3 py-1.5 rounded-md border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs font-semibold"
            >
              {languages.map((l) => (
                <option key={l.id} value={l.id}>
                  {l.name} ({l.textCoverage.toFixed(1)}% coverage ·{" "}
                  {l.textGames.toLocaleString()} games)
                </option>
              ))}
            </select>
          </div>
        </div>

        <a
          href={exportUrl}
          download
          className="flex items-center gap-1.5 px-3.5 py-2 rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-700 text-xs font-medium text-slate-700 dark:text-slate-200 shadow-sm transition-colors shrink-0"
        >
          <Download className="w-3.5 h-3.5 text-slate-500" />
          <span>Export {detail?.name || "Language"} Games to CSV</span>
        </a>
      </div>

      {loading && !detail ? (
        <div className="p-16 text-center text-slate-400 font-mono text-xs">
          Loading language breakdown...
        </div>
      ) : detail ? (
        <>
          {/* Key Metrics Banner */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
              <span className="text-[11px] font-medium text-slate-500 block mb-1">
                Text/Interface Coverage
              </span>
              <div className="text-2xl font-bold font-mono text-indigo-600 dark:text-indigo-400">
                {detail.textCoverage.toFixed(1)}%
              </div>
              <span className="text-[11px] text-slate-400 font-mono mt-0.5 block">
                {detail.textGames.toLocaleString()} /{" "}
                {detail.eligibleGames.toLocaleString()} games
              </span>
            </div>

            <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
              <span className="text-[11px] font-medium text-slate-500 block mb-1">
                Full-Audio Coverage
              </span>
              <div className="text-2xl font-bold font-mono text-emerald-600 dark:text-emerald-400">
                {detail.audioCoverage.toFixed(1)}%
              </div>
              <span className="text-[11px] text-slate-400 font-mono mt-0.5 block">
                {detail.audioGames.toLocaleString()} catalog games
              </span>
            </div>

            <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
              <span className="text-[11px] font-medium text-slate-500 block mb-1">
                Audio Conversion
              </span>
              <div className="text-2xl font-bold font-mono text-slate-900 dark:text-white">
                {detail.audioConversion.toFixed(1)}%
              </div>
              <span className="text-[11px] text-slate-400 font-mono mt-0.5 block">
                Audio games / text games
              </span>
            </div>

            <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
              <span className="text-[11px] font-medium text-slate-500 block mb-1">
                Text-to-Audio Gap
              </span>
              <div className="text-2xl font-bold font-mono text-rose-600 dark:text-rose-400">
                -{detail.percentagePointGap.toFixed(1)} pt
              </div>
              <span className="text-[11px] text-slate-400 font-mono mt-0.5 block">
                Absolute percentage points
              </span>
            </div>
          </div>

          {/* Genre Coverage (Highest vs Lowest) */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="p-5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
              <h4 className="text-xs font-semibold text-slate-800 dark:text-slate-200 uppercase tracking-wider mb-3 pb-2 border-b border-slate-100 dark:border-slate-800 flex items-center justify-between">
                <span>Genres with Highest Representation</span>
                <span className="text-[11px] font-normal normal-case text-slate-400">
                  min 50 games
                </span>
              </h4>
              <div className="space-y-3">
                {detail.topGenres?.map((g) => (
                  <div key={g.genreId} className="space-y-1 text-xs">
                    <div className="flex justify-between">
                      <span className="font-medium text-slate-800 dark:text-slate-200">
                        {g.genreName}
                      </span>
                      <span className="font-mono font-semibold text-indigo-600 dark:text-indigo-400">
                        {g.textCoverage.toFixed(1)}%{" "}
                        <span className="text-slate-400 font-normal font-sans">
                          ({g.textGames.toLocaleString()} /{" "}
                          {g.eligibleGames.toLocaleString()})
                        </span>
                      </span>
                    </div>
                    <div className="w-full h-2 bg-slate-100 dark:bg-slate-800 rounded-sm overflow-hidden">
                      <div
                        className="h-full bg-indigo-600 rounded-sm"
                        style={{ width: `${g.textCoverage}%` }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="p-5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
              <h4 className="text-xs font-semibold text-slate-800 dark:text-slate-200 uppercase tracking-wider mb-3 pb-2 border-b border-slate-100 dark:border-slate-800 flex items-center justify-between">
                <span>Genres with Lowest Representation</span>
                <span className="text-[11px] font-normal normal-case text-slate-400">
                  min 50 games
                </span>
              </h4>
              <div className="space-y-3">
                {detail.lowestGenres?.map((g) => (
                  <div key={g.genreId} className="space-y-1 text-xs">
                    <div className="flex justify-between">
                      <span className="font-medium text-slate-800 dark:text-slate-200">
                        {g.genreName}
                      </span>
                      <span className="font-mono font-semibold text-slate-700 dark:text-slate-300">
                        {g.textCoverage.toFixed(1)}%{" "}
                        <span className="text-slate-400 font-normal font-sans">
                          ({g.textGames.toLocaleString()} /{" "}
                          {g.eligibleGames.toLocaleString()})
                        </span>
                      </span>
                    </div>
                    <div className="w-full h-2 bg-slate-100 dark:bg-slate-800 rounded-sm overflow-hidden">
                      <div
                        className="h-full bg-slate-400 dark:bg-slate-600 rounded-sm"
                        style={{ width: `${g.textCoverage}%` }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Release Year Trends */}
          {detail.yearlyTrends && detail.yearlyTrends.length > 0 && (
            <div className="p-5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
              <div className="flex items-center justify-between mb-4 pb-2 border-b border-slate-100 dark:border-slate-800">
                <h4 className="text-xs font-semibold text-slate-800 dark:text-slate-200 uppercase tracking-wider flex items-center gap-2">
                  <Calendar className="w-4 h-4 text-indigo-600" />
                  Release Year Adoption Trend (2010–2025)
                </h4>
                <div className="flex items-center gap-4 text-xs font-mono">
                  <span className="flex items-center gap-1.5 text-indigo-600">
                    <span className="w-2.5 h-2.5 rounded-sm bg-indigo-600" />
                    Text Coverage %
                  </span>
                  <span className="flex items-center gap-1.5 text-emerald-600">
                    <span className="w-2.5 h-2.5 rounded-sm bg-emerald-600" />
                    Audio Coverage %
                  </span>
                </div>
              </div>

              <div className="overflow-x-auto">
                <div className="min-w-[600px] flex items-end gap-2 h-40 pt-4 px-2">
                  {detail.yearlyTrends.map((yt) => {
                    const textH = Math.max(4, yt.textCoverage * 1.3);
                    const audioH = Math.max(4, yt.audioCoverage * 1.3);
                    return (
                      <div
                        key={yt.year}
                        className="flex-1 flex flex-col items-center gap-1.5 group"
                      >
                        <div className="w-full flex items-end justify-center gap-0.5 h-32 relative">
                          {/* Tooltip on hover */}
                          <div className="absolute -top-10 opacity-0 group-hover:opacity-100 transition-opacity bg-slate-900 text-white text-[10px] p-1.5 rounded shadow pointer-events-none z-20 whitespace-nowrap font-mono">
                            {yt.year}: Text {yt.textCoverage.toFixed(1)}% |
                            Audio {yt.audioCoverage.toFixed(1)}% ({yt.textGames}{" "}
                            games)
                          </div>
                          <div
                            className="w-1/2 bg-indigo-600 rounded-t-sm transition-all"
                            style={{ height: `${textH}px` }}
                          />
                          <div
                            className="w-1/2 bg-emerald-500 rounded-t-sm transition-all"
                            style={{ height: `${audioH}px` }}
                          />
                        </div>
                        <span className="text-[10px] font-mono text-slate-400 rotate-45 md:rotate-0 mt-1">
                          {String(yt.year).slice(2)}'
                        </span>
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>
          )}

          {/* Sample Games Listing Table */}
          <div className="rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm overflow-hidden">
            <div className="px-5 py-3 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
              <div>
                <h4 className="text-xs font-semibold text-slate-800 dark:text-slate-200 uppercase tracking-wider">
                  Representative Games Listing {detail.name}
                </h4>
                <p className="text-[11px] text-slate-400">
                  Showing {games.length} of {gamesTotal.toLocaleString()} titles
                </p>
              </div>

              {/* Pagination */}
              <div className="flex items-center gap-2 text-xs">
                <button
                  disabled={page === 0 || gamesLoading}
                  onClick={() => setPage(page - 1)}
                  className="p-1 rounded border border-slate-200 dark:border-slate-700 disabled:opacity-30 hover:bg-slate-50 dark:hover:bg-slate-800"
                >
                  <ChevronLeft className="w-3.5 h-3.5" />
                </button>
                <span className="font-mono text-slate-500">
                  Page {page + 1} of {Math.max(1, Math.ceil(gamesTotal / 15))}
                </span>
                <button
                  disabled={(page + 1) * 15 >= gamesTotal || gamesLoading}
                  onClick={() => setPage(page + 1)}
                  className="p-1 rounded border border-slate-200 dark:border-slate-700 disabled:opacity-30 hover:bg-slate-50 dark:hover:bg-slate-800"
                >
                  <ChevronRight className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-xs text-left">
                <thead className="bg-slate-50 dark:bg-slate-950 text-slate-500 border-b border-slate-200 dark:border-slate-800 font-mono text-[11px]">
                  <tr>
                    <th className="py-2.5 px-4 w-24">App ID</th>
                    <th className="py-2.5 px-4">Title</th>
                    <th className="py-2.5 px-4">Release</th>
                    <th className="py-2.5 px-4">Est. Owners</th>
                    <th className="py-2.5 px-4 text-center">Text Support</th>
                    <th className="py-2.5 px-4 text-center">Full Audio</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 dark:divide-slate-800/80">
                  {gamesLoading ? (
                    <tr>
                      <td
                        colSpan={6}
                        className="py-8 text-center text-slate-400 font-mono"
                      >
                        Loading games...
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
                        <td className="py-2 px-4 font-medium text-slate-800 dark:text-slate-200">
                          {g.name}
                        </td>
                        <td className="py-2 px-4 font-mono text-slate-500">
                          {g.releaseDate}
                        </td>
                        <td className="py-2 px-4 font-mono text-slate-500">
                          {g.estimatedOwners}
                        </td>
                        <td className="py-2 px-4 text-center">
                          {g.hasTextSupport ? (
                            <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-emerald-50 dark:bg-emerald-950 text-emerald-700 dark:text-emerald-300">
                              Listed
                            </span>
                          ) : (
                            <span className="text-slate-300">—</span>
                          )}
                        </td>
                        <td className="py-2 px-4 text-center">
                          {g.hasFullAudio ? (
                            <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-indigo-50 dark:bg-indigo-950 text-indigo-700 dark:text-indigo-300">
                              Full Audio
                            </span>
                          ) : (
                            <span className="text-slate-300">—</span>
                          )}
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </>
      ) : null}
    </div>
  );
};
