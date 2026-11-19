import { Database, Moon, SlidersHorizontal, Sun } from "lucide-react";
import React from "react";
import { PageId } from "./Sidebar";

interface HeaderProps {
  currentPage: PageId;
  activeFilterCount: number;
  filtersOpen: boolean;
  onToggleFilters: () => void;
  isDark: boolean;
  onToggleDark: () => void;
  onOpenImport: () => void;
}

const titles: Record<PageId, { title: string; subtitle: string }> = {
  overview: {
    title: "Overview Dashboard",
    subtitle:
      "High-level representation and structure of language support across the Steam games catalog.",
  },
  matrix: {
    title: "Localization Coverage Matrix",
    subtitle:
      "Interactive cross-tabulation of languages against game genres with distinct, genre-specific denominators.",
  },
  languages: {
    title: "Language Explorer",
    subtitle:
      "Granular investigation of an individual language: representation rates, genre distribution, and release trends.",
  },
  genres: {
    title: "Genre Analysis",
    subtitle:
      "Localization breadth by genre and direct genre-to-genre percentage-point comparisons.",
  },
  "audio-gaps": {
    title: "Audio Localization Gaps",
    subtitle:
      "Analysis of the discrepancy between listed text/interface support and listed full audio localization.",
  },
  ownership: {
    title: "Estimated Ownership Analysis",
    subtitle:
      "Localization coverage patterns across source-provided approximate ownership brackets.",
  },
  games: {
    title: "Game Catalog Explorer",
    subtitle:
      "Filterable catalog of imported titles with raw metadata and localization status breakdown.",
  },
  quality: {
    title: "Data Quality & Completeness",
    subtitle:
      "Transparent reporting on missing fields, anomalies, playtests, and parsing heuristics.",
  },
};

export const Header: React.FC<HeaderProps> = ({
  currentPage,
  activeFilterCount,
  filtersOpen,
  onToggleFilters,
  isDark,
  onToggleDark,
  onOpenImport,
}) => {
  const current = titles[currentPage] || {
    title: "Steam Localization Gap Atlas",
    subtitle: "",
  };

  return (
    <header className="h-16 border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 px-6 flex items-center justify-between sticky top-0 z-20">
      <div>
        <h2 className="text-base font-semibold text-slate-900 dark:text-white leading-tight">
          {current.title}
        </h2>
        <p className="text-xs text-slate-500 dark:text-slate-400 truncate max-w-xl">
          {current.subtitle}
        </p>
      </div>

      <div className="flex items-center gap-2.5">
        {/* Global filter toggle */}
        <button
          onClick={onToggleFilters}
          className={`flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md border transition-colors ${
            filtersOpen || activeFilterCount > 0
              ? "bg-indigo-50 dark:bg-indigo-950/60 border-indigo-200 dark:border-indigo-800 text-indigo-700 dark:text-indigo-300"
              : "border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800"
          }`}
        >
          <SlidersHorizontal className="w-3.5 h-3.5" />
          <span>Filters</span>
          {activeFilterCount > 0 && (
            <span className="w-4 h-4 rounded-full bg-indigo-600 text-white text-[10px] flex items-center justify-center font-bold">
              {activeFilterCount}
            </span>
          )}
        </button>

        {/* Dataset import modal */}
        <button
          onClick={onOpenImport}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
        >
          <Database className="w-3.5 h-3.5 text-slate-400" />
          <span>Dataset</span>
        </button>

        {/* Theme toggle */}
        <button
          onClick={onToggleDark}
          aria-label="Toggle theme"
          className="p-1.5 rounded-md border border-slate-200 dark:border-slate-700 text-slate-500 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
        >
          {isDark ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
        </button>
      </div>
    </header>
  );
};
