import { Database, Download, Terminal } from "lucide-react";
import React from "react";

interface EmptyStateProps {
  onOpenImport: () => void;
}

export const EmptyState: React.FC<EmptyStateProps> = ({ onOpenImport }) => {
  return (
    <div className="max-w-2xl mx-auto my-16 p-8 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-sm text-center space-y-6">
      <div className="w-14 h-14 rounded-full bg-indigo-50 dark:bg-indigo-950/60 border border-indigo-200 dark:border-indigo-800 flex items-center justify-center text-indigo-600 dark:text-indigo-400 mx-auto">
        <Database className="w-7 h-7" />
      </div>

      <div className="space-y-2">
        <h2 className="text-xl font-bold text-slate-900 dark:text-white">
          Welcome to Steam Localization Gap Atlas
        </h2>
        <p className="text-xs text-slate-500 dark:text-slate-400 max-w-md mx-auto leading-relaxed">
          An open-source descriptive research workstation investigating the
          distribution of listed text/interface and full audio language
          localization across the Steam catalog.
        </p>
      </div>

      <div className="p-4 rounded-lg bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 text-left space-y-3 text-xs">
        <div className="font-semibold text-slate-800 dark:text-slate-200 flex items-center gap-1.5">
          <Terminal className="w-4 h-4 text-indigo-600" />
          <span>Quick Dataset Ingestion Options:</span>
        </div>
        <p className="text-slate-600 dark:text-slate-400 text-[11px]">
          The analytical DuckDB database has not been initialized yet. You can
          import the official Hugging Face dataset (124k games) or provide a
          local Parquet file:
        </p>
        <div className="space-y-1.5 font-mono text-[11px] bg-white dark:bg-slate-900 p-2.5 rounded border border-slate-200 dark:border-slate-800 text-slate-800 dark:text-slate-200">
          <div># Option 1: Ingest via CLI</div>
          <div className="text-indigo-600 dark:text-indigo-400">
            go run ./cmd/atlas ingest --source huggingface
          </div>
          <div className="pt-1.5 text-slate-400">
            # Option 2: Ingest existing local Parquet
          </div>
          <div className="text-indigo-600 dark:text-indigo-400">
            go run ./cmd/atlas ingest --file ./data/source/games.parquet
          </div>
        </div>
      </div>

      <div className="flex items-center justify-center gap-3 pt-2">
        <button
          onClick={onOpenImport}
          className="px-5 py-2 rounded-md bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-semibold shadow-sm transition-colors flex items-center gap-2"
        >
          <Download className="w-4 h-4" />
          <span>Import Dataset in Browser</span>
        </button>
      </div>
    </div>
  );
};
