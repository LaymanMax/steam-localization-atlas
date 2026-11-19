package analytics

// MatrixCell represents the localization coverage of a language within a genre.
type MatrixCell struct {
	Numerator   int64   `json:"numerator"`
	Denominator int64   `json:"denominator"`
	Percentage  float64 `json:"percentage"`
	IsLowSample bool    `json:"isLowSample"`
}

// MatrixGenre represents column header metadata for the matrix.
type MatrixGenre struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	EligibleGames int64  `json:"eligibleGames"`
	TotalGames    int64  `json:"totalGames"`
}

// MatrixLanguageRow represents a single language row in the matrix.
type MatrixLanguageRow struct {
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	TotalTextGames  int64                 `json:"totalTextGames"`
	TotalAudioGames int64                 `json:"totalAudioGames"`
	OverallCoverage float64               `json:"overallCoverage"`
	Cells           map[string]MatrixCell `json:"cells"` // genre_id -> cell
}

// MatrixResponse is the payload returned by the localization matrix endpoint.
type MatrixResponse struct {
	Metric            string              `json:"metric"` // "text", "audio", "conversion"
	Genres            []MatrixGenre       `json:"genres"`
	Languages         []MatrixLanguageRow `json:"languages"`
	MinimumSampleSize int                 `json:"minimumSampleSize"`
	TotalEligible     int64               `json:"totalEligible"`
	TotalDatasetGames int64               `json:"totalDatasetGames"`
}

// OverviewSummary contains top-level descriptive statistics.
type OverviewSummary struct {
	TotalGames            int64                 `json:"totalGames"`
	GamesWithLanguages    int64                 `json:"gamesWithLanguages"`
	GamesWithAudio        int64                 `json:"gamesWithAudio"`
	GamesWithGenres       int64                 `json:"gamesWithGenres"`
	TotalRelationships    int64                 `json:"totalRelationships"`
	UniqueLanguages       int64                 `json:"uniqueLanguages"`
	UniqueGenres          int64                 `json:"uniqueGenres"`
	MedianTextLanguages   float64               `json:"medianTextLanguages"`
	AverageTextLanguages  float64               `json:"averageTextLanguages"`
	MedianAudioLanguages  float64               `json:"medianAudioLanguages"`
	AverageAudioLanguages float64               `json:"averageAudioLanguages"`
	LastImportDate        string                `json:"lastImportDate"`
	TopLanguages          []LanguageRankItem    `json:"topLanguages"`
	LanguageDistribution  []DistributionBin     `json:"languageDistribution"`
	AudioDistribution     []DistributionBin     `json:"audioDistribution"`
	TopGenresCoverage     []GenreRankItem       `json:"topGenresCoverage"`
	LocalizationGaps      []LocalizationGapItem `json:"localizationGaps"`
}

// LanguageRankItem represents a language in ranking lists.
type LanguageRankItem struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	TextGames          int64   `json:"textGames"`
	AudioGames         int64   `json:"audioGames"`
	TextCoverage       float64 `json:"textCoverage"`
	AudioCoverage      float64 `json:"audioCoverage"`
	AudioConversion    float64 `json:"audioConversion"`
	PercentagePointGap float64 `json:"percentagePointGap"`
}

// DistributionBin represents a bucket in a histogram.
type DistributionBin struct {
	BinLabel string  `json:"binLabel"`
	Count    int64   `json:"count"`
	Percent  float64 `json:"percent"`
}

// GenreRankItem represents a genre in ranking lists.
type GenreRankItem struct {
	ID                    string  `json:"id"`
	Name                  string  `json:"name"`
	TotalGames            int64   `json:"totalGames"`
	EligibleGames         int64   `json:"eligibleGames"`
	MedianTextLanguages   float64 `json:"medianTextLanguages"`
	AverageTextLanguages  float64 `json:"averageTextLanguages"`
	AverageAudioLanguages float64 `json:"averageAudioLanguages"`
}

// LocalizationGapItem represents the gap between text and audio support.
type LocalizationGapItem struct {
	LanguageID         string  `json:"languageId"`
	LanguageName       string  `json:"languageName"`
	TextGames          int64   `json:"textGames"`
	AudioGames         int64   `json:"audioGames"`
	EligibleGames      int64   `json:"eligibleGames"`
	TextCoverage       float64 `json:"textCoverage"`
	AudioCoverage      float64 `json:"audioCoverage"`
	PercentagePointGap float64 `json:"percentagePointGap"` // TextCoverage - AudioCoverage
	AudioConversion    float64 `json:"audioConversion"`    // (Audio / Text) * 100
}

// LanguageDetail represents a deep dive for a single language.
type LanguageDetail struct {
	ID                 string                  `json:"id"`
	Name               string                  `json:"name"`
	TextGames          int64                   `json:"textGames"`
	AudioGames         int64                   `json:"audioGames"`
	EligibleGames      int64                   `json:"eligibleGames"`
	TextCoverage       float64                 `json:"textCoverage"`
	AudioCoverage      float64                 `json:"audioCoverage"`
	PercentagePointGap float64                 `json:"percentagePointGap"`
	AudioConversion    float64                 `json:"audioConversion"`
	TopGenres          []GenreCoverageItem     `json:"topGenres"`
	LowestGenres       []GenreCoverageItem     `json:"lowestGenres"`
	OwnershipBreakdown []OwnershipCoverageItem `json:"ownershipBreakdown"`
	YearlyTrends       []YearlyTrendItem       `json:"yearlyTrends"`
}

