package config

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"

	"github.com/gochan-org/gochan/pkg/gcutil"
)

const (
	randomStringSize = 16
	cookieMaxAgeEx   = ` (example: "1 year 2 months 3 days 4 hours", or "1y2mo3d4h"`
	/* currentConfig = iota
	oldConfig
	invalidConfig */
)

var (
	cfg      *GochanConfig
	cfgPath  string
	defaults = map[string]any{
		"WebRoot": "/",
		// SiteConfig
		"FirstPage":       []string{"index.html", "firstrun.html", "1.html"},
		"CookieMaxAge":    "1y",
		"LockdownMessage": "This imageboard has temporarily disabled posting. We apologize for the inconvenience",
		"SiteName":        "Gochan",
		"MinifyHTML":      true,
		"MinifyJS":        true,
		"MaxRecentPosts":  12,
		"EnableAppeals":   true,
		"MaxLogDays":      14,

		// BoardConfig
		"DateTimeFormat": "Mon, January 02, 2006 3:04:05 PM",

		// PostConfig
		"NewThreadDelay":           30,
		"ReplyDelay":               7,
		"MaxLineLength":            150,
		"ThreadsPerPage":           15,
		"RepliesOnBoardPage":       3,
		"StickyRepliesOnBoardPage": 1,
		"BanMessage":               "USER WAS BANNED FOR THIS POST",
		"EmbedWidth":               200,
		"EmbedHeight":              164,
		"EnableEmbeds":             true,
		"ImagesOpenNewTab":         true,
		"NewTabOnOutlinks":         true,

		// UploadConfig
		"ThumbWidth":         200,
		"ThumbHeight":        200,
		"ThumbWidthReply":    125,
		"ThumbHeightReply":   125,
		"ThumbWidthCatalog":  50,
		"ThumbHeightCatalog": 50,
	}

	boardConfigs    = map[string]BoardConfig{}
	acceptedDrivers = []string{"mysql", "postgres", "sqlite3"}
)

type GochanConfig struct {
	SystemCriticalConfig
	SiteConfig
	BoardConfig
	BoardListConfig
	jsonLocation string `json:"-"`
	testing      bool
}

func (gcfg *GochanConfig) setField(field string, value interface{}) {
	structValue := reflect.ValueOf(gcfg).Elem()
	structFieldValue := structValue.FieldByName(field)
	if !structFieldValue.IsValid() {
		return
	}
	if !structFieldValue.CanSet() {
		return
	}
	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		return
	}

	structFieldValue.Set(val)
}

