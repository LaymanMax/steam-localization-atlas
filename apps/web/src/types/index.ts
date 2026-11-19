export interface FilterParams {
  releaseYearMin?: number;
  releaseYearMax?: number;
  genres?: string[];
  ownership?: string[];
  isFree?: boolean;
  platform?: string; // "windows" | "mac" | "linux"
  excludePlaytests: boolean;
  requireLanguages: boolean;
  search?: string;
}

export interface MatrixCell {
  numerator: number;
  denominator: number;
  percentage: number;
  isLowSample: boolean;
}

export interface MatrixGenre {
  id: string;
  name: string;
  eligibleGames: number;
  totalGames: number;
}

export interface MatrixLanguageRow {
  id: string;
  name: string;
  totalTextGames: number;
  totalAudioGames: number;
  overallCoverage: number;
  cells: Record<string, MatrixCell>;
}

export interface MatrixResponse {
  metric: "text" | "audio" | "conversion";
  genres: MatrixGenre[];
  languages: MatrixLanguageRow[];
  minimumSampleSize: number;
  totalEligible: number;
  totalDatasetGames: number;
}

export interface LanguageRankItem {
  id: string;
  name: string;
  textGames: number;
  audioGames: number;
  textCoverage: number;
  audioCoverage: number;
  audioConversion: number;
  percentagePointGap: number;
}

export interface DistributionBin {
  binLabel: string;
  count: number;
  percent: number;
}

export interface GenreRankItem {
  id: string;
  name: string;
  totalGames: number;
  eligibleGames: number;
  medianTextLanguages: number;
  averageTextLanguages: number;
  averageAudioLanguages: number;
}

export interface LocalizationGapItem {
  languageId: string;
  languageName: string;
  textGames: number;
  audioGames: number;
  eligibleGames: number;
  textCoverage: number;
  audioCoverage: number;
  percentagePointGap: number;
  audioConversion: number;
}

export interface OverviewSummary {
  totalGames: number;
  gamesWithLanguages: number;
  gamesWithAudio: number;
  gamesWithGenres: number;
  totalRelationships: number;
  uniqueLanguages: number;
  uniqueGenres: number;
  medianTextLanguages: number;
  averageTextLanguages: number;
  medianAudioLanguages: number;
  averageAudioLanguages: number;
  lastImportDate: string;
  topLanguages: LanguageRankItem[];
  languageDistribution: DistributionBin[];
  audioDistribution: DistributionBin[];
  topGenresCoverage: GenreRankItem[];
  localizationGaps: LocalizationGapItem[];
}

export interface GenreCoverageItem {
  genreId: string;
  genreName: string;
  textGames: number;
  audioGames: number;
  eligibleGames: number;
  textCoverage: number;
  audioCoverage: number;
  percentagePointGap: number;
  audioConversion: number;
  isLowSample: boolean;
}

export interface OwnershipCoverageItem {
  bracket: string;
  totalGames: number;
  textGames: number;
  audioGames: number;
  textCoverage: number;
  audioCoverage: number;
  percentagePointGap: number;
  audioConversion: number;
}

export interface YearlyTrendItem {
  year: number;
  totalGames: number;
  textGames: number;
  audioGames: number;
  textCoverage: number;
  audioCoverage: number;
  percentagePointGap: number;
}

export interface LanguageDetail {
  id: string;
  name: string;
  textGames: number;
  audioGames: number;
  eligibleGames: number;
  textCoverage: number;
  audioCoverage: number;
  percentagePointGap: number;
  audioConversion: number;
  topGenres: GenreCoverageItem[];
  lowestGenres: GenreCoverageItem[];
  ownershipBreakdown: OwnershipCoverageItem[];
  yearlyTrends: YearlyTrendItem[];
}

export interface GenreDetail {
  id: string;
  name: string;
  totalGames: number;
  eligibleGames: number;
  medianLanguages: number;
  averageLanguages: number;
  topLanguages: LanguageRankItem[];
  lowestLanguages: LanguageRankItem[];
  languageDistribution: DistributionBin[];
  ownershipBreakdown: OwnershipCoverageItem[];
  yearlyTrends: YearlyTrendItem[];
}

export interface GenreCompareItem {
  languageId: string;
  languageName: string;
  coverageA: number;
  coverageB: number;
  differencePt: number;
  gamesA: number;
  gamesB: number;
}

export interface GenreComparison {
  genreA: MatrixGenre;
  genreB: MatrixGenre;
  comparison: GenreCompareItem[];
}

export interface GameSummary {
  appId: number;
  name: string;
  releaseDate: string;
  releaseYear: number;
  estimatedOwners: string;
  price: number;
  isFree: boolean;
  windows: boolean;
  mac: boolean;
  linux: boolean;
  metacriticScore: number;
  headerImage: string;
  developers: string;
  publishers: string;
  genres: string[];
  textLanguages: string[];
  audioLanguages: string[];
  isPlaytest: boolean;
  hasUsableLanguages: boolean;
  qualityFlags: string[];
}

export interface QualitySummary {
  totalRecords: number;
  uniqueAppIDs: number;
  duplicateAppIDs: number;
  missingNames: number;
  missingGenres: number;
  missingReleaseDates: number;
  missingLanguages: number;
  suspectedPlaytests: number;
  audioNotInText: number;
  usableLanguagePct: number;
  usableGenrePct: number;
  issueCounts: Record<string, number>;
  sampleIssues: QualityIssueItem[];
}

export interface QualityIssueItem {
  id: string;
  issueType: string;
  severity: string;
  appId: number;
  gameName: string;
  description: string;
}

export interface DatasetStatus {
  hasData: boolean;
  totalRecords: number;
  lastImport: string;
  databasePath: string;
  databaseSizeBytes: number;
  isImporting: boolean;
  importStatus: string;
  importError: string;
}

export interface FilterOptions {
  genres: MatrixGenre[];
  ownershipBrackets: string[];
  releaseYearMin: number;
  releaseYearMax: number;
}
