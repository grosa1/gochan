package main

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/gochan-org/gochan/pkg/building"
	"github.com/gochan-org/gochan/pkg/config"
	"github.com/gochan-org/gochan/pkg/gcsql"
	"github.com/gochan-org/gochan/pkg/gctemplates"
	"github.com/gochan-org/gochan/pkg/gcutil"
	"github.com/gochan-org/gochan/pkg/manage"
	"github.com/gochan-org/gochan/pkg/posting"
	"github.com/gochan-org/gochan/pkg/serverutil"
)

var (
	server *gochanServer
)

type gochanServer struct {
	namespaces map[string]func(http.ResponseWriter, *http.Request)
}

func (s gochanServer) serveFile(writer http.ResponseWriter, request *http.Request) {
	systemCritical := config.GetSystemCriticalConfig()
	siteConfig := config.GetSiteConfig()

	filePath := path.Join(systemCritical.DocumentRoot, request.URL.Path)
	var fileBytes []byte
	results, err := os.Stat(filePath)
	if err != nil {
		// the requested path isn't a file or directory, 404
		serverutil.ServeNotFound(writer, request)
		return
	}

	//the file exists, or there is a folder here
	if results.IsDir() {
		//check to see if one of the specified index pages exists
		var found bool
		for _, value := range siteConfig.FirstPage {
			newPath := path.Join(filePath, value)
			_, err := os.Stat(newPath)
			if err == nil {
				filePath = newPath
				found = true
				break
			}
		}
		if !found {
			serverutil.ServeNotFound(writer, request)
			return
		}
	}
	s.setFileHeaders(filePath, writer)

	// serve the requested file
	fileBytes, _ = os.ReadFile(filePath)
	gcutil.LogAccess(request).Int("status", 200).Send()
	writer.Write(fileBytes)
}

// set mime type/cache headers according to the file's extension
func (*gochanServer) setFileHeaders(filename string, writer http.ResponseWriter) {
	extension := strings.ToLower(gcutil.GetFileExtension(filename))
	switch extension {
	case "png":
		writer.Header().Set("Content-Type", "image/png")
		writer.Header().Set("Cache-Control", "max-age=86400")
	case "gif":
		writer.Header().Set("Content-Type", "image/gif")
		writer.Header().Set("Cache-Control", "max-age=86400")
	case "jpg":
		fallthrough
	case "jpeg":
		writer.Header().Set("Content-Type", "image/jpeg")
		writer.Header().Set("Cache-Control", "max-age=86400")
	case "css":
		writer.Header().Set("Content-Type", "text/css")
		writer.Header().Set("Cache-Control", "max-age=43200")
	case "js":
		writer.Header().Set("Content-Type", "text/javascript")
		writer.Header().Set("Cache-Control", "max-age=43200")
	case "json":
		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("Cache-Control", "max-age=5, must-revalidate")
	case "webm":
		writer.Header().Set("Content-Type", "video/webm")
		writer.Header().Set("Cache-Control", "max-age=86400")
	case "htm":
		fallthrough
	case "html":
		writer.Header().Set("Content-Type", "text/html")
		writer.Header().Set("Cache-Control", "max-age=5, must-revalidate")
	default:
		writer.Header().Set("Content-Type", "application/octet-stream")
		writer.Header().Set("Cache-Control", "max-age=86400")
	}
}

func (s gochanServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	for name, namespaceFunction := range s.namespaces {
		if request.URL.Path == config.WebPath(name) {
			namespaceFunction(writer, request)
			return
		}
	}
	s.serveFile(writer, request)
}

