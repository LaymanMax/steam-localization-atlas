import {
  AlertCircle,
  CheckCircle2,
  Database,
  Download,
  FileText,
  X,
} from "lucide-react";
import React, { useState } from "react";
import { DatasetStatus } from "../types";

interface ImportModalProps {
  isOpen: boolean;
  onClose: () => void;
  status?: DatasetStatus | null;
  onTriggerImport: (source: string, file?: string) => Promise<void>;
  onRefreshStatus: () => void;
}

export const ImportModal: React.FC<ImportModalProps> = ({
  isOpen,
  onClose,
  status,
  onTriggerImport,
  onRefreshStatus,
}) => {
  const [source, setSource] = useState<"huggingface" | "local">("huggingface");
  const [localPath, setLocalPath] = useState("./data/source/games.parquet");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    try {
      await onTriggerImport(source, source === "local" ? localPath : undefined);
      setTimeout(onRefreshStatus, 1000);
    } catch (err: any) {
      setError(err?.message || "Failed to start import");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-sm">
      <div className="w-full max-w-lg rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-xl overflow-hidden text-slate-900 dark:text-slate-100">
        <div className="flex items-center justify-between px-5 py-4 border-b border-slate-200 dark:border-slate-800">
          <div className="flex items-center gap-2">
            <Database className="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
            <h3 className="text-sm font-semibold">Dataset Management</h3>
          </div>
          <button
            onClick={onClose}
            className="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        <div className="p-5 space-y-4">
          {/* Current Status Box */}
          <div className="rounded border border-slate-200 dark:border-slate-800 p-3 bg-slate-50 dark:bg-slate-950 text-xs">
            <div className="flex items-center justify-between mb-1.5 font-medium">
              <span>Database Status:</span>
              <span
                className={
                  status?.hasData
                    ? "text-emerald-600 dark:text-emerald-400 font-semibold flex items-center gap-1"
                    : "text-amber-600 flex items-center gap-1"
                }
              >
                {status?.hasData ? (
                  <CheckCircle2 className="w-3.5 h-3.5" />
                ) : (
                  <AlertCircle className="w-3.5 h-3.5" />
                )}
                {status?.hasData ? "Ready" : "No Data Loaded"}
              </span>
            </div>
            <div className="font-mono text-[11px] text-slate-600 dark:text-slate-400 space-y-0.5">
              <div>
                Total Records:{" "}
                {status?.totalRecords
                  ? status.totalRecords.toLocaleString()
                  : 0}
              </div>
              {status?.lastImport && (
                <div>
                  Last Updated: {new Date(status.lastImport).toLocaleString()}
                </div>
              )}
              {status?.databaseSizeBytes && (
                <div>
                  Storage Size:{" "}
                  {(status.databaseSizeBytes / 1048576).toFixed(1)} MB
                </div>
              )}
            </div>

            {status?.isImporting && (
              <div className="mt-3 pt-2.5 border-t border-slate-200 dark:border-slate-800">
                <div className="flex items-center justify-between text-[11px] font-semibold text-indigo-600 dark:text-indigo-400 mb-1">
                  <span>Import in progress...</span>
                  <span className="animate-pulse">●</span>
                </div>
                <div className="text-[11px] font-mono text-slate-600 dark:text-slate-300">
                  {status.importStatus || "Processing records..."}
                </div>
              </div>
            )}

            {status?.importError && (
              <div className="mt-2 text-[11px] text-rose-600 font-mono">
                Error: {status.importError}
              </div>
            )}
          </div>

          <form onSubmit={handleSubmit} className="space-y-3">
            <div className="space-y-2">
              <label className="text-xs font-semibold text-slate-700 dark:text-slate-300 block">
                Select Ingestion Source:
              </label>

              <div className="grid grid-cols-2 gap-3">
                <label
                  className={`flex items-start gap-2.5 p-3 rounded border text-xs cursor-pointer transition-colors ${
                    source === "huggingface"
                      ? "border-indigo-600 bg-indigo-50/50 dark:bg-indigo-950/40 text-indigo-900 dark:text-indigo-200"
                      : "border-slate-200 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/50"
                  }`}
                >
                  <input
                    type="radio"
                    name="source"
                    checked={source === "huggingface"}
                    onChange={() => setSource("huggingface")}
                    className="mt-0.5"
                  />
                  <div>
                    <div className="font-semibold flex items-center gap-1">
                      <Download className="w-3.5 h-3.5" />
                      Hugging Face
                    </div>
                    <p className="text-[11px] text-slate-500 mt-0.5">
                      FronkonGames public repository (124k records)
                    </p>
                  </div>
                </label>

                <label
                  className={`flex items-start gap-2.5 p-3 rounded border text-xs cursor-pointer transition-colors ${
                    source === "local"
                      ? "border-indigo-600 bg-indigo-50/50 dark:bg-indigo-950/40 text-indigo-900 dark:text-indigo-200"
                      : "border-slate-200 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/50"
                  }`}
                >
                  <input
                    type="radio"
                    name="source"
                    checked={source === "local"}
                    onChange={() => setSource("local")}
                    className="mt-0.5"
                  />
                  <div>
                    <div className="font-semibold flex items-center gap-1">
                      <FileText className="w-3.5 h-3.5" />
                      Local Parquet
                    </div>
                    <p className="text-[11px] text-slate-500 mt-0.5">
                      Load an existing .parquet file from disk
                    </p>
                  </div>
                </label>
              </div>
            </div>

            {source === "local" && (
              <div>
                <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
                  Parquet File Path
                </label>
                <input
                  type="text"
                  value={localPath}
                  onChange={(e) => setLocalPath(e.target.value)}
                  className="w-full px-3 py-1.5 rounded border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs font-mono"
                  placeholder="./data/source/games.parquet"
                />
              </div>
            )}

            {error && (
              <div className="text-xs text-rose-600 bg-rose-50 dark:bg-rose-950/50 p-2.5 rounded border border-rose-200 dark:border-rose-900">
                {error}
              </div>
            )}

            <div className="pt-2 flex items-center justify-end gap-2">
              <button
                type="button"
                onClick={onClose}
                className="px-3 py-1.5 rounded border border-slate-200 dark:border-slate-700 text-xs font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800"
              >
                Close
              </button>
              <button
                type="submit"
                disabled={submitting || status?.isImporting}
                className="px-4 py-1.5 rounded bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-medium shadow-sm disabled:opacity-50 transition-colors"
              >
                {submitting || status?.isImporting
                  ? "Ingesting..."
                  : "Start Dataset Ingestion"}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};
