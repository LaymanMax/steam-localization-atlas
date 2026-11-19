import { ArrowRightLeft, Download, Tags } from "lucide-react";
import React, { useEffect, useState } from "react";
import { api } from "../../lib/api";
import {
  FilterParams,
  GenreComparison,
  GenreDetail,
  MatrixGenre,
} from "../../types";

interface GenreAnalysisViewProps {
  filters: FilterParams;
  initialGenreId?: string;
  onSelectLanguage: (langId: string) => void;
}

export const GenreAnalysisView: React.FC<GenreAnalysisViewProps> = ({
  filters,
  initialGenreId,
  onSelectLanguage,
}) => {
  const [genres, setGenres] = useState<MatrixGenre[]>([]);
  const [selectedGenreId, setSelectedGenreId] = useState<string>(
    initialGenreId || "rpg",
  );
  const [compareGenreId, setCompareGenreId] = useState<string>("strategy");
  const [activeTab, setActiveTab] = useState<"profile" | "compare">("profile");

  const [detail, setDetail] = useState<GenreDetail | null>(null);
  const [comparison, setComparison] = useState<GenreComparison | null>(null);
  const [loading, setLoading] = useState(true);

  // Load available genres
  useEffect(() => {
    api.getFilterOptions().then((opts) => {
      if (opts?.genres?.length) {
        setGenres(opts.genres);
        if (
          !initialGenreId &&
          !opts.genres.some((g) => g.id === selectedGenreId)
        ) {
          setSelectedGenreId(opts.genres[0].id);
        }
      }
    });
  }, []);

  useEffect(() => {
    if (initialGenreId) {
      setSelectedGenreId(initialGenreId);
    }
  }, [initialGenreId]);

  // Load genre profile
  useEffect(() => {
    if (!selectedGenreId) return;
    setLoading(true);
    api
      .getGenreDetail(selectedGenreId, filters)
      .then((data) => setDetail(data))
      .finally(() => setLoading(false));
  }, [selectedGenreId, filters]);

  // Load comparison when compare tab is active
  useEffect(() => {
    if (activeTab === "compare" && selectedGenreId && compareGenreId) {
      api
        .compareGenres(selectedGenreId, compareGenreId, filters)
        .then((data) => setComparison(data));
    }
  }, [activeTab, selectedGenreId, compareGenreId, filters]);

  const exportUrl = selectedGenreId
    ? api.getExportGenreUrl(selectedGenreId, filters)
    : "#";

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      {/* Top Selector & Mode Bar */}
      <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
        <div className="flex flex-wrap items-center gap-3">
          <Tags className="w-5 h-5 text-indigo-600 shrink-0" />
          <div className="w-56">
            <label className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block mb-1">
              Primary Genre:
            </label>
            <select
              value={selectedGenreId}
              onChange={(e) => setSelectedGenreId(e.target.value)}
              className="w-full px-3 py-1.5 rounded-md border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs font-semibold"
            >
              {genres.map((g) => (
                <option key={g.id} value={g.id}>
                  {g.name} ({g.eligibleGames.toLocaleString()} games)
                </option>
              ))}
            </select>
          </div>

          {/* Mode Switch Tabs */}
          <div className="flex rounded-md border border-slate-200 dark:border-slate-700 p-0.5 bg-slate-50 dark:bg-slate-950 text-xs font-medium mt-4 md:mt-0">
            <button
              onClick={() => setActiveTab("profile")}
              className={`px-3 py-1.5 rounded transition-colors ${
                activeTab === "profile"
                  ? "bg-indigo-600 text-white font-semibold shadow-sm"
                  : "text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
              }`}
            >
              Genre Profile
            </button>
            <button
              onClick={() => setActiveTab("compare")}
              className={`px-3 py-1.5 rounded transition-colors flex items-center gap-1.5 ${
                activeTab === "compare"
                  ? "bg-indigo-600 text-white font-semibold shadow-sm"
                  : "text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
              }`}
            >
              <ArrowRightLeft className="w-3.5 h-3.5" />
              <span>Genre-to-Genre Delta</span>
            </button>
          </div>
        </div>

        <a
          href={exportUrl}
          download
          className="flex items-center gap-1.5 px-3.5 py-2 rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-700 text-xs font-medium text-slate-700 dark:text-slate-200 shadow-sm transition-colors shrink-0"
        >
          <Download className="w-3.5 h-3.5 text-slate-500" />
          <span>Export Genre Stats to CSV</span>
        </a>
      </div>

      {activeTab === "profile" ? (
        loading && !detail ? (
          <div className="p-16 text-center text-slate-400 font-mono text-xs">
            Loading genre analytics...
          </div>
        ) : detail ? (
          <>
            {/* Headline Metrics */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
              <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
                <span className="text-[11px] font-medium text-slate-500 block mb-1">
                  Eligible Genre Titles
                </span>
                <div className="text-2xl font-bold font-mono text-indigo-600 dark:text-indigo-400">
                  {detail.eligibleGames.toLocaleString()}
                </div>
                <span className="text-[11px] text-slate-400 font-mono mt-0.5 block">
                  {detail.totalGames.toLocaleString()} total games in genre
                </span>
              </div>

              <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
                <span className="text-[11px] font-medium text-slate-500 block mb-1">
                  Median Languages per Game
                </span>
                <div className="text-2xl font-bold font-mono text-slate-900 dark:text-white">
                  {detail.medianLanguages}
                </div>
                <span className="text-[11px] text-slate-400 font-mono mt-0.5 block">
                  Robust central tendency
                </span>
              </div>

              <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
                <span className="text-[11px] font-medium text-slate-500 block mb-1">
                  Average Languages per Game
                </span>
                <div className="text-2xl font-bold font-mono text-slate-900 dark:text-white">
                  {detail.averageLanguages.toFixed(1)}
                </div>
                <span className="text-[11px] text-slate-400 font-mono mt-0.5 block">
                  Arithmetic mean
                </span>
              </div>

              <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
                <span className="text-[11px] font-medium text-slate-500 block mb-1">
                  Multi-Genre Methodology Note
                </span>
                <div className="text-xs text-slate-600 dark:text-slate-300 leading-snug">
                  Games belonging to multiple genres count once per applicable
                  genre denominator.
                </div>
              </div>
            </div>

            {/* Top Languages and Lowest Languages in Genre */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Top Supported Languages */}
              <div className="rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm overflow-hidden">
                <div className="px-5 py-3.5 border-b border-slate-200 dark:border-slate-800">
                  <h4 className="text-xs font-semibold text-slate-800 dark:text-slate-200 uppercase tracking-wider">
                    Most Frequently Listed Languages in {detail.name}
                  </h4>
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full text-xs text-left">
                    <thead className="bg-slate-50 dark:bg-slate-950 text-slate-500 border-b border-slate-200 dark:border-slate-800 font-mono text-[11px]">
                      <tr>
                        <th className="py-2.5 px-4">Language</th>
                        <th className="py-2.5 px-3 text-right">Text %</th>
                        <th className="py-2.5 px-3 text-right">Audio %</th>
                        <th className="py-2.5 px-3 text-right">Conversion</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100 dark:divide-slate-800/80">
                      {detail.topLanguages?.map((l) => (
                        <tr
                          key={l.id}
                          onClick={() => onSelectLanguage(l.id)}
                          className="hover:bg-slate-50 dark:hover:bg-slate-800/40 cursor-pointer transition-colors"
                        >
                          <td className="py-2 px-4 font-medium text-slate-800 dark:text-slate-200">
                            {l.name}
                          </td>
                          <td className="py-2 px-3 text-right font-mono font-semibold text-indigo-600 dark:text-indigo-400">
                            {l.textCoverage.toFixed(1)}%
                          </td>
                          <td className="py-2 px-3 text-right font-mono text-emerald-600 dark:text-emerald-400">
                            {l.audioCoverage.toFixed(1)}%
                          </td>
                          <td className="py-2 px-3 text-right font-mono text-slate-600 dark:text-slate-400">
                            {l.audioConversion.toFixed(1)}%
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>

              {/* Least Frequently Listed Languages */}
              <div className="rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm overflow-hidden">
                <div className="px-5 py-3.5 border-b border-slate-200 dark:border-slate-800">
                  <h4 className="text-xs font-semibold text-slate-800 dark:text-slate-200 uppercase tracking-wider">
                    Least Frequently Listed Languages (min 5 games)
                  </h4>
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full text-xs text-left">
                    <thead className="bg-slate-50 dark:bg-slate-950 text-slate-500 border-b border-slate-200 dark:border-slate-800 font-mono text-[11px]">
                      <tr>
                        <th className="py-2.5 px-4">Language</th>
                        <th className="py-2.5 px-3 text-right">Text %</th>
                        <th className="py-2.5 px-3 text-right">Games</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100 dark:divide-slate-800/80">
                      {detail.lowestLanguages?.map((l) => (
                        <tr
                          key={l.id}
                          onClick={() => onSelectLanguage(l.id)}
                          className="hover:bg-slate-50 dark:hover:bg-slate-800/40 cursor-pointer transition-colors"
                        >
                          <td className="py-2 px-4 font-medium text-slate-800 dark:text-slate-200">
                            {l.name}
                          </td>
                          <td className="py-2 px-3 text-right font-mono font-medium text-slate-700 dark:text-slate-300">
                            {l.textCoverage.toFixed(2)}%
                          </td>
                          <td className="py-2 px-3 text-right font-mono text-slate-400">
                            {l.textGames}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </>
        ) : null
      ) : (
        /* Genre Comparison Tool */
        <div className="space-y-5">
          <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm flex flex-wrap items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <span className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                Compare Against:
              </span>
              <select
                value={compareGenreId}
                onChange={(e) => setCompareGenreId(e.target.value)}
                className="px-3 py-1.5 rounded-md border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs font-semibold"
              >
                {genres
                  .filter((g) => g.id !== selectedGenreId)
                  .map((g) => (
                    <option key={g.id} value={g.id}>
                      {g.name} ({g.eligibleGames.toLocaleString()} games)
                    </option>
                  ))}
              </select>
            </div>

            <div className="text-[11px] font-mono text-slate-500">
              Delta = (Coverage in {comparison?.genreA.name}) - (Coverage in{" "}
              {comparison?.genreB.name})
            </div>
          </div>

          {comparison && (
            <div className="rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm overflow-hidden">
              <div className="px-5 py-3 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between text-xs">
                <span className="font-semibold text-slate-800 dark:text-slate-200">
                  {comparison.genreA.name} (n=
                  {comparison.genreA.eligibleGames.toLocaleString()}) vs{" "}
                  {comparison.genreB.name} (n=
                  {comparison.genreB.eligibleGames.toLocaleString()})
                </span>
                <span className="text-[11px] text-slate-400 font-mono">
                  Percentage-point differences
                </span>
              </div>

              <div className="overflow-x-auto">
                <table className="w-full text-xs text-left">
                  <thead className="bg-slate-50 dark:bg-slate-950 text-slate-500 border-b border-slate-200 dark:border-slate-800 font-mono text-[11px]">
                    <tr>
                      <th className="py-2.5 px-4">Language</th>
                      <th className="py-2.5 px-4 text-right">
                        {comparison.genreA.name} Coverage %
                      </th>
                      <th className="py-2.5 px-4 text-right">
                        {comparison.genreB.name} Coverage %
                      </th>
                      <th className="py-2.5 px-4 text-right">
                        Difference (pt)
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100 dark:divide-slate-800/80">
                    {comparison.comparison.map((item) => {
                      const isPositive = item.differencePt > 0;
                      const isNegative = item.differencePt < 0;

                      return (
                        <tr
                          key={item.languageId}
                          onClick={() => onSelectLanguage(item.languageId)}
                          className="hover:bg-slate-50 dark:hover:bg-slate-800/40 cursor-pointer transition-colors"
                        >
                          <td className="py-2 px-4 font-medium text-slate-800 dark:text-slate-200">
                            {item.languageName}
                          </td>
                          <td className="py-2 px-4 text-right font-mono text-slate-700 dark:text-slate-300">
                            {item.coverageA.toFixed(1)}% (
                            {item.gamesA.toLocaleString()})
                          </td>
                          <td className="py-2 px-4 text-right font-mono text-slate-700 dark:text-slate-300">
                            {item.coverageB.toFixed(1)}% (
                            {item.gamesB.toLocaleString()})
                          </td>
                          <td className="py-2 px-4 text-right font-mono font-bold">
                            <span
                              className={`inline-flex items-center gap-0.5 px-2 py-0.5 rounded text-[11px] ${
                                isPositive
                                  ? "bg-emerald-50 dark:bg-emerald-950 text-emerald-700 dark:text-emerald-300"
                                  : isNegative
                                    ? "bg-rose-50 dark:bg-rose-950 text-rose-700 dark:text-rose-300"
                                    : "text-slate-400"
                              }`}
                            >
                              {isPositive && "+"}
                              {item.differencePt.toFixed(1)} pt
                            </span>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};