func initServer() {
	systemCritical := config.GetSystemCriticalConfig()
	siteConfig := config.GetSiteConfig()

	listener, err := net.Listen("tcp", systemCritical.ListenIP+":"+strconv.Itoa(systemCritical.Port))
	if err != nil {
		gcutil.Logger().Fatal().
			Err(err).
			Str("ListenIP", systemCritical.ListenIP).
			Int("Port", systemCritical.Port).Send()
		fmt.Printf("Failed listening on %s:%d: %s", systemCritical.ListenIP, systemCritical.Port, err.Error())
	}
	server = new(gochanServer)
	server.namespaces = make(map[string]func(http.ResponseWriter, *http.Request))

	// Check if Akismet API key is usable at startup.
	err = serverutil.CheckAkismetAPIKey(siteConfig.AkismetAPIKey)
	if err != nil && err != serverutil.ErrBlankAkismetKey {
		gcutil.Logger().Err(err).
			Msg("Akismet spam protection will be disabled")
		fmt.Println("Got error when initializing Akismet spam protection, it will be disabled:", err)
	}

	server.namespaces["banned"] = posting.BanHandler
	server.namespaces["captcha"] = posting.ServeCaptcha
	server.namespaces["manage"] = manage.CallManageFunction
	server.namespaces["post"] = posting.MakePost
	server.namespaces["util"] = utilHandler
	server.namespaces["example"] = func(writer http.ResponseWriter, request *http.Request) {
		if writer != nil {
			http.Redirect(writer, request, "https://www.youtube.com/watch?v=dQw4w9WgXcQ", http.StatusFound)
		}
	}
	// Eventually plugins will be able to register new namespaces or they will be restricted to something
	// like /plugin

	if systemCritical.UseFastCGI {
		err = fcgi.Serve(listener, server)
	} else {
		err = http.Serve(listener, server)
	}

	if err != nil {
		gcutil.Logger().Fatal().
			Err(err).
			Msg("Error initializing server")
		fmt.Println("Error initializing server:", err.Error())
	}
}