// ValidateValues checks to make sure that the configuration options are usable
// (e.g., ListenIP is a valid IP address, Port isn't a negative number, etc)
func (gcfg *GochanConfig) ValidateValues() error {
	if net.ParseIP(gcfg.ListenIP) == nil {
		return &InvalidValueError{Field: "ListenIP", Value: gcfg.ListenIP}
	}
	changed := false

	if gcfg.WebRoot == "" {
		gcfg.WebRoot = "/"
		changed = true
	}
	if len(gcfg.FirstPage) == 0 {
		gcfg.FirstPage = defaults["FirstPage"].([]string)
		changed = true
	}
	if gcfg.CookieMaxAge == "" {
		gcfg.CookieMaxAge = defaults["CookieMaxAge"].(string)
		changed = true
	}
	_, err := gcutil.ParseDurationString(gcfg.CookieMaxAge)
	if err == gcutil.ErrInvalidDurationString {
		return &InvalidValueError{Field: "CookieMaxAge", Value: gcfg.CookieMaxAge, Details: err.Error() + cookieMaxAgeEx}
	} else if err != nil {
		return err
	}

	if gcfg.LockdownMessage == "" {
		gcfg.LockdownMessage = defaults["LockdownMessage"].(string)
	}
	if gcfg.DBtype == "postgresql" {
		gcfg.DBtype = "postgres"
	}
	found := false
	for _, driver := range acceptedDrivers {
		if gcfg.DBtype == driver {
			found = true
			break
		}
	}

	if !found {
		return &InvalidValueError{
			Field:   "DBtype",
			Value:   gcfg.DBtype,
			Details: "currently supported values: " + strings.Join(acceptedDrivers, ",")}
	}
	if len(gcfg.Styles) == 0 {
		return &InvalidValueError{Field: "Styles", Value: gcfg.Styles}
	}
	if gcfg.DefaultStyle == "" {
		gcfg.DefaultStyle = gcfg.Styles[0].Filename
		changed = true
	}

	if gcfg.SiteName == "" {
		gcfg.SiteName = defaults["SiteName"].(string)
	}

	if gcfg.MaxLineLength == 0 {
		gcfg.MaxLineLength = defaults["MaxLineLength"].(int)
		changed = true
	}
	if gcfg.ThumbWidth == 0 {
		gcfg.ThumbWidth = defaults["ThumbWidth"].(int)
		changed = true
	}
	if gcfg.ThumbHeight == 0 {
		gcfg.ThumbHeight = defaults["ThumbHeight"].(int)
		changed = true
	}
	if gcfg.ThumbWidthReply == 0 {
		gcfg.ThumbWidthReply = defaults["ThumbWidthReply"].(int)
		changed = true
	}
	if gcfg.ThumbHeightReply == 0 {
		gcfg.ThumbHeightReply = defaults["ThumbHeightReply"].(int)
		changed = true
	}

	if gcfg.ThumbWidthCatalog == 0 {
		gcfg.ThumbWidthCatalog = defaults["ThumbWidthCatalog"].(int)
		changed = true
	}
	if gcfg.ThumbHeightCatalog == 0 {
		gcfg.ThumbHeightCatalog = defaults["ThumbHeightCatalog"].(int)
		changed = true
	}
	if gcfg.ThreadsPerPage == 0 {
		gcfg.ThreadsPerPage = defaults["ThreadsPerPage"].(int)
		changed = true
	}
	if gcfg.RepliesOnBoardPage == 0 {
		gcfg.RepliesOnBoardPage = defaults["RepliesOnBoardPage"].(int)
		changed = true
	}
	if gcfg.StickyRepliesOnBoardPage == 0 {
		gcfg.StickyRepliesOnBoardPage = defaults["StickyRepliesOnBoardPage"].(int)
		changed = true
	}
	if gcfg.BanMessage == "" {
		gcfg.BanMessage = defaults["BanMessage"].(string)
		changed = true
	}
	if gcfg.DateTimeFormat == "" {
		gcfg.DateTimeFormat = defaults["DateTimeFormat"].(string)
		changed = true
	}

	if gcfg.EnableGeoIP {
		if gcfg.GeoIPDBlocation == "" {
			return &InvalidValueError{Field: "GeoIPDBlocation", Value: "", Details: "GeoIPDBlocation must be set in gochan.json if EnableGeoIP is true"}
		}
	}

	if gcfg.MaxLogDays == 0 {
		gcfg.MaxLogDays = defaults["MaxLogDays"].(int)
		changed = true
	}

	if gcfg.RandomSeed == "" {
		gcfg.RandomSeed = gcutil.RandomString(randomStringSize)
		changed = true
	}

	if gcfg.StripImageMetadata == "exif" || gcfg.StripImageMetadata == "all" {
		if gcfg.ExiftoolPath == "" {
			if gcfg.ExiftoolPath, err = exec.LookPath("exiftool"); err != nil {
				return &InvalidValueError{
					Field: "ExiftoolPath", Value: "", Details: "unable to find exiftool in the system path",
				}
			}
		} else {
			if _, err = exec.LookPath(gcfg.ExiftoolPath); err != nil {
				return &InvalidValueError{
					Field: "ExiftoolPath", Value: gcfg.ExiftoolPath, Details: "unable to find exiftool at the given location",
				}
			}
		}
	} else if gcfg.StripImageMetadata != "" && gcfg.StripImageMetadata != "none" {
		return &InvalidValueError{
			Field:   "StripImageMetadata",
			Value:   gcfg.StripImageMetadata,
			Details: `valid values are "","none","exif", or "all"`,
		}
	}

	if !changed {
		return nil
	}
	return gcfg.Write()
}

func (gcfg *GochanConfig) Write() error {
	str, err := json.MarshalIndent(gcfg, "", "\t")
	if err != nil {
		return err
	}
	if gcfg.testing {
		// don't try to write anything if we're doing a test
		return nil
	}
	return os.WriteFile(gcfg.jsonLocation, str, GC_FILE_MODE)
}

/*
SystemCriticalConfig contains configuration options that are extremely important, and fucking with them while
the server is running could have site breaking consequences. It should only be changed by modifying the configuration
file and restarting the server.
*/
type SystemCriticalConfig struct {
	ListenIP     string `critical:"true"`
	Port         int    `critical:"true"`
	UseFastCGI   bool   `critical:"true"`
	DocumentRoot string `critical:"true"`
	TemplateDir  string `critical:"true"`
	LogDir       string `critical:"true"`
	Plugins      []string

	SiteHeaderURL string
	WebRoot       string `description:"The HTTP root appearing in the browser (e.g. '/', 'https://yoursite.net/', etc) that all internal links start with"`
	SiteDomain    string `description:"The server's domain (e.g. gochan.org, 127.0.0.1, etc)"`

	DBtype     string `critical:"true"`
	DBhost     string `critical:"true"`
	DBname     string `critical:"true"`
	DBusername string `critical:"true"`
	DBpassword string `critical:"true"`
	DBprefix   string `description:"Each table's name in the database will start with this, if it is set"`

	DebugMode  bool `description:"Disables several spam/browser checks that can cause problems when hosting an instance locally."`
	RandomSeed string
	Version    *GochanVersion `json:"-"`
	TimeZone   int            `json:"-"`
}

