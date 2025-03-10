package gctemplates

import (
	"errors"
	"fmt"
	"html"
	"html/template"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gochan-org/gochan/pkg/config"
	"github.com/gochan-org/gochan/pkg/gcsql"
	"github.com/gochan-org/gochan/pkg/gcutil"
)

var (
	ErrInvalidKey = errors.New("template map expects string keys")
	ErrInvalidMap = errors.New("invalid template map call")
	maxFilename   = 10
)

var funcMap = template.FuncMap{
	// Arithmetic functions
	"add": func(a, b int) int {
		return a + b
	},
	"subtract": func(a, b int) int {
		return a - b
	},

	// Comparison functions (some copied from text/template for compatibility)
	"ge": func(a int, b int) bool {
		return a >= b
	},
	"gt": func(a int, b int) bool {
		return a > b
	},
	"le": func(a int, b int) bool {
		return a <= b
	},
	"lt": func(a int, b int) bool {
		return a < b
	},
	"intEq": func(a, b int) bool {
		return a == b
	},
	"isNil": func(i interface{}) bool {
		return i == nil
	},

	// Array functions
	"getSlice": func(arr []interface{}, start, length int) []interface{} {
		if start < 0 {
			start = 0
		}
		if length > len(arr) {
			length = len(arr)
		}
		return arr[start:length]
	},

	// String functions
	"intToString":  strconv.Itoa,
	"escapeString": html.EscapeString,
	"formatFilesize": func(sizeInt int) string {
		size := float32(sizeInt)
		if size < 1000 {
			return fmt.Sprintf("%d B", sizeInt)
		} else if size <= 100000 {
			return fmt.Sprintf("%0.1f KB", size/1024)
		} else if size <= 100000000 {
			return fmt.Sprintf("%0.2f MB", size/1024/1024)
		}
		return fmt.Sprintf("%0.2f GB", size/1024/1024/1024)
	},
	"formatTimestamp": func(t time.Time) string {
		return t.Format(config.GetBoardConfig("").DateTimeFormat)
	},
	"stringAppend": func(strings ...string) string {
		var appended string
		for _, str := range strings {
			appended += str
		}
		return appended
	},
	"truncateFilename": func(filename string) string {
		if len(filename) <= maxFilename {
			return filename
		}
		arr := strings.Split(filename, ".")
		if len(arr) == 1 {
			return arr[0][:maxFilename]
		}
		base := strings.Join(arr[:len(arr)-1], ".")
		if len(base) >= maxFilename {
			base = base[:maxFilename]
		}
		ext := arr[len(arr)-1:][0]
		return base + "." + ext
	},
	"truncateMessage": func(msg string, limit int, maxLines int) string {
		var truncated bool
		split := strings.Split(msg, "<br />")

		if len(split) > maxLines {
			split = split[:maxLines]
			msg = strings.Join(split, "<br />")
			truncated = true
		}

		if len(msg) < limit {
			if truncated {
				msg = msg + "..."
			}
			return msg
		}
		msg = msg[:limit]
		truncated = true

		if truncated {
			msg = msg + "..."
		}
		return msg
	},
	"truncateHTMLMessage": truncateHTML,
	"stripHTML": func(htmlStr template.HTML) string {
		return gcutil.StripHTML(string(htmlStr))
	},
	"truncateString": func(msg string, limit int, ellipsis bool) string {
		if len(msg) > limit {
			if ellipsis {
				return msg[:limit] + "..."
			}
			return msg[:limit]
		}
		return msg
	},
	"map": func(values ...interface{}) (map[string]interface{}, error) {
		dict := make(map[string]interface{})
		if len(values)%2 != 0 {
			return nil, ErrInvalidMap
		}
		for k := 0; k < len(values); k += 2 {
			key, ok := values[k].(string)
			if !ok {
				fmt.Printf("%q\n\n", key)
				return nil, ErrInvalidKey
			}
			dict[key] = values[k+1]
		}
		return dict, nil
	},
	"until": func(t time.Time) string {
		return t.Sub(time.Now()).String()
	},
	"dereference": func(a *int) int {
		if a == nil {
			return 0
		}
		return *a
	},
	// Imageboard functions
	"bannedForever": func(ban *gcsql.IPBan) bool {
		return ban.IsActive && ban.Permanent && !ban.CanAppeal
	},
	"isBanned": func(ban *gcsql.IPBan, board string) bool {
		return ban.IsActive && ban.BoardID != nil
	},
	"getBoardDirFromID": func(id int) string {
		dir, _ := gcsql.GetBoardDir(id)
		return dir
	},
	"intPtrToBoardDir": func(id *int, ifNil string, ifErr string) string {
		if id == nil {
			return ifNil
		}
		dir, err := gcsql.GetBoardDir(*id)
		if err != nil {
			return ifErr
		}
		return dir
	},
	"getStaffNameFromID": func(id int) string {
		username, err := gcsql.GetStaffUsernameFromID(id)
		if err != nil {
			return "?"
		}
		return username
	},
	"getAppealBanIP": func(appealID int) string {
		ban, err := gcsql.GetIPBanByID(appealID)
		if err != nil || ban == nil {
			return "?"
		}
		return ban.IP
	},
	"getCatalogThumbnail": func(img string) string {
		return gcutil.GetThumbnailPath("catalog", img)
	},
	"getTopPostID": func(post *gcsql.Post) int {
		id, _ := post.TopPostID()
		return id
	},
	"getThreadThumbnail": func(img string) string {
		return gcutil.GetThumbnailPath("thread", img)
	},
	"getUploadType": func(name string) string {
		return gcutil.GetThumbnailExt(name)
	},
	"imageToThumbnailPath": func(thumbType string, img string) string {
		filetype := strings.ToLower(img[strings.LastIndex(img, ".")+1:])
		if filetype == "gif" || filetype == "webm" || filetype == "mp4" {
			filetype = "jpg"
		}
		index := strings.LastIndex(img, ".")
		if index < 0 || index > len(img) {
			return ""
		}
		thumbSuffix := "t." + filetype
		if thumbType == "catalog" {
			thumbSuffix = "c." + filetype
		}
		return img[0:index] + thumbSuffix
	},
	"numReplies": func(boardid, opID int) int {
		num, err := gcsql.GetThreadReplyCountFromOP(opID)
		if err != nil {
			return 0
		}
		return num
	},
	"getBoardDir": func(id int) string {
		dir, err := gcsql.GetBoardDir(id)
		if err != nil {
			return ""
		}
		return dir
	},
	"boardPagePath": func(board *gcsql.Board, page int) string {
		return config.WebPath(board.Dir, strconv.Itoa(page)+".html")
	},
	"webPath": func(part ...string) string {
		return config.WebPath(part...)
	},
	// Template convenience functions
	"makeLoop": func(n int, offset int) []int {
		loopArr := make([]int, n)
		for i := range loopArr {
			loopArr[i] = i + offset
		}
		return loopArr
	},
	"generateConfigTable": func() template.HTML {
		siteCfg := config.GetSiteConfig()
		boardCfg := config.GetBoardConfig("")
		tableOut := `<table style="border-collapse: collapse;" id="config"><tr><th>Field name</th><th>Value</th><th>Type</th><th>Description</th></tr>`

		tableOut += configTable(siteCfg) +
			configTable(boardCfg) +
			"</table>"
		return template.HTML(tableOut)
	},
	"isStyleDefault": func(style string) bool {
		return style == config.GetBoardConfig("").DefaultStyle
	},
	"version": func() string {
		return config.GetVersion().String()
	},
}

