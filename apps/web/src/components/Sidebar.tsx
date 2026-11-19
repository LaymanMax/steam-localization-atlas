import {
  Compass,
  Database,
  ExternalLink,
  Grid3X3,
  Languages,
  LayoutDashboard,
  ShieldAlert,
  Tags,
  Users,
  VolumeX,
} from "lucide-react";
import React from "react";
import { DatasetStatus } from "../types";

export type PageId =
  | "overview"
  | "matrix"
  | "languages"
  | "genres"
  | "audio-gaps"
  | "ownership"
  | "games"
  | "quality";

interface SidebarProps {
  currentPage: PageId;
  onSelectPage: (p: PageId) => void;
  datasetStatus?: DatasetStatus | null;
  onOpenImport: () => void;
}

export const Sidebar: React.FC<SidebarProps> = ({
  currentPage,
  onSelectPage,
  datasetStatus,
  onOpenImport,
}) => {
  const navItems = [
    { id: "overview" as PageId, label: "Overview", icon: LayoutDashboard },
    {
      id: "matrix" as PageId,
      label: "Localization Matrix",
      icon: Grid3X3,
      badge: "Flagship",
    },
    { id: "languages" as PageId, label: "Language Explorer", icon: Languages },
    { id: "genres" as PageId, label: "Genre Analysis", icon: Tags },
    {
      id: "audio-gaps" as PageId,
      label: "Audio Localization Gaps",
      icon: VolumeX,
    },
    { id: "ownership" as PageId, label: "Ownership Analysis", icon: Users },
    { id: "games" as PageId, label: "Game Explorer", icon: Compass },
    { id: "quality" as PageId, label: "Data Quality", icon: ShieldAlert },
  ];

  return (
    <aside className="w-64 border-r border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 flex flex-col h-screen shrink-0 sticky top-0">
      {/* Brand */}
      <div className="p-5 border-b border-slate-200 dark:border-slate-800">
        <div className="flex items-center gap-2.5">
          <div className="w-8 h-8 rounded bg-indigo-600 flex items-center justify-center text-white font-bold text-sm tracking-wider shadow-sm">
            SL
          </div>
          <div>
            <h1 className="font-semibold text-sm leading-tight tracking-tight text-slate-900 dark:text-white">
              Steam Localization
            </h1>
            <p className="text-xs text-slate-500 dark:text-slate-400 font-mono">
              Gap Atlas v1.0
            </p>
          </div>
        </div>
      </div>

      {/* Nav List */}
      <nav className="flex-1 p-3 space-y-1 overflow-y-auto">
        <div className="px-3 py-1.5 text-[10px] font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500">
          Research Workstation
        </div>
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = currentPage === item.id;
          return (
            <button
              key={item.id}
              onClick={() => onSelectPage(item.id)}
              className={`w-full flex items-center justify-between px-3 py-2 text-xs font-medium rounded-md transition-colors ${
                isActive
                  ? "bg-indigo-50 dark:bg-indigo-950/60 text-indigo-700 dark:text-indigo-300 font-semibold"
                  : "text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800/60 hover:text-slate-900 dark:hover:text-slate-200"
              }`}
            >
              <div className="flex items-center gap-2.5">
                <Icon
                  className={`w-4 h-4 ${isActive ? "text-indigo-600 dark:text-indigo-400" : "text-slate-400"}`}
                />
                <span>{item.label}</span>
              </div>
              {item.badge && (
                <span className="px-1.5 py-0.5 text-[9px] font-medium tracking-wide uppercase bg-indigo-100 dark:bg-indigo-900/70 text-indigo-700 dark:text-indigo-300 rounded">
                  {item.badge}
                </span>
              )}
            </button>
          );
        })}
      </nav>

      {/* Dataset status footer */}
      <div className="p-3 border-t border-slate-200 dark:border-slate-800 bg-slate-50/60 dark:bg-slate-950/40">
        <div className="rounded-md border border-slate-200 dark:border-slate-800 p-2.5 bg-white dark:bg-slate-900 text-xs">
          <div className="flex items-center justify-between mb-1.5">
            <span className="text-[11px] font-medium text-slate-500 dark:text-slate-400 flex items-center gap-1.5">
              <Database className="w-3.5 h-3.5 text-slate-400" />
              Catalog Status
            </span>
            <span
              className={`inline-block w-2 h-2 rounded-full ${
                datasetStatus?.hasData ? "bg-emerald-500" : "bg-amber-500"
              }`}
            />
          </div>

          <div className="text-[11px] font-mono text-slate-700 dark:text-slate-300 mb-2">
            {datasetStatus?.hasData ? (
              <>
                <span className="font-semibold text-slate-900 dark:text-white">
                  {datasetStatus.totalRecords.toLocaleString()}
                </span>{" "}
                records loaded
              </>
            ) : (
              <span className="text-amber-600 dark:text-amber-400">
                Empty Database
              </span>
            )}
          </div>

          <button
            onClick={onOpenImport}
            className="w-full py-1 px-2 text-[11px] font-medium rounded border border-slate-200 dark:border-slate-700 hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300 transition-colors"
          >
            {datasetStatus?.isImporting ? "Importing..." : "Manage Dataset"}
          </button>
        </div>

        <div className="mt-2.5 px-1 flex items-center justify-between text-[10px] text-slate-400 dark:text-slate-500">
          <span>MIT Licensed Data</span>
          <a
            href="https://huggingface.co/datasets/FronkonGames/steam-games-dataset"
            target="_blank"
            rel="noreferrer"
            className="hover:underline flex items-center gap-0.5"
          >
            Upstream <ExternalLink className="w-2.5 h-2.5" />
          </a>
        </div>
      </div>
    </aside>
  );
};
