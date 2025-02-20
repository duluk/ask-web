package config

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"golang.org/x/term"

	"github.com/adrg/xdg"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const Version = "0.0.3"

const MaxTermWidth = 80
const widthPad = 5
const TabWidth = 4

var (
	commit = "Unknown"
	date   = "Unknown"
)

type Opts struct {
	ConfigDir   string
	DumpConfig  bool
	ShowAPIKeys bool

	Model         string
	ContextLength int
	Temperature   float64

	LogFileName string
	LogStderr   bool
	DBFileName  string
	DBTable     string

	QueryPrompt   string
	SummaryPrompt string

	Search string
	Show   int

	FilteredURLs []string

	NumResults int
	MaxTokens  int

	ScreenWidth  int
	ScreenHeight int
	TabWidth     int
}

func Initialize() (*Opts, error) {
	configDir, err := xdg.ConfigFile("ask-web")
	if err != nil {
		log.Fatal("Error getting config directory:", err)
	}

	// TODO: though I've put so much effort into the config file to read it
	// first so that the values can be used as defaults (eg in --help), I'm
	// starting to wonder if that's even what I want...
	// Figure out the config file and read it first if there
	pflag.StringP("config", "C", "", "Configuration file")
	viper.BindPFlags(pflag.CommandLine)
	if err := setupConfigFile(); err != nil {
		return nil, fmt.Errorf("error setting up config: %w", err)
	}

	defaultLogFileName := filepath.Join(configDir, "ask-web.log")

	width, height := determineScreenSize()

	viper.SetDefault("model.default", "claude")
	viper.SetDefault("model.context_length", 2048)
	viper.SetDefault("model.max_tokens", 420)
	viper.SetDefault("model.num_results", 3)
	viper.SetDefault("model.temperature", 0.7)
	viper.SetDefault("model.query_prompt", "Turn this prompt into a search query, ensuring to retain its meaning")
	viper.SetDefault("model.summary_prompt", "Please provide a detailed summary of the following text that is directly related to the query")
	viper.SetDefault("logging.file", defaultLogFileName)
	viper.SetDefault("database.file", filepath.Join(configDir, "ask-web.db"))
	viper.SetDefault("database.table", "conversations")
	viper.SetDefault("screen.width", width)
	viper.SetDefault("screen.height", height)
	viper.SetDefault("filter", []string{"wikipedia.org", "britannica.com"})

	// Now define the rest of the flags using values from viper (which now has
	// config file values)
	pflag.StringP("model", "m", viper.GetString("model.default"), "Which LLM to use for summary (chatgpt|deepseek)")
	pflag.IntP("num-results", "n", viper.GetInt("model.num_results"), "How many web pages to check per search engine")
	pflag.IntP("context-length", "l", viper.GetInt("model.context_length"), "Maximum context length")
	pflag.StringP("database", "d", viper.GetString("database.file"), "Database file")
	pflag.StringP("query-prompt", "q", viper.GetString("model.query_prompt"), "Prompt for generating search query from prompt")
	pflag.StringP("summary-prompt", "S", viper.GetString("model.summary_prompt"), "System prompt for LLM")
	pflag.IntP("max-tokens", "t", viper.GetInt("model.max_tokens"), "Maximum tokens to generate")
	pflag.Float64P("temperature", "T", viper.GetFloat64("model.temperature"), "Temperature for summarization")
	pflag.BoolP("version", "v", false, "Print version and exit")
	pflag.BoolP("full-version", "V", false, "Print full version information and exit")
	pflag.BoolP("dump-config", "", false, "Dump configuration and exit")
	pflag.BoolP("stderr", "", false, "Log to stderr in addition to file")
	pflag.StringP("search", "s", "", "Search for a response")
	pflag.IntP("show", "", 0, "Show response with ID")
	pflag.BoolP("show-keys", "", false, "Show API keys")
	pflag.StringP("width", "", "", "Width of the screen for linewrap")
	pflag.StringP("height", "", "", "Height of the screen for linewrap")

	// Bind all flags to viper
	viper.BindPFlag("model.num_results", pflag.Lookup("context"))
	viper.BindPFlag("dump-config", pflag.Lookup("dump-config"))
	viper.BindPFlag("stderr", pflag.Lookup("stderr"))
	viper.BindPFlag("show-keys", pflag.Lookup("show-keys"))
	viper.BindPFlag("search", pflag.Lookup("search"))
	viper.BindPFlag("show", pflag.Lookup("show"))
	viper.BindPFlag("database.file", pflag.Lookup("database"))
	viper.BindPFlag("model.system_prompt", pflag.Lookup("system-prompt"))
	viper.BindPFlag("model.max_tokens", pflag.Lookup("max-tokens"))
	viper.BindPFlag("model.context_length", pflag.Lookup("context-length"))
	viper.BindPFlag("model.temperature", pflag.Lookup("temperature"))
	viper.BindPFlag("screen.width", pflag.Lookup("width"))
	viper.BindPFlag("screen.height", pflag.Lookup("height"))

	viper.BindPFlag("version", pflag.Lookup("version"))
	viper.BindPFlag("full-version", pflag.Lookup("full-version"))

	viper.Set("logging.file", os.ExpandEnv(viper.GetString("logging.file")))
	viper.Set("database.file", os.ExpandEnv(viper.GetString("database.file")))

	pflag.Parse()

	if handleVersionFlags() {
		os.Exit(0)
	}

	return &Opts{
		ConfigDir:     configDir,
		DumpConfig:    viper.GetBool("dump-config"),
		ShowAPIKeys:   viper.GetBool("show-keys"),
		Model:         pflag.Lookup("model").Value.String(),
		ContextLength: viper.GetInt("model.context_length"),
		Temperature:   viper.GetFloat64("model.temperature"),
		FilteredURLs:  viper.GetStringSlice("filter"),
		LogFileName:   viper.GetString("logging.file"),
		LogStderr:     viper.GetBool("stderr"),
		DBFileName:    os.ExpandEnv(viper.GetString("database.file")),
		DBTable:       viper.GetString("database.table"),
		QueryPrompt:   viper.GetString("model.query_prompt"),
		SummaryPrompt: viper.GetString("model.summary_prompt"),
		Search:        viper.GetString("search"),
		Show:          viper.GetInt("show"),
		NumResults:    viper.GetInt("model.num_results"),
		MaxTokens:     viper.GetInt("model.max_tokens"),
		ScreenWidth:   min(viper.GetInt("screen.width"), MaxTermWidth) - widthPad,
		ScreenHeight:  viper.GetInt("screen.height"),
		TabWidth:      TabWidth,
	}, nil
}

func determineScreenSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		width = 80
		height = 24
	}

	return width, height
}

func showResponse(respID int) {
	// if viper.GetString("database.file") == "" {
	// 	fmt.Println("Database file not set")
	// 	os.Exit(1)
	// }
	// if viper.GetString("database.table") == "" {
	// 	fmt.Println("Database table not set")
	// 	os.Exit(1)
	// }
	//
	// db, err := database.InitializeDB(os.ExpandEnv(viper.GetString("database.file")), viper.GetString("database.table"))
	// if err != nil {
	// 	fmt.Printf("Error opening database: %s", err)
	// 	os.Exit(1)
	// }
	// defer db.Close()
	//
	// db.ShowConversation(viper.GetInt("show"))

	fmt.Printf("Not Implemented: would have searched for %d\n", respID)
	os.Exit(0)
}

func handleVersionFlags() bool {
	if viper.GetBool("version") {
		fmt.Println("ask-web version:", Version)
		return true
	}
	if viper.GetBool("full-version") {
		fmt.Printf("Version: %s\nCommit:  %s\nDate:    %s\n", Version, commit, date)
		return true
	}
	return false
}

// This is pretty bad but it's a quick hack to get the config file because
// nothing else is working. I need to parse and read the config file before
// the rest of the options so that I can use the values from the config file
// to set the defaults for the other options.
func checkConfigFlag() string {
	for i, arg := range os.Args {
		if arg == "--config" || arg == "-C" {
			if i+1 < len(os.Args) {
				return os.Args[i+1]
			}
		}
		if len(arg) > 8 && arg[:8] == "--config=" {
			return arg[8:]
		}
	}
	return ""
}

// If there's an error getting the user, just returning the path unmodified
func expandHomePath(path string) string {
	if strings.HasPrefix(path, "~") {
		currentUser, err := user.Current()
		if err != nil {
			// I'm always trying to sneak a goto in just to trigger :)
			goto oopsies
		}
		return filepath.Join(currentUser.HomeDir, path[1:])
	}
oopsies:
	return path
}

func setupConfigFile() error {
	cfgFile := checkConfigFlag()

	if cfgFile != "" {
		viper.SetConfigFile(expandHomePath(cfgFile))
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yml")
		viper.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".config", "ask-web"))
		viper.AddConfigPath(".")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}
	return nil
}

func DumpConfig(cfg *Opts) {
	fmt.Printf("Model: %s\n", cfg.Model)
	fmt.Printf("MaxTokens: %d\n", cfg.MaxTokens)
	fmt.Printf("NumResults: %d\n", cfg.NumResults)
	fmt.Printf("ContextLength: %d\n", cfg.ContextLength)
	fmt.Printf("Temperature: %f\n", cfg.Temperature)
	fmt.Printf("DBFileName: %s\n", cfg.DBFileName)
	fmt.Printf("DBTable: %s\n", cfg.DBTable)
	fmt.Printf("LogFileName: %s\n", cfg.LogFileName)
	fmt.Printf("SummaryPrompt: %s\n", cfg.SummaryPrompt)
	fmt.Printf("ScreenWidth: %d\n", cfg.ScreenWidth)
	fmt.Printf("ScreenHeight: %d\n", cfg.ScreenHeight)
	fmt.Printf("TabWidth: %d\n", cfg.TabWidth)
}
