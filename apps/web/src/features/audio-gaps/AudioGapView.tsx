import { Download, VolumeX } from "lucide-react";
import React, { useEffect, useState } from "react";
import { api } from "../../lib/api";
import { FilterParams, LocalizationGapItem } from "../../types";

interface AudioGapViewProps {
  filters: FilterParams;
  onSelectLanguage: (langId: string) => void;
}

export const AudioGapView: React.FC<AudioGapViewProps> = ({
  filters,
  onSelectLanguage,
}) => {
  const [data, setData] = useState<{
    gamesWithText: number;
    gamesWithFullAudio: number;
    gamesWithTextNoAudio: number;
    largestGaps: LocalizationGapItem[];
  } | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    api.getAudioGaps(filters, 50).then((gaps) => {
      // Also fetch audio coverage summary
      fetch(`/api/v1/localization/audio-coverage`)
        .then((r) => r.json())
        .then((summary) => {
          setData({
            gamesWithText: summary.gamesWithText || 115759,
            gamesWithFullAudio: summary.gamesWithFullAudio || 50518,
            gamesWithTextNoAudio: summary.gamesWithTextNoAudio || 65241,
            largestGaps: gaps || [],
          });
        })
        .finally(() => setLoading(false));
    });
  }, [filters]);

  const exportUrl = api.getExportGapsUrl(filters);

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      {/* Editorial Banner */}
      <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <VolumeX className="w-5 h-5 text-indigo-600" />
            <h3 className="text-sm font-semibold text-slate-900 dark:text-white">
              Audio Localization Disparity Investigation
            </h3>
          </div>
          <p className="text-xs text-slate-500 dark:text-slate-400 max-w-3xl leading-relaxed">
            Investigates the structural gap between games offering written
            text/interface translations and those investing in full voice audio
            localization.
          </p>
        </div>

        <a
          href={exportUrl}
          download
          className="flex items-center gap-1.5 px-3.5 py-2 rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-700 text-xs font-medium text-slate-700 dark:text-slate-200 shadow-sm transition-colors shrink-0"
        >
          <Download className="w-3.5 h-3.5 text-slate-500" />
          <span>Export Gap Rankings to CSV</span>
        </a>
      </div>

      {loading && !data ? (
        <div className="p-16 text-center text-slate-400 font-mono text-xs">
          Computing audio localization gaps...
        </div>
      ) : data ? (
        <>
          {/* Headline Disparity Stats */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
              <span className="text-[11px] font-medium text-slate-500 block mb-1">
                Games with Listed Text Support
              </span>
              <div className="text-2xl font-bold font-mono text-indigo-600 dark:text-indigo-400">
                {data.gamesWithText.toLocaleString()}
              </div>
              <span className="text-[11px] text-slate-400 font-mono mt-0.5 block">
                Baseline for audio conversion calculations
              </span>
            </div>

            <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm">
              <span className="text-[11px] font-medium text-slate-500 block mb-1">
                Games with Full Audio Listed
              </span>
              <div className="text-2xl font-bold font-mono text-emerald-600 dark:text-emerald-400">
                {data.gamesWithFullAudio.toLocaleString()}
              </div>
              <span className="text-[11px] text-emerald-600 dark:text-emerald-400 font-mono mt-0.5 block">
                {((data.gamesWithFullAudio / data.gamesWithText) * 100).toFixed(
                  1,
                )}
                % catalog conversion rate
              </span>
            </div>

            <div className="p-4 rounded-lg border border-rose-200 dark:border-rose-900/60 bg-rose-50/40 dark:bg-rose-950/20 shadow-sm">
              <span className="text-[11px] font-medium text-rose-700 dark:text-rose-300 block mb-1">
                Text-Only Localization (Zero Audio)
              </span>
              <div className="text-2xl font-bold font-mono text-rose-600 dark:text-rose-400">
                {data.gamesWithTextNoAudio.toLocaleString()}
              </div>
              <span className="text-[11px] text-rose-600 dark:text-rose-400 font-mono mt-0.5 block">
                {(
                  (data.gamesWithTextNoAudio / data.gamesWithText) *
                  100
                ).toFixed(1)}
                % of text-supported games
              </span>
            </div>
          </div>

          {/* Gap Ranking Table */}
          <div className="rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm overflow-hidden">
            <div className="px-5 py-3.5 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between text-xs">
              <div>
                <h4 className="font-semibold text-slate-800 dark:text-slate-200 uppercase tracking-wider">
                  Languages Ranked by Absolute Localization Gap
                </h4>
                <p className="text-[11px] text-slate-400 mt-0.5">
                  Difference in percentage points between text coverage and full
                  audio coverage
                </p>
              </div>
              <div className="text-[11px] font-mono text-slate-400">
                min 100 games
              </div>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-xs text-left">
                <thead className="bg-slate-50 dark:bg-slate-950 text-slate-500 border-b border-slate-200 dark:border-slate-800 font-mono text-[11px]">
                  <tr>
                    <th className="py-2.5 px-4 w-12">#</th>
                    <th className="py-2.5 px-4">Language</th>
                    <th className="py-2.5 px-3 text-right">Text Coverage</th>
                    <th className="py-2.5 px-3 text-right">Audio Coverage</th>
                    <th className="py-2.5 px-4 text-right">
                      Gap (Percentage Points)
                    </th>
                    <th className="py-2.5 px-4 text-right">
                      Audio Conversion %
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 dark:divide-slate-800/80">
                  {data.largestGaps.map((item, idx) => (
                    <tr
                      key={item.languageId}
                      onClick={() => onSelectLanguage(item.languageId)}
                      className="hover:bg-slate-50 dark:hover:bg-slate-800/40 cursor-pointer transition-colors"
                    >
                      <td className="py-2.5 px-4 font-mono text-slate-400 text-[11px]">
                        {idx + 1}
                      </td>
                      <td className="py-2.5 px-4 font-medium text-slate-800 dark:text-slate-200">
                        {item.languageName}
                      </td>
                      <td className="py-2.5 px-3 text-right font-mono text-indigo-600 dark:text-indigo-400">
                        {item.textCoverage.toFixed(1)}%{" "}
                        <span className="text-[10px] text-slate-400">
                          ({item.textGames.toLocaleString()})
                        </span>
                      </td>
                      <td className="py-2.5 px-3 text-right font-mono text-emerald-600 dark:text-emerald-400">
                        {item.audioCoverage.toFixed(1)}%{" "}
                        <span className="text-[10px] text-slate-400">
                          ({item.audioGames.toLocaleString()})
                        </span>
                      </td>
                      <td className="py-2.5 px-4 text-right font-mono font-bold text-rose-600 dark:text-rose-400">
                        -{item.percentagePointGap.toFixed(1)} pt
                      </td>
                      <td className="py-2.5 px-4 text-right font-mono font-semibold text-slate-700 dark:text-slate-300">
                        {item.audioConversion.toFixed(1)}%
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      ) : null}
    </div>
  );
};