// SiteConfig contains information about the site/community, e.g. the name of the site, the slogan (if set),
// the first page to look for if a directory is requested, etc
type SiteConfig struct {
	FirstPage       []string
	Username        string
	CookieMaxAge    string `description:"The amount of time that session cookies will exist before they expire (ex: 1y2mo3d4h or 1 year 2 months 3 days 4 hours). Default is 1 year"`
	Lockdown        bool   `description:"Disables posting."`
	LockdownMessage string `description:"Message displayed when someone tries to post while the site is on lockdown."`

	SiteName   string `description:"The name of the site that appears in the header of the front page."`
	SiteSlogan string `description:"The text that appears below SiteName on the home page"`
	Modboard   string `description:"A super secret clubhouse board that only staff can view/post to."`

	MaxRecentPosts        int  `description:"The maximum number of posts to show on the Recent Posts list on the front page."`
	RecentPostsWithNoFile bool `description:"If checked, recent posts with no image/upload are shown on the front page (as well as those with images"`
	Verbosity             int
	EnableAppeals         bool
	MaxLogDays            int `description:"The maximum number of days to keep messages in the moderation/staff log file."`

	MinifyHTML      bool   `description:"If checked, gochan will minify html files when building"`
	MinifyJS        bool   `description:"If checked, gochan will minify js and json files when building"`
	GeoIPDBlocation string `description:"Specifies the location of the GeoIP database file. If you're using CloudFlare, you can set it to cf to rely on CloudFlare for GeoIP information."`
	AkismetAPIKey   string `description:"The API key to be sent to Akismet for post spam checking. If the key is invalid, Akismet won't be used."`

	Captcha CaptchaConfig
}

type CaptchaConfig struct {
	Type                 string
	OnlyNeededForThreads bool
	SiteKey              string
	AccountSecret        string
}

func (cc *CaptchaConfig) UseCaptcha() bool {
	return cc.SiteKey != "" && cc.AccountSecret != ""
}

type BoardCooldowns struct {
	NewThread  int `json:"threads"`
	Reply      int `json:"replies"`
	ImageReply int `json:"images"`
}

type PageBanner struct {
	Filename string
	Width    int
	Height   int
}

// BoardConfig contains information about a specific board to be stored in /path/to/board/board.json
// If a board doesn't have board.json, the site's default board config (with values set in gochan.json) will be used
type BoardConfig struct {
	InheritGlobalStyles bool     `description:"If checked, a board uses the global Styles array + the board config's styles (with duplicates removed)"`
	Styles              []Style  `description:"List of styles (one per line) that should be accessed online at <SiteWebFolder>/css/<Style>"`
	DefaultStyle        string   `description:"Filename of the default Style. If this unset, the first entry in the Styles array will be used."`
	Sillytags           []string `description:"List of randomly selected fake staff tags separated by line, e.g. ## Mod, to be randomly assigned to posts if UseSillytags is checked. Don't include the \"## \""`
	UseSillytags        bool     `description:"Use Sillytags"`
	Banners             []PageBanner

	PostConfig
	UploadConfig

	DateTimeFormat         string `description:"The format used for dates. See <a href=\"https://golang.org/pkg/time/#Time.Format\">here</a> for more info."`
	AkismetAPIKey          string `description:"The API key to be sent to Akismet for post spam checking. If the key is invalid, Akismet won't be used."`
	ShowPosterID           bool
	EnableSpoileredImages  bool
	EnableSpoileredThreads bool
	Worksafe               bool
	ThreadPage             int
	Cooldowns              BoardCooldowns
	ThreadsPerPage         int
	EnableGeoIP            bool
}

type BoardListConfig struct {
	CustomLinks map[string]string // <a href="value">index</a> - can be internal or external
	HideBoards  []string          // test,boardtohide,modboard,etc
}

// Style represents a theme (Pipes, Dark, etc)
type Style struct {
	Name     string
	Filename string
}

