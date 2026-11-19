import { Info, Users } from "lucide-react";
import React, { useEffect, useState } from "react";
import { api } from "../../lib/api";
import {
  FilterParams,
  LanguageRankItem,
  OwnershipCoverageItem,
} from "../../types";

interface OwnershipAnalysisViewProps {
  filters: FilterParams;
  onSelectLanguage?: (langId: string) => void;
}

export const OwnershipAnalysisView: React.FC<OwnershipAnalysisViewProps> = ({
  filters,
}) => {
  const [items, setItems] = useState<OwnershipCoverageItem[]>([]);
  const [languages, setLanguages] = useState<LanguageRankItem[]>([]);
  const [selectedLang, setSelectedLang] = useState<string>("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.getLanguages(filters, 50).then((res) => {
      setLanguages(res.languages || []);
    });
  }, [filters]);

  useEffect(() => {
    setLoading(true);
    api
      .getOwnershipCoverage(filters, selectedLang || undefined)
      .then((data) => setItems(data || []))
      .finally(() => setLoading(false));
  }, [filters, selectedLang]);

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      {/* Editorial Disclaimer Box */}
      <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm flex items-start gap-3">
        <Info className="w-4 h-4 text-indigo-600 dark:text-indigo-400 shrink-0 mt-0.5" />
        <div className="text-xs text-slate-600 dark:text-slate-300 leading-relaxed">
          <span className="font-semibold text-slate-900 dark:text-white">
            Methodological Safeguard:
          </span>{" "}
          Estimated ownership brackets in this dataset originate from public
          estimate ranges (e.g. SteamSpy algorithms), not verified Steam
          financial records. Brackets should be interpreted as categorical
          groupings, not exact sales counts or commercial returns. Do not infer
          that larger ownership caused better localization or vice versa.
        </div>
      </div>

      {/* Language filter control */}
      <div className="p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm flex flex-wrap items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <Users className="w-5 h-5 text-indigo-600" />
          <div className="w-64">
            <label className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block mb-1">
              Analyze Specific Language Across Brackets:
            </label>
            <select
              value={selectedLang}
              onChange={(e) => setSelectedLang(e.target.value)}
              className="w-full px-3 py-1.5 rounded-md border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs font-semibold"
            >
              <option value="">All Catalog Games (Overall)</option>
              {languages.map((l) => (
                <option key={l.id} value={l.id}>
                  {l.name}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="text-xs font-mono text-slate-500">
          {items.length} Estimated Ownership Brackets Analyzed
        </div>
      </div>

      {/* Main Table */}
      <div className="rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm overflow-hidden">
        <div className="px-5 py-3 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between text-xs">
          <h4 className="font-semibold text-slate-800 dark:text-slate-200 uppercase tracking-wider">
            {selectedLang
              ? `Coverage of ${selectedLang} by Ownership Bracket`
              : "Catalog Localization by Ownership Bracket"}
          </h4>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-xs text-left">
            <thead className="bg-slate-50 dark:bg-slate-950 text-slate-500 border-b border-slate-200 dark:border-slate-800 font-mono text-[11px]">
              <tr>
                <th className="py-2.5 px-4">Estimated Owners Bracket</th>
                <th className="py-2.5 px-4 text-right">Total Games</th>
                <th className="py-2.5 px-4 text-right">Text Supported Games</th>
                <th className="py-2.5 px-4 text-right">Text Coverage %</th>
                <th className="py-2.5 px-4 text-right">Audio Games</th>
                <th className="py-2.5 px-4 text-right">Audio Coverage %</th>
                <th className="py-2.5 px-4 text-right">Gap (pt)</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-slate-800/80">
              {loading ? (
                <tr>
                  <td
                    colSpan={7}
                    className="py-8 text-center text-slate-400 font-mono"
                  >
                    Loading ownership breakdown...
                  </td>
                </tr>
              ) : (
                items.map((item) => (
                  <tr
                    key={item.bracket}
                    className="hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors"
                  >
                    <td className="py-2.5 px-4 font-mono font-medium text-slate-800 dark:text-slate-200">
                      {item.bracket}
                    </td>
                    <td className="py-2.5 px-4 text-right font-mono text-slate-500">
                      {item.totalGames.toLocaleString()}
                    </td>
                    <td className="py-2.5 px-4 text-right font-mono text-slate-700 dark:text-slate-300">
                      {item.textGames.toLocaleString()}
                    </td>
                    <td className="py-2.5 px-4 text-right font-mono font-bold text-indigo-600 dark:text-indigo-400">
                      {item.textCoverage.toFixed(1)}%
                    </td>
                    <td className="py-2.5 px-4 text-right font-mono text-slate-700 dark:text-slate-300">
                      {item.audioGames.toLocaleString()}
                    </td>
                    <td className="py-2.5 px-4 text-right font-mono font-semibold text-emerald-600 dark:text-emerald-400">
                      {item.audioCoverage.toFixed(1)}%
                    </td>
                    <td className="py-2.5 px-4 text-right font-mono text-rose-600 dark:text-rose-400">
                      -{item.percentagePointGap.toFixed(1)} pt
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
