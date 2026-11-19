import {
  ArrowRight,
  BarChart3,
  Info,
  Languages,
  Layers,
  TrendingDown,
  Volume2,
} from "lucide-react";
import React, { useState } from "react";
import { PageId } from "../../components/Sidebar";
import { OverviewSummary } from "../../types";

interface OverviewViewProps {
  summary: OverviewSummary | null;
  loading: boolean;
  onNavigate: (page: PageId, extraParams?: Record<string, any>) => void;
}

export const OverviewView: React.FC<OverviewViewProps> = ({
  summary,
  loading,
  onNavigate,
}) => {
  const [chartAMetric, setChartAMetric] = useState<"text" | "audio">("text");

  if (loading && !summary) {
    return (
      <div className="p-8 text-center text-slate-500 font-mono text-xs">
        Loading descriptive analytics...
      </div>
    );
  }

  if (!summary) {
    return null;
  }

  const topLangs = summary.topLanguages || [];
  const maxRankCount = Math.max(
    ...topLangs.map((l) =>
      chartAMetric === "text" ? l.textGames : l.audioGames,
    ),
    1,
  );

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      {/* Editorial Notice Banner */}
      <div className="p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm flex items-start gap-3">
        <Info className="w-4 h-4 text-indigo-600 dark:text-indigo-400 shrink-0 mt-0.5" />
        <div className="text-xs text-slate-600 dark:text-slate-300 leading-relaxed">
          <span className="font-semibold text-slate-900 dark:text-white">
            Descriptive Research Notice:
          </span>{" "}
          This application measures observed localization coverage listed in the
          Steam games catalog. It distinguishes listed text/interface support
          from full voice audio support. Observed representation rates reflect
          publisher listings, not verified commercial demand, regional player
          populations, or sales opportunities.
        </div>
      </div>

      {/* Primary Headline Metrics */}
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-7 gap-3">
        <div className="p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
          <span className="text-[11px] font-medium text-slate-500 dark:text-slate-400 block mb-1">
            Total Catalog Games
          </span>
          <div className="text-xl font-bold font-mono tracking-tight text-slate-900 dark:text-white">
            {summary.totalGames.toLocaleString()}
          </div>
          <span className="text-[10px] text-slate-400 font-mono mt-1 block">
            Imported Parquet Records
          </span>
        </div>

        <div className="p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
          <span className="text-[11px] font-medium text-slate-500 dark:text-slate-400 block mb-1">
            Eligible Text Games
          </span>
          <div className="text-xl font-bold font-mono tracking-tight text-slate-900 dark:text-white">
            {summary.gamesWithLanguages.toLocaleString()}
          </div>
          <span className="text-[10px] text-emerald-600 dark:text-emerald-400 font-mono mt-1 block">
            {((summary.gamesWithLanguages / summary.totalGames) * 100).toFixed(
              1,
            )}
            % usable language metadata
          </span>
        </div>

        <div className="p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
          <span className="text-[11px] font-medium text-slate-500 dark:text-slate-400 block mb-1">
            Full Audio Support
          </span>
          <div className="text-xl font-bold font-mono tracking-tight text-slate-900 dark:text-white">
            {summary.gamesWithAudio.toLocaleString()}
          </div>
          <span className="text-[10px] text-indigo-600 dark:text-indigo-400 font-mono mt-1 block">
            {(
              (summary.gamesWithAudio / summary.gamesWithLanguages) *
              100
            ).toFixed(1)}
            % of eligible games
          </span>
        </div>

        <div className="p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
          <span className="text-[11px] font-medium text-slate-500 dark:text-slate-400 block mb-1">
            Unique Languages
          </span>
          <div className="text-xl font-bold font-mono tracking-tight text-slate-900 dark:text-white">
            {summary.uniqueLanguages}
          </div>
          <span className="text-[10px] text-slate-400 font-mono mt-1 block">
            Normalized Language IDs
          </span>
        </div>

        <div className="p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
          <span className="text-[11px] font-medium text-slate-500 dark:text-slate-400 block mb-1">
            Unique Genres
          </span>
          <div className="text-xl font-bold font-mono tracking-tight text-slate-900 dark:text-white">
            {summary.uniqueGenres}
          </div>
          <span className="text-[10px] text-slate-400 font-mono mt-1 block">
            Distinct Genre Slugs
          </span>
        </div>

        <div className="p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
          <span className="text-[11px] font-medium text-slate-500 dark:text-slate-400 block mb-1">
            Median Text Languages
          </span>
          <div className="text-xl font-bold font-mono tracking-tight text-slate-900 dark:text-white">
            {summary.medianTextLanguages}
          </div>
          <span className="text-[10px] text-slate-400 font-mono mt-1 block">
            Avg: {summary.averageTextLanguages.toFixed(1)} per game
          </span>
        </div>

        <div className="p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
          <span className="text-[11px] font-medium text-slate-500 dark:text-slate-400 block mb-1">
            Median Audio Languages
          </span>
          <div className="text-xl font-bold font-mono tracking-tight text-slate-900 dark:text-white">
            {summary.medianAudioLanguages}
          </div>
          <span className="text-[10px] text-slate-400 font-mono mt-1 block">
            Avg: {summary.averageAudioLanguages.toFixed(1)} per game
          </span>
        </div>
      </div>

      {/* Grid: Chart A (Most Supported Languages) & Chart B (Language Distribution) */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Chart A: Most Supported Languages */}
        <div className="rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-5 shadow-sm">
          <div className="flex items-center justify-between mb-4 pb-3 border-b border-slate-100 dark:border-slate-800">
            <div>
              <h3 className="text-sm font-semibold text-slate-900 dark:text-white flex items-center gap-2">
                <Languages className="w-4 h-4 text-indigo-600" />
                Chart A: Most Listed Languages
              </h3>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                Ranked by game count across eligible catalog titles (
                {summary.gamesWithLanguages.toLocaleString()})
              </p>
            </div>

            <div className="flex rounded border border-slate-200 dark:border-slate-700 p-0.5 text-xs">
              <button
                onClick={() => setChartAMetric("text")}
                className={`px-2.5 py-1 rounded text-xs font-medium transition-colors ${
                  chartAMetric === "text"
                    ? "bg-indigo-600 text-white font-semibold"
                    : "text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
                }`}
              >
                Text Support
              </button>
              <button
                onClick={() => setChartAMetric("audio")}
                className={`px-2.5 py-1 rounded text-xs font-medium transition-colors ${
                  chartAMetric === "audio"
                    ? "bg-indigo-600 text-white font-semibold"
                    : "text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
                }`}
              >
                Full Audio
              </button>
            </div>
          </div>

          <div className="space-y-2.5">
            {topLangs.slice(0, 10).map((l, idx) => {
              const count =
                chartAMetric === "text" ? l.textGames : l.audioGames;
              const cov =
                chartAMetric === "text" ? l.textCoverage : l.audioCoverage;
              const barWidth = Math.max(2, (count / maxRankCount) * 100);

              return (
                <div
                  key={l.id}
                  onClick={() => onNavigate("languages", { languageId: l.id })}
                  className="group flex items-center gap-3 text-xs cursor-pointer hover:bg-slate-50 dark:hover:bg-slate-800/40 p-1.5 rounded transition-colors"
                >
                  <span className="w-4 text-slate-400 font-mono text-[11px] text-right">
                    {idx + 1}
                  </span>
                  <span className="w-36 font-medium text-slate-800 dark:text-slate-200 truncate group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors">
                    {l.name}
                  </span>
                  <div className="flex-1 h-4 bg-slate-100 dark:bg-slate-800 rounded-sm overflow-hidden flex items-center">
                    <div
                      className={`h-full transition-all duration-500 ${
                        chartAMetric === "text"
                          ? "bg-indigo-600"
                          : "bg-emerald-600"
                      }`}
                      style={{ width: `${barWidth}%` }}
                    />
                  </div>
                  <span className="w-16 font-mono font-medium text-slate-900 dark:text-white text-right">
                    {cov.toFixed(1)}%
                  </span>
                  <span className="w-20 font-mono text-slate-400 text-[11px] text-right">
                    {count.toLocaleString()}
                  </span>
                </div>
              );
            })}
          </div>

          <div className="mt-4 pt-3 border-t border-slate-100 dark:border-slate-800 flex justify-end">
            <button
              onClick={() => onNavigate("matrix")}
              className="text-xs font-medium text-indigo-600 dark:text-indigo-400 hover:underline flex items-center gap-1"
            >
              <span>Explore All Languages in Matrix</span>
              <ArrowRight className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        {/* Chart B: Distribution of Languages per Game */}
        <div className="rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-5 shadow-sm flex flex-col justify-between">
          <div>
            <div className="mb-4 pb-3 border-b border-slate-100 dark:border-slate-800">
              <h3 className="text-sm font-semibold text-slate-900 dark:text-white flex items-center gap-2">
                <BarChart3 className="w-4 h-4 text-indigo-600" />
                Chart B: Language Support Breadth Distribution
              </h3>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                Proportion of games listing single-language vs multi-language
                localization
              </p>
            </div>

            <div className="space-y-3.5">
              {summary.languageDistribution?.map((bin) => (
                <div key={bin.binLabel} className="space-y-1">
                  <div className="flex justify-between text-xs">
                    <span className="font-medium text-slate-700 dark:text-slate-300">
                      {bin.binLabel}
                    </span>
                    <span className="font-mono text-slate-900 dark:text-white font-semibold">
                      {bin.percent.toFixed(1)}%{" "}
                      <span className="text-slate-400 font-normal">
                        ({bin.count.toLocaleString()} games)
                      </span>
                    </span>
                  </div>
                  <div className="w-full h-2.5 bg-slate-100 dark:bg-slate-800 rounded-sm overflow-hidden">
                    <div
                      className="h-full bg-slate-700 dark:bg-slate-300 rounded-sm"
                      style={{ width: `${bin.percent}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Key Insight Box */}
          <div className="mt-5 p-3 rounded bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 text-[11px] text-slate-600 dark:text-slate-400 leading-normal">
            <span className="font-semibold text-slate-800 dark:text-slate-200">
              Catalog skew:
            </span>{" "}
            Most titles in the catalog offer either 1 language (often English or
            native developer language) or a standard multi-language localization
            bundle of 6–10 languages (EFIGS + CJK).
          </div>
        </div>
      </div>

      {/* Grid: Chart C & Chart D & Chart E */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Chart C: Text vs Full Audio Gap */}
        <div className="rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-5 shadow-sm">
          <div className="mb-4 pb-3 border-b border-slate-100 dark:border-slate-800">
            <h3 className="text-sm font-semibold text-slate-900 dark:text-white flex items-center gap-2">
              <TrendingDown className="w-4 h-4 text-rose-500" />
              Chart C: Text vs Audio Coverage
            </h3>
            <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              Overall catalog comparison
            </p>
          </div>

          <div className="space-y-4">
            <div className="p-4 rounded-lg bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 text-center">
              <span className="text-xs text-slate-500 block mb-1">
                Catalog Text Coverage
              </span>
              <div className="text-3xl font-mono font-bold text-indigo-600 dark:text-indigo-400">
                {(
                  (summary.gamesWithLanguages / summary.totalGames) *
                  100
                ).toFixed(1)}
                %
              </div>
              <span className="text-[11px] text-slate-400 font-mono mt-1 block">
                {summary.gamesWithLanguages.toLocaleString()} games
              </span>
            </div>

            <div className="p-4 rounded-lg bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 text-center">
              <span className="text-xs text-slate-500 block mb-1">
                Catalog Audio Coverage
              </span>
              <div className="text-3xl font-mono font-bold text-emerald-600 dark:text-emerald-400">
                {((summary.gamesWithAudio / summary.totalGames) * 100).toFixed(
                  1,
                )}
                %
              </div>
              <span className="text-[11px] text-slate-400 font-mono mt-1 block">
                {summary.gamesWithAudio.toLocaleString()} games
              </span>
            </div>

            <div className="p-3 rounded border border-rose-100 dark:border-rose-950/60 bg-rose-50/50 dark:bg-rose-950/20 text-xs">
              <div className="flex justify-between items-center text-rose-800 dark:text-rose-300 font-semibold mb-1">
                <span>Text-Without-Audio Titles:</span>
                <span className="font-mono text-sm">
                  {(
                    summary.gamesWithLanguages - summary.gamesWithAudio
                  ).toLocaleString()}
                </span>
              </div>
              <p className="text-[11px] text-slate-600 dark:text-slate-400 leading-tight">
                {(
                  ((summary.gamesWithLanguages - summary.gamesWithAudio) /
                    summary.gamesWithLanguages) *
                  100
                ).toFixed(1)}
                % of text-localized games list zero full-audio languages.
              </p>
            </div>
          </div>
        </div>

        {/* Chart D: Top Genres by Language Diversity */}
        <div className="rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-5 shadow-sm">
          <div className="mb-4 pb-3 border-b border-slate-100 dark:border-slate-800">
            <h3 className="text-sm font-semibold text-slate-900 dark:text-white flex items-center gap-2">
              <Layers className="w-4 h-4 text-indigo-600" />
              Chart D: Broadest Genre Localization
            </h3>
            <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              Ranked by median listed languages per game
            </p>
          </div>

          <div className="space-y-3">
            {summary.topGenresCoverage?.slice(0, 6).map((g) => (
              <div
                key={g.id}
                onClick={() => onNavigate("genres", { genreId: g.id })}
                className="group flex items-center justify-between text-xs p-2 rounded hover:bg-slate-50 dark:hover:bg-slate-800/50 cursor-pointer transition-colors"
              >
                <div>
                  <span className="font-medium text-slate-800 dark:text-slate-200 block group-hover:text-indigo-600 dark:group-hover:text-indigo-400">
                    {g.name}
                  </span>
                  <span className="text-[10px] text-slate-400 font-mono">
                    {g.eligibleGames.toLocaleString()} eligible games
                  </span>
                </div>
                <div className="text-right font-mono">
                  <div className="font-bold text-slate-900 dark:text-white">
                    {g.medianTextLanguages} langs
                  </div>
                  <div className="text-[10px] text-slate-400">
                    avg {g.averageTextLanguages.toFixed(1)}
                  </div>
                </div>
              </div>
            ))}
          </div>

          <div className="mt-4 pt-3 border-t border-slate-100 dark:border-slate-800 flex justify-end">
            <button
              onClick={() => onNavigate("genres")}
              className="text-xs font-medium text-indigo-600 dark:text-indigo-400 hover:underline flex items-center gap-1"
            >
              <span>View All Genre Breakdowns</span>
              <ArrowRight className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        {/* Chart E: Largest Localization Gaps */}
        <div className="rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-5 shadow-sm">
          <div className="mb-4 pb-3 border-b border-slate-100 dark:border-slate-800">
            <h3 className="text-sm font-semibold text-slate-900 dark:text-white flex items-center gap-2">
              <Volume2 className="w-4 h-4 text-indigo-600" />
              Chart E: Largest Text-to-Audio Gaps
            </h3>
            <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              Languages with highest absolute percentage point drops
            </p>
          </div>

          <div className="space-y-3">
            {summary.localizationGaps?.slice(0, 6).map((gap) => (
              <div
                key={gap.languageId}
                onClick={() =>
                  onNavigate("languages", { languageId: gap.languageId })
                }
                className="group flex items-center justify-between text-xs p-2 rounded hover:bg-slate-50 dark:hover:bg-slate-800/50 cursor-pointer transition-colors"
              >
                <div>
                  <span className="font-medium text-slate-800 dark:text-slate-200 block group-hover:text-indigo-600 dark:group-hover:text-indigo-400">
                    {gap.languageName}
                  </span>
                  <span className="text-[10px] text-slate-400 font-mono">
                    Text: {gap.textCoverage.toFixed(1)}% · Audio:{" "}
                    {gap.audioCoverage.toFixed(1)}%
                  </span>
                </div>
                <div className="text-right font-mono">
                  <div className="font-bold text-rose-600 dark:text-rose-400">
                    -{gap.percentagePointGap.toFixed(1)} pt
                  </div>
                  <div className="text-[10px] text-slate-400">
                    {gap.audioConversion.toFixed(1)}% conversion
                  </div>
                </div>
              </div>
            ))}
          </div>

          <div className="mt-4 pt-3 border-t border-slate-100 dark:border-slate-800 flex justify-end">
            <button
              onClick={() => onNavigate("audio-gaps")}
              className="text-xs font-medium text-indigo-600 dark:text-indigo-400 hover:underline flex items-center gap-1"
            >
              <span>Explore Audio Gap Analysis</span>
              <ArrowRight className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