// GenreCoverageItem represents language coverage in a genre.
type GenreCoverageItem struct {
	GenreID            string  `json:"genreId"`
	GenreName          string  `json:"genreName"`
	TextGames          int64   `json:"textGames"`
	AudioGames         int64   `json:"audioGames"`
	EligibleGames      int64   `json:"eligibleGames"`
	TextCoverage       float64 `json:"textCoverage"`
	AudioCoverage      float64 `json:"audioCoverage"`
	PercentagePointGap float64 `json:"percentagePointGap"`
	AudioConversion    float64 `json:"audioConversion"`
	IsLowSample        bool    `json:"isLowSample"`
}

// OwnershipCoverageItem represents language coverage in an ownership bracket.
type OwnershipCoverageItem struct {
	Bracket            string  `json:"bracket"`
	TotalGames         int64   `json:"totalGames"`
	TextGames          int64   `json:"textGames"`
	AudioGames         int64   `json:"audioGames"`
	TextCoverage       float64 `json:"textCoverage"`
	AudioCoverage      float64 `json:"audioCoverage"`
	PercentagePointGap float64 `json:"percentagePointGap"`
	AudioConversion    float64 `json:"audioConversion"`
}

// YearlyTrendItem represents release year trends.
type YearlyTrendItem struct {
	Year               int     `json:"year"`
	TotalGames         int64   `json:"totalGames"`
	TextGames          int64   `json:"textGames"`
	AudioGames         int64   `json:"audioGames"`
	TextCoverage       float64 `json:"textCoverage"`
	AudioCoverage      float64 `json:"audioCoverage"`
	PercentagePointGap float64 `json:"percentagePointGap"`
}

// GenreDetail represents a deep dive for a genre.
type GenreDetail struct {
	ID                   string                  `json:"id"`
	Name                 string                  `json:"name"`
	TotalGames           int64                   `json:"totalGames"`
	EligibleGames        int64                   `json:"eligibleGames"`
	MedianLanguages      float64                 `json:"medianLanguages"`
	AverageLanguages     float64                 `json:"averageLanguages"`
	TopLanguages         []LanguageRankItem      `json:"topLanguages"`
	LowestLanguages      []LanguageRankItem      `json:"lowestLanguages"`
	LanguageDistribution []DistributionBin       `json:"languageDistribution"`
	OwnershipBreakdown   []OwnershipCoverageItem `json:"ownershipBreakdown"`
	YearlyTrends         []YearlyTrendItem       `json:"yearlyTrends"`
}

// GenreComparison compares two genres with percentage point differences.
type GenreComparison struct {
	GenreA     MatrixGenre        `json:"genreA"`
	GenreB     MatrixGenre        `json:"genreB"`
	Comparison []GenreCompareItem `json:"comparison"`
}

// GenreCompareItem shows language differences between Genre A and Genre B.
type GenreCompareItem struct {
	LanguageID   string  `json:"languageId"`
	LanguageName string  `json:"languageName"`
	CoverageA    float64 `json:"coverageA"`
	CoverageB    float64 `json:"coverageB"`
	DifferencePt float64 `json:"differencePt"` // CoverageA - CoverageB
	GamesA       int64   `json:"gamesA"`
	GamesB       int64   `json:"gamesB"`
}

// GameSummary represents an individual game in lists or details.
type GameSummary struct {
	AppID              int64    `json:"appId"`
	Name               string   `json:"name"`
	ReleaseDate        string   `json:"releaseDate"`
	ReleaseYear        int      `json:"releaseYear"`
	EstimatedOwners    string   `json:"estimatedOwners"`
	Price              float64  `json:"price"`
	IsFree             bool     `json:"isFree"`
	Windows            bool     `json:"windows"`
	Mac                bool     `json:"mac"`
	Linux              bool     `json:"linux"`
	MetacriticScore    int      `json:"metacriticScore"`
	HeaderImage        string   `json:"headerImage"`
	Developers         string   `json:"developers"`
	Publishers         string   `json:"publishers"`
	Genres             []string `json:"genres"`
	TextLanguages      []string `json:"textLanguages"`
	AudioLanguages     []string `json:"audioLanguages"`
	IsPlaytest         bool     `json:"isPlaytest"`
	HasUsableLanguages bool     `json:"hasUsableLanguages"`
	QualityFlags       []string `json:"qualityFlags"`
}

// QualitySummary holds data quality diagnostics.
type QualitySummary struct {
	TotalRecords        int64              `json:"totalRecords"`
	UniqueAppIDs        int64              `json:"uniqueAppIDs"`
	DuplicateAppIDs     int64              `json:"duplicateAppIDs"`
	MissingNames        int64              `json:"missingNames"`
	MissingGenres       int64              `json:"missingGenres"`
	MissingReleaseDates int64              `json:"missingReleaseDates"`
	MissingLanguages    int64              `json:"missingLanguages"`
	SuspectedPlaytests  int64              `json:"suspectedPlaytests"`
	AudioNotInText      int64              `json:"audioNotInText"`
	UsableLanguagePct   float64            `json:"usableLanguagePct"`
	UsableGenrePct      float64            `json:"usableGenrePct"`
	IssueCounts         map[string]int64   `json:"issueCounts"`
	SampleIssues        []QualityIssueItem `json:"sampleIssues"`
}

// QualityIssueItem represents a single logged issue.
type QualityIssueItem struct {
	ID          string `json:"id"`
	IssueType   string `json:"issueType"`
	Severity    string `json:"severity"`
	AppID       int64  `json:"appId"`
	GameName    string `json:"gameName"`
	Description string `json:"description"`
}
