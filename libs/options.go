package libs

// Options global options
type Options struct {
	RootFolder       string
	SignFolder       string
	ResourcesFolder  string
	ThirdPartyFolder string
	ScanID           string
	ConfigFile       string
	FoundCmd         string
	QuietFormat      string
	Output           string
	SummaryOutput    string
	SummaryVuln      string
	LogFile          string
	Proxy            string
	Selectors        string
	InlineDetection  string
	Params           []string
	Headers          []string
	Signs            []string
	//Excludes         []string
	SelectedSigns    []string
	ParallelSigns    []string
	GlobalVar        map[string]string

	Level             int
	Concurrency       int
	Threads           int
	Delay             int
	Timeout           int
	Refresh           int
	Retry             int
	SaveRaw           bool
	JsonOutput        bool
	VerboseSummary    bool
	Quiet             bool
	FullHelp          bool
	Verbose           bool
	Version           bool
	Debug             bool
	NoDB              bool
	NoBackGround      bool
	NoOutput          bool
	EnableFormatInput bool
	DisableParallel   bool

	// Chunk Options
	ChunkDir     string
	ChunkRun     bool
	ChunkThreads int
	ChunkSize    int
	ChunkLimit   int

	Fuzz   Fuzz
	Mics   Mics
	Scan   Scan
	Server Server
	Report Report
	Config Config
}

type Spider struct {
	Domain 		string
	Sub 		bool
	Level		string
	Exclude 	string
	Output		string
	Placeholder	string
	Quiet		bool
}



type Fuzz struct {
	Site       string
	Sites      string
	Proxy      string
	Output     string
	UserAgent  string
	Header     []string
	Threads    int
	Concurrent int
	Depth      int
	Delay      int
	RandomDelay	int
	Timeout		int
	Base		bool
	OtherSource	bool
	IncludeSubs	bool
	IncludeOtherSource	bool
	Verbose		bool
	Quiet		bool
}

// Scan options for api server
type Scan struct {
	RawRequest      string
	EnableGenReport bool
}

// Mics some shortcut options
type Mics struct {
	FullHelp         bool
	AlwaysTrue       bool
	BaseRoot         bool
	BurpProxy        bool
	DisableReplicate bool
}

// Report options for api server
type Report struct {
	VerboseReport bool
	ReportName    string
	TemplateFile  string
	VTemplateFile string
	OutputPath    string
	Title         string
}

// Server options for api server
type Server struct {
	NoAuth       bool
	DBPath       string
	Bind         string
	JWTSecret    string
	Cors         string
	DefaultSign  string
	SecretCollab string
	Username     string
	Password     string
	Key          string
}

// Config options for api server
type Config struct {
	Forced     bool
	SkipMics   bool
	Username   string
	Password   string
	Repo       string
	PrivateKey string
}

// Job define job for running routine
type Job struct {
	URL  string
	Sign Signature
}

// PJob define job for running routine
type PJob struct {
	Req  Request
	ORec Record
	Sign Signature
}

// VulnData vulnerable Data
type VulnData struct {
	ScanID          string
	SignID          string
	SignName        string
	URL             string
	Risk            string
	DetectionString string
	DetectResult    string
	Confidence      string
	Req             string
	Res             string
}
