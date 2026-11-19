import React, { useEffect, useState } from "react";
import { EmptyState } from "./components/EmptyState";
import { GameDetailModal } from "./components/GameDetailModal";
import { GlobalFilterBar } from "./components/GlobalFilterBar";
import { Header } from "./components/Header";
import { ImportModal } from "./components/ImportModal";
import { PageId, Sidebar } from "./components/Sidebar";
import { AudioGapView } from "./features/audio-gaps/AudioGapView";
import { DataQualityView } from "./features/data-quality/DataQualityView";
import { GameExplorerView } from "./features/game-explorer/GameExplorerView";
import { GenreAnalysisView } from "./features/genre-analysis/GenreAnalysisView";
import { LanguageExplorerView } from "./features/language-explorer/LanguageExplorerView";
import { MatrixView } from "./features/localization-matrix/MatrixView";
import { OverviewView } from "./features/overview/OverviewView";
import { OwnershipAnalysisView } from "./features/ownership-analysis/OwnershipAnalysisView";
import { api } from "./lib/api";
import {
  DatasetStatus,
  FilterOptions,
  FilterParams,
  GameSummary,
  OverviewSummary,
} from "./types";

export const App: React.FC = () => {
  const [currentPage, setCurrentPage] = useState<PageId>("overview");
  const [isDark, setIsDark] = useState<boolean>(() => {
    return (
      window.matchMedia &&
      window.matchMedia("(prefers-color-scheme: dark)").matches
    );
  });

  // Filter state
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [filters, setFilters] = useState<FilterParams>({
    excludePlaytests: true,
    requireLanguages: false,
  });
  const [filterOptions, setFilterOptions] = useState<FilterOptions | null>(
    null,
  );

  // Status and dataset state
  const [status, setStatus] = useState<DatasetStatus | null>(null);
  const [overview, setOverview] = useState<OverviewSummary | null>(null);
  const [overviewLoading, setOverviewLoading] = useState(false);

  // Modals
  const [importModalOpen, setImportModalOpen] = useState(false);
  const [selectedGame, setSelectedGame] = useState<GameSummary | null>(null);

  // Sub-navigation targets
  const [targetLangId, setTargetLangId] = useState<string | undefined>();
  const [targetGenreId, setTargetGenreId] = useState<string | undefined>();
  const [gameExplorerParams, setGameExplorerParams] = useState<
    Record<string, any> | undefined
  >();

  // Dark mode effect
  useEffect(() => {
    if (isDark) {
      document.documentElement.classList.add("dark");
    } else {
      document.documentElement.classList.remove("dark");
    }
  }, [isDark]);

  // Load dataset status & filter options
  const fetchStatus = () => {
    api
      .getDatasetStatus()
      .then((s) => setStatus(s))
      .catch(() => {});
  };

  useEffect(() => {
    fetchStatus();
    api
      .getFilterOptions()
      .then((opts) => setFilterOptions(opts))
      .catch(() => {});
  }, []);

  // Poll status during import
  useEffect(() => {
    if (!status?.isImporting) return;
    const interval = setInterval(fetchStatus, 1500);
    return () => clearInterval(interval);
  }, [status?.isImporting]);

  // Load overview data when on overview tab and data is available
  useEffect(() => {
    if (currentPage === "overview" && status?.hasData) {
      setOverviewLoading(true);
      api
        .getOverview(filters)
        .then((data) => setOverview(data))
        .catch(() => {})
        .finally(() => setOverviewLoading(false));
    }
  }, [currentPage, filters, status?.hasData]);

  // Handle navigation with extra params
  const handleNavigate = (page: PageId, extraParams?: Record<string, any>) => {
    if (page === "languages" && extraParams?.languageId) {
      setTargetLangId(extraParams.languageId);
    }
    if (page === "genres" && extraParams?.genreId) {
      setTargetGenreId(extraParams.genreId);
    }
    if (page === "games") {
      setGameExplorerParams(extraParams);
    }
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const handleSelectGame = (appId: number) => {
    api
      .getGame(appId)
      .then((g) => setSelectedGame(g))
      .catch((err) => alert(`Could not load game ${appId}: ${err.message}`));
  };

  // Compute active filter count
  let activeFilterCount = 0;
  if (filters.releaseYearMin) activeFilterCount++;
  if (filters.releaseYearMax) activeFilterCount++;
  if (filters.genres?.length) activeFilterCount += filters.genres.length;
  if (filters.ownership?.length) activeFilterCount += filters.ownership.length;
  if (filters.isFree !== undefined) activeFilterCount++;
  if (filters.platform) activeFilterCount++;
  if (!filters.excludePlaytests) activeFilterCount++; // non-default
  if (filters.requireLanguages) activeFilterCount++;

  return (
    <div className="flex min-h-screen bg-slate-50 dark:bg-slate-950 font-sans">
      {/* Sidebar */}
      <Sidebar
        currentPage={currentPage}
        onSelectPage={(p) => handleNavigate(p)}
        datasetStatus={status}
        onOpenImport={() => setImportModalOpen(true)}
      />

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col min-w-0">
        <Header
          currentPage={currentPage}
          activeFilterCount={activeFilterCount}
          filtersOpen={filtersOpen}
          onToggleFilters={() => setFiltersOpen(!filtersOpen)}
          isDark={isDark}
          onToggleDark={() => setIsDark(!isDark)}
          onOpenImport={() => setImportModalOpen(true)}
        />

        {/* Global Filter Bar */}
        <GlobalFilterBar
          isOpen={filtersOpen}
          onClose={() => setFiltersOpen(false)}
          filters={filters}
          onChangeFilters={setFilters}
          filterOptions={filterOptions}
        />

        {/* Page Content */}
        <main className="flex-1 pb-16">
          {!status?.hasData ? (
            <EmptyState onOpenImport={() => setImportModalOpen(true)} />
          ) : (
            <>
              {currentPage === "overview" && (
                <OverviewView
                  summary={overview}
                  loading={overviewLoading}
                  onNavigate={handleNavigate}
                />
              )}

              {currentPage === "matrix" && (
                <MatrixView filters={filters} onNavigate={handleNavigate} />
              )}

              {currentPage === "languages" && (
                <LanguageExplorerView
                  filters={filters}
                  initialLanguageId={targetLangId}
                  onSelectGame={handleSelectGame}
                />
              )}

              {currentPage === "genres" && (
                <GenreAnalysisView
                  filters={filters}
                  initialGenreId={targetGenreId}
                  onSelectLanguage={(l) =>
                    handleNavigate("languages", { languageId: l })
                  }
                />
              )}

              {currentPage === "audio-gaps" && (
                <AudioGapView
                  filters={filters}
                  onSelectLanguage={(l) =>
                    handleNavigate("languages", { languageId: l })
                  }
                />
              )}

              {currentPage === "ownership" && (
                <OwnershipAnalysisView
                  filters={filters}
                  onSelectLanguage={(l) =>
                    handleNavigate("languages", { languageId: l })
                  }
                />
              )}

              {currentPage === "games" && (
                <GameExplorerView
                  filters={filters}
                  initialParams={gameExplorerParams}
                  onSelectGame={handleSelectGame}
                />
              )}

              {currentPage === "quality" && (
                <DataQualityView onSelectGame={handleSelectGame} />
              )}
            </>
          )}
        </main>
      </div>

      {/* Modals */}
      <ImportModal
        isOpen={importModalOpen}
        onClose={() => setImportModalOpen(false)}
        status={status}
        onTriggerImport={async (source, file) => {
          await api.startImport(source, file);
        }}
        onRefreshStatus={fetchStatus}
      />

      <GameDetailModal
        game={selectedGame}
        onClose={() => setSelectedGame(null)}
      />
    </div>
  );
};