type UploadConfig struct {
	RejectDuplicateImages bool `description:"Enabling this will cause gochan to reject a post if the image has already been uploaded for another post.\nThis may end up being removed or being made board-specific in the future."`
	ThumbWidth            int  `description:"OP thumbnails use this as their max width.<br />To keep the aspect ratio, the image will be scaled down to the ThumbWidth or ThumbHeight, whichever is larger."`
	ThumbHeight           int  `description:"OP thumbnails use this as their max height.<br />To keep the aspect ratio, the image will be scaled down to the ThumbWidth or ThumbHeight, whichever is larger."`
	ThumbWidthReply       int  `description:"Same as ThumbWidth and ThumbHeight but for reply images."`
	ThumbHeightReply      int  `description:"Same as ThumbWidth and ThumbHeight but for reply images."`
	ThumbWidthCatalog     int  `description:"Same as ThumbWidth and ThumbHeight but for catalog images."`
	ThumbHeightCatalog    int  `description:"Same as ThumbWidth and ThumbHeight but for catalog images."`

	// Sets what (if any) metadata to remove from uploaded images using exiftool.
	// Valid values are "", "none" (has the same effect as ""), "exif", or "all" (for stripping all metadata)
	StripImageMetadata string
	// The path to the exiftool command. If unset or empty, the system path will be used to find it
	ExiftoolPath string
}

type PostConfig struct {
	MaxLineLength int      `description:"Any line in a post that exceeds this will be split into two (or more) lines.<br />I'm not really sure why this is here, so it may end up being removed."`
	ReservedTrips []string `description:"Secure tripcodes (!!Something) can be reserved here.<br />Each reservation should go on its own line and should look like this:<br />TripPassword1##Tripcode1<br />TripPassword2##Tripcode2"`

	ThreadsPerPage           int
	RepliesOnBoardPage       int `description:"Number of replies to a thread to show on the board page."`
	StickyRepliesOnBoardPage int `description:"Same as above for stickied threads."`
	NewThreadsRequireUpload  bool

	BanColors        []string
	BanMessage       string `description:"The default public ban message."`
	EmbedWidth       int    `description:"The width for inline/expanded videos."`
	EmbedHeight      int    `description:"The height for inline/expanded videos."`
	EnableEmbeds     bool   `description:"If checked, adds [Embed] after a Youtube, Vimeo, etc link to toggle an inline video frame."`
	ImagesOpenNewTab bool   `description:"If checked, thumbnails will open the respective image/video in a new tab instead of expanding them." `
	NewTabOnOutlinks bool   `description:"If checked, links to external sites will open in a new tab."`
	DisableBBcode    bool   `description:"If checked, gochan will not compile bbcode into HTML"`
}

func WriteConfig() error {
	return cfg.Write()
}

// GetSystemCriticalConfig returns system-critical configuration options like listening IP
// Unlike the other functions returning the sub-configs (GetSiteConfig, GetBoardConfig, etc),
// GetSystemCriticalConfig returns the value instead of a pointer to it, because it is not usually
// safe to edit while Gochan is running.
func GetSystemCriticalConfig() SystemCriticalConfig {
	return cfg.SystemCriticalConfig
}

// GetSiteConfig returns the global site configuration (site name, slogan, etc)
func GetSiteConfig() *SiteConfig {
	return &cfg.SiteConfig
}

// GetBoardConfig returns the custom configuration for the specified board (if it exists)
// or the global board configuration if board is an empty string or it doesn't exist
func GetBoardConfig(board string) *BoardConfig {
	bc, exists := boardConfigs[board]
	if board == "" || !exists {
		return &cfg.BoardConfig
	}
	return &bc
}

// UpdateBoardConfig updates or establishes the configuration for the given board
func UpdateBoardConfig(dir string) error {
	ba, err := os.ReadFile(path.Join(cfg.DocumentRoot, dir, "board.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// board doesn't have a custom config, use global config
			return nil
		}
		return err
	}
	var config BoardConfig
	if err = json.Unmarshal(ba, &config); err != nil {
		return err
	}
	boardConfigs[dir] = config
	return nil
}

// DeleteBoardConfig removes the custom board configuration data, normally should be used
// when a board is deleted
func DeleteBoardConfig(dir string) {
	delete(boardConfigs, dir)
}

func GetDebugMode() bool {
	return cfg.testing || cfg.SystemCriticalConfig.DebugMode
}

func GetVersion() *GochanVersion {
	return cfg.Version
}

// SetVersion should (in most cases) only be used for tests, where a config file wouldn't be loaded
func SetVersion(version string) {
	if cfg == nil {
		cfg = &GochanConfig{}
		cfg.Version = ParseVersion(version)
	}
}
