import {
  AlertCircle,
  CheckCircle,
  ExternalLink,
  FileText,
  Shield,
  Volume2,
  X,
} from "lucide-react";
import React from "react";
import { GameSummary } from "../types";

interface GameDetailModalProps {
  game: GameSummary | null;
  onClose: () => void;
}

export const GameDetailModal: React.FC<GameDetailModalProps> = ({
  game,
  onClose,
}) => {
  if (!game) return null;

  const audioSet = new Set(game.audioLanguages || []);
  const textSet = new Set(game.textLanguages || []);

  const bothLanguages = (game.textLanguages || []).filter((l) =>
    audioSet.has(l),
  );
  const textOnlyLanguages = (game.textLanguages || []).filter(
    (l) => !audioSet.has(l),
  );
  const audioOnlyLanguages = (game.audioLanguages || []).filter(
    (l) => !textSet.has(l),
  );

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm animate-in fade-in">
      <div className="w-full max-w-2xl max-h-[90vh] rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-2xl flex flex-col overflow-hidden text-slate-900 dark:text-slate-100">
        {/* Header Image & Title */}
        <div className="relative bg-slate-950 h-44 overflow-hidden shrink-0 border-b border-slate-200 dark:border-slate-800">
          {game.headerImage ? (
            <img
              src={game.headerImage}
              alt={game.name}
              className="w-full h-full object-cover object-center opacity-85"
              onError={(e) => {
                (e.target as HTMLElement).style.display = "none";
              }}
            />
          ) : (
            <div className="w-full h-full flex items-center justify-center text-slate-600 font-mono text-sm">
              [No Header Image Available]
            </div>
          )}

          <div className="absolute inset-0 bg-gradient-to-t from-slate-950 via-slate-950/40 to-transparent" />

          <button
            onClick={onClose}
            className="absolute top-3 right-3 p-1.5 rounded-full bg-black/60 text-white/80 hover:text-white hover:bg-black/90 transition-colors"
          >
            <X className="w-4 h-4" />
          </button>

          <div className="absolute bottom-3.5 left-5 right-5">
            <div className="flex items-center gap-2 mb-1">
              <span className="px-2 py-0.5 rounded bg-indigo-600/90 text-white text-[10px] font-mono font-medium">
                App ID: {game.appId}
              </span>
              {game.isPlaytest && (
                <span className="px-2 py-0.5 rounded bg-amber-500/90 text-black text-[10px] font-semibold">
                  Playtest Title
                </span>
              )}
            </div>
            <h2 className="text-xl font-bold text-white tracking-tight leading-snug drop-shadow-sm">
              {game.name}
            </h2>
          </div>
        </div>

        {/* Scrollable Content */}
        <div className="flex-1 overflow-y-auto p-5 space-y-5 text-xs">
          {/* Metadata Grid */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3 p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 font-mono text-[11px]">
            <div>
              <span className="text-slate-400 block mb-0.5">Release Date:</span>
              <span className="font-semibold text-slate-800 dark:text-slate-200">
                {game.releaseDate || "Unknown"}
              </span>
            </div>
            <div>
              <span className="text-slate-400 block mb-0.5">Est. Owners:</span>
              <span className="font-semibold text-slate-800 dark:text-slate-200">
                {game.estimatedOwners || "Unknown"}
              </span>
            </div>
            <div>
              <span className="text-slate-400 block mb-0.5">Price:</span>
              <span className="font-semibold text-slate-800 dark:text-slate-200">
                {game.isFree ? "Free to Play" : `$${game.price.toFixed(2)}`}
              </span>
            </div>
            <div>
              <span className="text-slate-400 block mb-0.5">Platforms:</span>
              <span className="font-semibold text-slate-800 dark:text-slate-200">
                {[
                  game.windows && "Win",
                  game.mac && "Mac",
                  game.linux && "Linux",
                ]
                  .filter(Boolean)
                  .join(" / ") || "None"}
              </span>
            </div>
          </div>

          {/* Genres & Credits */}
          <div className="space-y-2">
            <div>
              <span className="text-slate-400 text-[11px] font-medium block mb-1">
                Genres:
              </span>
              <div className="flex flex-wrap gap-1.5">
                {game.genres && game.genres.length > 0 ? (
                  game.genres.map((g) => (
                    <span
                      key={g}
                      className="px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 font-medium text-[11px]"
                    >
                      {g}
                    </span>
                  ))
                ) : (
                  <span className="text-slate-400 italic">
                    No genres listed in source
                  </span>
                )}
              </div>
            </div>

            {(game.developers || game.publishers) && (
              <div className="grid grid-cols-2 gap-3 pt-2 text-[11px] text-slate-600 dark:text-slate-400">
                <div>
                  <span className="text-slate-400">Developer:</span>{" "}
                  {game.developers || "N/A"}
                </div>
                <div>
                  <span className="text-slate-400">Publisher:</span>{" "}
                  {game.publishers || "N/A"}
                </div>
              </div>
            )}
          </div>

          {/* Language Breakdown */}
          <div className="space-y-3 pt-1">
            <h4 className="font-semibold text-xs text-slate-800 dark:text-slate-200 uppercase tracking-wider border-b border-slate-200 dark:border-slate-800 pb-1.5 flex items-center justify-between">
              <span>Language Support Breakdown</span>
              <span className="text-[11px] font-normal normal-case text-slate-500">
                {game.textLanguages?.length || 0} Text ·{" "}
                {game.audioLanguages?.length || 0} Full Audio
              </span>
            </h4>

            {/* Both Text & Audio */}
            {bothLanguages.length > 0 && (
              <div>
                <div className="flex items-center gap-1.5 text-[11px] font-semibold text-emerald-600 dark:text-emerald-400 mb-1.5">
                  <CheckCircle className="w-3.5 h-3.5" />
                  <span>
                    Full Audio & Text/Interface Listed ({bothLanguages.length})
                  </span>
                </div>
                <div className="flex flex-wrap gap-1.5">
                  {bothLanguages.map((l) => (
                    <span
                      key={l}
                      className="px-2 py-0.5 rounded bg-emerald-50 dark:bg-emerald-950/60 border border-emerald-200 dark:border-emerald-800 text-emerald-800 dark:text-emerald-200 font-medium text-[11px] flex items-center gap-1"
                    >
                      <Volume2 className="w-3 h-3 text-emerald-600" />
                      {l}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {/* Text Only */}
            {textOnlyLanguages.length > 0 && (
              <div>
                <div className="flex items-center gap-1.5 text-[11px] font-semibold text-sky-600 dark:text-sky-400 mb-1.5">
                  <FileText className="w-3.5 h-3.5" />
                  <span>Text/Interface Only ({textOnlyLanguages.length})</span>
                </div>
                <div className="flex flex-wrap gap-1.5">
                  {textOnlyLanguages.map((l) => (
                    <span
                      key={l}
                      className="px-2 py-0.5 rounded bg-sky-50 dark:bg-sky-950/60 border border-sky-200 dark:border-sky-800 text-sky-800 dark:text-sky-200 font-medium text-[11px]"
                    >
                      {l}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {/* Audio without text */}
            {audioOnlyLanguages.length > 0 && (
              <div>
                <div className="flex items-center gap-1.5 text-[11px] font-semibold text-amber-600 dark:text-amber-400 mb-1.5">
                  <AlertCircle className="w-3.5 h-3.5" />
                  <span>
                    Audio Listed without Text Support (
                    {audioOnlyLanguages.length})
                  </span>
                </div>
                <div className="flex flex-wrap gap-1.5">
                  {audioOnlyLanguages.map((l) => (
                    <span
                      key={l}
                      className="px-2 py-0.5 rounded bg-amber-50 dark:bg-amber-950/60 border border-amber-200 dark:border-amber-800 text-amber-800 dark:text-amber-200 font-medium text-[11px]"
                    >
                      {l}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {!game.textLanguages?.length && !game.audioLanguages?.length && (
              <div className="p-3 rounded border border-dashed border-slate-300 dark:border-slate-700 text-slate-500 text-center italic">
                No language metadata listed in source dataset
              </div>
            )}
          </div>

          {/* Quality flags */}
          {game.qualityFlags && game.qualityFlags.length > 0 && (
            <div className="pt-2 border-t border-slate-200 dark:border-slate-800">
              <span className="text-[11px] font-medium text-slate-400 flex items-center gap-1 mb-1">
                <Shield className="w-3.5 h-3.5 text-slate-400" />
                Data Quality Flags:
              </span>
              <div className="flex flex-wrap gap-1.5">
                {game.qualityFlags.map((flag) => (
                  <span
                    key={flag}
                    className="px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 font-mono text-[10px]"
                  >
                    {flag}
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="px-5 py-3 border-t border-slate-200 dark:border-slate-800 bg-slate-50/60 dark:bg-slate-950/60 flex items-center justify-between">
          <a
            href={`https://store.steampowered.com/app/${game.appId}`}
            target="_blank"
            rel="noreferrer"
            className="flex items-center gap-1.5 text-xs font-medium text-indigo-600 dark:text-indigo-400 hover:underline"
          >
            <span>View on Steam Store</span>
            <ExternalLink className="w-3 h-3" />
          </a>

          <button
            onClick={onClose}
            className="px-4 py-1.5 rounded border border-slate-200 dark:border-slate-700 text-xs font-medium hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
};
