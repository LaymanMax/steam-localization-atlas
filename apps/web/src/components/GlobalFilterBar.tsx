import { RotateCcw, X } from "lucide-react";
import React from "react";
import { FilterOptions, FilterParams } from "../types";

interface GlobalFilterBarProps {
  isOpen: boolean;
  onClose: () => void;
  filters: FilterParams;
  onChangeFilters: (f: FilterParams) => void;
  filterOptions?: FilterOptions | null;
}

export const GlobalFilterBar: React.FC<GlobalFilterBarProps> = ({
  isOpen,
  onClose,
  filters,
  onChangeFilters,
  filterOptions,
}) => {
  if (!isOpen) return null;

  const minYear = filterOptions?.releaseYearMin || 2000;
  const maxYear = filterOptions?.releaseYearMax || 2026;

  const handleReset = () => {
    onChangeFilters({
      excludePlaytests: true,
      requireLanguages: false,
    });
  };

  const toggleGenre = (genreId: string) => {
    const current = filters.genres || [];
    if (current.includes(genreId)) {
      onChangeFilters({
        ...filters,
        genres: current.filter((g) => g !== genreId),
      });
    } else {
      onChangeFilters({ ...filters, genres: [...current, genreId] });
    }
  };

  const toggleOwnership = (bracket: string) => {
    const current = filters.ownership || [];
    if (current.includes(bracket)) {
      onChangeFilters({
        ...filters,
        ownership: current.filter((o) => o !== bracket),
      });
    } else {
      onChangeFilters({ ...filters, ownership: [...current, bracket] });
    }
  };

  return (
    <div className="border-b border-slate-200 dark:border-slate-800 bg-slate-50/80 dark:bg-slate-900/80 px-6 py-4 transition-all">
      <div className="flex items-center justify-between pb-3 mb-3 border-b border-slate-200 dark:border-slate-800">
        <div className="flex items-center gap-2">
          <span className="text-xs font-semibold uppercase tracking-wider text-slate-700 dark:text-slate-300">
            Global Analysis Filters
          </span>
          <span className="text-[11px] text-slate-500">
            (Applies consistently across all metrics and visualizations)
          </span>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={handleReset}
            className="flex items-center gap-1 text-xs text-slate-500 hover:text-slate-800 dark:hover:text-slate-200 transition-colors"
          >
            <RotateCcw className="w-3 h-3" />
            <span>Reset Filters</span>
          </button>
          <button
            onClick={onClose}
            className="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 text-xs">
        {/* Release Period */}
        <div>
          <label className="block font-medium text-slate-700 dark:text-slate-300 mb-1.5">
            Release Period
          </label>
          <div className="flex items-center gap-2">
            <input
              type="number"
              placeholder={String(minYear)}
              value={filters.releaseYearMin || ""}
              onChange={(e) =>
                onChangeFilters({
                  ...filters,
                  releaseYearMin: e.target.value
                    ? parseInt(e.target.value, 10)
                    : undefined,
                })
              }
              className="w-full px-2.5 py-1.5 rounded border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs"
            />
            <span className="text-slate-400">—</span>
            <input
              type="number"
              placeholder={String(maxYear)}
              value={filters.releaseYearMax || ""}
              onChange={(e) =>
                onChangeFilters({
                  ...filters,
                  releaseYearMax: e.target.value
                    ? parseInt(e.target.value, 10)
                    : undefined,
                })
              }
              className="w-full px-2.5 py-1.5 rounded border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs"
            />
          </div>
        </div>

        {/* Platform & Price */}
        <div>
          <label className="block font-medium text-slate-700 dark:text-slate-300 mb-1.5">
            Platform & Pricing
          </label>
          <div className="grid grid-cols-2 gap-2">
            <select
              value={filters.platform || ""}
              onChange={(e) =>
                onChangeFilters({
                  ...filters,
                  platform: e.target.value || undefined,
                })
              }
              className="w-full px-2 py-1.5 rounded border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs"
            >
              <option value="">All Platforms</option>
              <option value="windows">Windows</option>
              <option value="mac">macOS</option>
              <option value="linux">Linux</option>
            </select>

            <select
              value={filters.isFree === undefined ? "" : String(filters.isFree)}
              onChange={(e) => {
                const val = e.target.value;
                onChangeFilters({
                  ...filters,
                  isFree: val === "" ? undefined : val === "true",
                });
              }}
              className="w-full px-2 py-1.5 rounded border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs"
            >
              <option value="">All Pricing</option>
              <option value="false">Paid Games</option>
              <option value="true">Free to Play</option>
            </select>
          </div>
        </div>

        {/* Quality Safeguards */}
        <div className="md:col-span-2">
          <label className="block font-medium text-slate-700 dark:text-slate-300 mb-1.5">
            Methodology Safeguards
          </label>
          <div className="flex flex-wrap items-center gap-4 pt-1">
            <label className="flex items-center gap-2 cursor-pointer select-none">
              <input
                type="checkbox"
                checked={filters.excludePlaytests}
                onChange={(e) =>
                  onChangeFilters({
                    ...filters,
                    excludePlaytests: e.target.checked,
                  })
                }
                className="rounded border-slate-300 text-indigo-600 focus:ring-indigo-500 h-3.5 w-3.5"
              />
              <span className="text-slate-700 dark:text-slate-300">
                Exclude suspected playtests/betas (8,166 titles)
              </span>
            </label>

            <label className="flex items-center gap-2 cursor-pointer select-none">
              <input
                type="checkbox"
                checked={filters.requireLanguages}
                onChange={(e) =>
                  onChangeFilters({
                    ...filters,
                    requireLanguages: e.target.checked,
                  })
                }
                className="rounded border-slate-300 text-indigo-600 focus:ring-indigo-500 h-3.5 w-3.5"
              />
              <span className="text-slate-700 dark:text-slate-300">
                Exclude games missing language metadata
              </span>
            </label>
          </div>
        </div>
      </div>

      {/* Genre pills quick filter */}
      {filterOptions?.genres && (
        <div className="mt-3 pt-3 border-t border-slate-200 dark:border-slate-800/80">
          <span className="text-[11px] font-medium text-slate-500 dark:text-slate-400 block mb-1.5">
            Filter by Top Genres:
          </span>
          <div className="flex flex-wrap gap-1.5 max-h-16 overflow-y-auto">
            {filterOptions.genres.slice(0, 16).map((g) => {
              const isSelected = filters.genres?.includes(g.id);
              return (
                <button
                  key={g.id}
                  onClick={() => toggleGenre(g.id)}
                  className={`px-2 py-0.5 rounded text-[11px] font-medium transition-colors ${
                    isSelected
                      ? "bg-indigo-600 text-white"
                      : "bg-slate-200/70 dark:bg-slate-800 text-slate-700 dark:text-slate-300 hover:bg-slate-300 dark:hover:bg-slate-700"
                  }`}
                >
                  {g.name}
                </button>
              );
            })}
          </div>
        </div>
      )}

      {/* Ownership brackets quick filter */}
      {filterOptions?.ownershipBrackets && (
        <div className="mt-2.5 pt-2 border-t border-slate-200 dark:border-slate-800/80">
          <span className="text-[11px] font-medium text-slate-500 dark:text-slate-400 block mb-1.5">
            Filter by Approximate Estimated Ownership:
          </span>
          <div className="flex flex-wrap gap-1.5">
            {filterOptions.ownershipBrackets.slice(0, 8).map((b) => {
              const isSelected = filters.ownership?.includes(b);
              return (
                <button
                  key={b}
                  onClick={() => toggleOwnership(b)}
                  className={`px-2 py-0.5 rounded text-[11px] font-mono transition-colors ${
                    isSelected
                      ? "bg-slate-800 dark:bg-slate-100 text-white dark:text-slate-900 font-semibold"
                      : "bg-slate-200/70 dark:bg-slate-800 text-slate-700 dark:text-slate-300 hover:bg-slate-300 dark:hover:bg-slate-700"
                  }`}
                >
                  {b}
                </button>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
};