// handles requests to /util
func utilHandler(writer http.ResponseWriter, request *http.Request) {
	action := request.FormValue("action")
	password := request.FormValue("password")
	board := request.FormValue("board")
	boardid := request.FormValue("boardid")
	fileOnly := request.FormValue("fileonly") == "on"
	deleteBtn := request.PostFormValue("delete_btn")
	reportBtn := request.PostFormValue("report_btn")
	editBtn := request.PostFormValue("edit_btn")
	doEdit := request.PostFormValue("doedit")
	systemCritical := config.GetSystemCriticalConfig()
	wantsJSON := request.PostFormValue("json") == "1"
	if wantsJSON {
		writer.Header().Set("Content-Type", "application/json")
	}

	if action == "" && deleteBtn != "Delete" && reportBtn != "Report" && editBtn != "Edit" && doEdit != "1" {
		gcutil.LogAccess(request).Int("status", 400).Msg("received invalid /util request")
		if wantsJSON {
			writer.WriteHeader(400)
			serverutil.ServeJSON(writer, map[string]interface{}{"error": "Invalid /util request"})
		} else {
			http.Redirect(writer, request, path.Join(systemCritical.WebRoot, "/"), http.StatusBadRequest)
		}
		return
	}

	var err error
	var id int
	var checkedPosts []int
	for key, val := range request.Form {
		// get checked posts into an array
		if _, err = fmt.Sscanf(key, "check%d", &id); err != nil || val[0] != "on" {
			err = nil
			continue
		}
		checkedPosts = append(checkedPosts, id)
	}

	if reportBtn == "Report" {
		// submitted request appears to be a report
		if err = posting.HandleReport(request); err != nil {
			gcutil.LogError(err).
				Str("IP", gcutil.GetRealIP(request)).
				Ints("posts", checkedPosts).
				Str("board", board).
				Msg("Error submitting report")
			serverutil.ServeError(writer, err.Error(), wantsJSON, map[string]interface{}{
				"posts": checkedPosts,
				"board": board,
			})
			return
		}
		gcutil.LogWarning().
			Ints("reportedPosts", checkedPosts).
			Str("board", board).
			Str("IP", gcutil.GetRealIP(request)).Send()

		redirectTo := request.Referer()
		if redirectTo == "" {
			// request doesn't have a referer for some reason, redirect to board
			redirectTo = path.Join(systemCritical.WebRoot, board)
		}
		http.Redirect(writer, request, redirectTo, http.StatusFound)
		return
	}

	if editBtn == "Edit" {
		var err error
		if len(checkedPosts) == 0 {
			serverutil.ServeErrorPage(writer, "You need to select one post to edit.")
			return
		} else if len(checkedPosts) > 1 {
			serverutil.ServeErrorPage(writer, "You can only edit one post at a time.")
			return
		} else {
			rank := manage.GetStaffRank(request)
			if password == "" && rank == 0 {
				serverutil.ServeErrorPage(writer, "Password required for post editing")
				return
			}
			passwordMD5 := gcutil.Md5Sum(password)

			var post gcsql.Post
			post, err = gcsql.GetSpecificPost(checkedPosts[0], true)
			if err != nil {
				gcutil.Logger().Error().
					Err(err).
					Msg("Error getting post information")
				return
			}

			if post.Password != passwordMD5 && rank == 0 {
				serverutil.ServeErrorPage(writer, "Wrong password")
				return
			}

			if err = gctemplates.PostEdit.Execute(writer, map[string]interface{}{
				"systemCritical": config.GetSystemCriticalConfig(),
				"siteConfig":     config.GetSiteConfig(),
				"boardConfig":    config.GetBoardConfig(""),
				"post":           post,
				"referrer":       request.Referer(),
			}); err != nil {
				gcutil.Logger().Error().
					Err(err).
					Str("IP", gcutil.GetRealIP(request)).
					Msg("Error executing edit post template")

				serverutil.ServeError(writer, "Error executing edit post template: "+err.Error(), wantsJSON, nil)
				return
			}
		}
	}
	if doEdit == "1" {
		var password string
		postid, err := strconv.Atoi(request.FormValue("postid"))
		if err != nil {
			gcutil.Logger().Error().
				Err(err).
				Str("IP", gcutil.GetRealIP(request)).
				Msg("Invalid form data")
			serverutil.ServeErrorPage(writer, "Invalid form data: "+err.Error())
			return
		}
		boardid, err := strconv.Atoi(request.FormValue("boardid"))
		if err != nil {
			gcutil.Logger().Error().
				Err(err).
				Str("IP", gcutil.GetRealIP(request)).
				Msg("Invalid form data")
			serverutil.ServeErrorPage(writer, "Invalid form data: "+err.Error())
			return
		}
		password, err = gcsql.GetPostPassword(postid)
		if err != nil {
			gcutil.Logger().Error().
				Err(err).
				Str("IP", gcutil.GetRealIP(request)).
				Msg("Invalid form data")
			return
		}

		rank := manage.GetStaffRank(request)
		if request.FormValue("password") != password && rank == 0 {
			serverutil.ServeErrorPage(writer, "Wrong password")
			return
		}

		var board gcsql.Board
		if err = board.PopulateData(boardid); err != nil {
			serverutil.ServeErrorPage(writer, "Invalid form data: "+err.Error())
			gcutil.Logger().Error().
				Err(err).
				Str("IP", gcutil.GetRealIP(request)).
				Msg("Invalid form data")
			return
		}

		if err = gcsql.UpdatePost(postid, request.FormValue("editemail"), request.FormValue("editsubject"),
			posting.FormatMessage(request.FormValue("editmsg"), board.Dir), request.FormValue("editmsg")); err != nil {
			gcutil.Logger().Error().
				Err(err).
				Str("IP", gcutil.GetRealIP(request)).
				Msg("Unable to edit post")
			serverutil.ServeErrorPage(writer, "Unable to edit post: "+err.Error())
			return
		}

		building.BuildBoards(false, boardid)
		building.BuildFrontPage()
		if request.FormValue("parentid") == "0" {
			http.Redirect(writer, request, "/"+board.Dir+"/res/"+strconv.Itoa(postid)+".html", http.StatusFound)
		} else {
			http.Redirect(writer, request, "/"+board.Dir+"/res/"+request.FormValue("parentid")+".html#"+strconv.Itoa(postid), http.StatusFound)
		}
		return
	}

	if deleteBtn == "Delete" {
		// Delete a post or thread
		writer.Header().Add("Content-Type", "text/plain")
		passwordMD5 := gcutil.Md5Sum(password)
		rank := manage.GetStaffRank(request)

		if password == "" && rank == 0 {
			serverutil.ServeErrorPage(writer, "Password required for post deletion")
			return
		}

		for _, checkedPostID := range checkedPosts {
			var post gcsql.Post
			var err error
			post.ID = checkedPostID
			post.BoardID, err = strconv.Atoi(boardid)
			if err != nil {
				gcutil.Logger().Error().
					Err(err).
					Str("requestType", "deletePost").
					Str("IP", gcutil.GetRealIP(request)).
					Str("boardid", boardid).
					Int("postid", checkedPostID).Send()

				serverutil.ServeError(writer,
					fmt.Sprintf("Invalid boardid '%s' in post deletion request (got error '%s')", boardid, err),
					wantsJSON, map[string]interface{}{
						"boardid": boardid,
						"postid":  checkedPostID,
					})
				return
			}

			post, err = gcsql.GetSpecificPost(post.ID, true)
			if err == sql.ErrNoRows {
				serverutil.ServeError(writer, "Post does not exist", wantsJSON, map[string]interface{}{
					"postid":  post.ID,
					"boardid": post.BoardID,
				})
			} else if err != nil {
				gcutil.Logger().Error().
					Str("requestType", "deletePost").
					Err(err).
					Int("postid", post.ID).
					Int("boardid", post.BoardID).
					Msg("Error deleting post")
				serverutil.ServeError(writer, "Error deleting post: "+err.Error(), wantsJSON, map[string]interface{}{
					"postid":  post.ID,
					"boardid": post.BoardID,
				})
			}

			if passwordMD5 != post.Password && rank == 0 {
				serverutil.ServeError(writer, fmt.Sprintf("Incorrect password for #%d", post.ID), wantsJSON, map[string]interface{}{
					"postid":  post.ID,
					"boardid": post.BoardID,
				})
				return
			}

			if fileOnly {
				fileName := post.Filename
				if fileName != "" && fileName != "deleted" {
					var files []string
					if files, err = post.GetFilePaths(); err != nil {
						gcutil.Logger().Error().
							Str("requestType", "deleteFile").
							Int("postid", post.ID).
							Err(err).
							Msg("Error getting file upload info")
						serverutil.ServeError(writer, "Error getting file upload info: "+err.Error(), wantsJSON, map[string]interface{}{
							"postid": post.ID,
						})
						return
					}

					if err = post.UnlinkUploads(true); err != nil {
						gcutil.Logger().Error().
							Str("requestType", "deleteFile").
							Int("postid", post.ID).
							Err(err).
							Msg("Error unlinking post uploads")
						serverutil.ServeError(writer, err.Error(), wantsJSON, map[string]interface{}{
							"postid": post.ID,
						})
						return
					}

					for _, filePath := range files {
						if err = os.Remove(filePath); err != nil {
							fileBase := path.Base(filePath)
							gcutil.Logger().Error().
								Str("requestType", "deleteFile").
								Int("postid", post.ID).
								Str("file", filePath).
								Err(err).
								Msg("Error unlinking post uploads")
							serverutil.ServeError(writer, fmt.Sprintf("Error deleting %s: %s", fileBase, err.Error()), wantsJSON, map[string]interface{}{
								"postid": post.ID,
								"file":   fileBase,
							})
							return
						}
					}
				}
				_board, _ := gcsql.GetBoardFromID(post.BoardID)
				building.BuildBoardPages(&_board)

				var opPost gcsql.Post
				if post.ParentID > 0 {
					// post is a reply, get the OP
					opPost, _ = gcsql.GetSpecificPost(post.ParentID, true)
				} else {
					opPost = post
				}
				building.BuildThreadPages(&opPost)
			} else {
				// delete the post
				if err = gcsql.DeletePost(post.ID, true); err != nil {
					gcutil.Logger().Error().
						Str("requestType", "deleteFile").
						Int("postid", post.ID).
						Err(err).
						Msg("Error deleting post")
					serverutil.ServeError(writer, "Error deleting post: "+err.Error(), wantsJSON, map[string]interface{}{
						"postid": post.ID,
					})
				}
				if post.ParentID == 0 {
					threadIndexPath := path.Join(systemCritical.DocumentRoot, board, "/res/", strconv.Itoa(post.ID))
					os.Remove(threadIndexPath + ".html")
					os.Remove(threadIndexPath + ".json")
				} else {
					_board, _ := gcsql.GetBoardFromID(post.BoardID)
					building.BuildBoardPages(&_board)
				}
				building.BuildBoards(false, post.BoardID)
			}
			gcutil.Logger().Info().
				Str("requestType", "deletePost").
				Str("IP", post.IP).
				Int("boardid", post.BoardID).
				Bool("fileOnly", fileOnly).
				Msg("Post deleted")
			if !wantsJSON {
				http.Redirect(writer, request, request.Referer(), http.StatusFound)
			}
		}
	}
}