func configTable(cfg interface{}) string {
	cVal := reflect.ValueOf(cfg)
	if cVal.Kind() == reflect.Ptr {
		cVal = cVal.Elem()
	}
	var tableOut string
	if cVal.Kind() != reflect.Struct {
		return ""
	}
	cType := cVal.Type()
	numFields := cVal.NumField()

	for f := 0; f < numFields; f++ {
		field := cType.Field(f)
		name := field.Name

		fVal := reflect.Indirect(cVal).FieldByName(name)
		fKind := fVal.Kind()
		// interf := cVal.Field(f).Interface()
		switch fKind {
		case reflect.Int:
			tableOut += `<input name="` + name + `" type="number" value="` + html.EscapeString(fmt.Sprintf("%v", f)) + `" class="config-text"/>`
		case reflect.String:
			tableOut += `<input name="` + name + `" type="text" value="` + html.EscapeString(fmt.Sprintf("%v", f)) + `" class="config-text"/>`
		case reflect.Bool:
			checked := ""
			if fVal.Bool() {
				checked = "checked"
			}
			tableOut += `<input name="` + name + `" type="checkbox" ` + checked + " />"

		case reflect.Slice:
			tableOut += `<textarea name="` + name + `" rows="4" cols="28">`
			arrLength := fVal.Len()
			for s := 0; s < arrLength; s++ {
				newLine := "\n"
				if s == arrLength-1 {
					newLine = ""
				}
				tableOut += html.EscapeString(fVal.Slice(s, s+1).Index(0).String()) + newLine
			}
			tableOut += "</textarea>"
		default:
			tableOut += fmt.Sprintf("%v", fKind)
		}

		tableOut += "</td><td>" + fKind.String() + "</td><td>"
		defaultTag := field.Tag.Get("default")
		var defaultTagHTML string
		if defaultTag != "" {
			defaultTagHTML = " <b>Default: " + defaultTag + "</b>"
		}
		tableOut += field.Tag.Get("description") + defaultTagHTML + "</td>"
		tableOut += "</tr>"
	}
	return tableOut
}
