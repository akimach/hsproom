package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"../config"
	"../models"
	"../templates"
	"../utils/log"

	"github.com/microcosm-cc/bluemonday"
)

type userViewMember struct {
	*templates.DefaultMember
	UserInfo     *models.User
	UserPrograms *[]models.Program
}

func userViewHandler(document http.ResponseWriter, request *http.Request) (err error) {

	var tmpl templates.Template
	tmpl.Layout = "default.tmpl"
	tmpl.Template = "userView.tmpl"

	rawUid := request.URL.Query().Get("u")
	uid, err := strconv.Atoi(rawUid)

	if err != nil {
		log.Debug(err)

		showError(document, request, "ユーザが見つかりませんでした。")
		return nil
	}

	var user models.User
	err = user.Load(uid)

	if err != nil {
		log.Debug(err)

		showError(document, request, "ユーザが見つかりませんでした。")
		return nil
	}

	var programs []models.Program

	_, err = models.GetProgramListByUser(models.ProgramColCreatedAt, &programs, user.ID, true, 0, 4)

	if err != nil {
		return errors.New("ユーザプログラム一覧の取得に失敗: \r\n" + err.Error())
	}

	return tmpl.Render(document, userViewMember{
		DefaultMember: &templates.DefaultMember{
			Title:  user.Name + " のプロフィール - " + config.SiteTitle,
			UserID: getSessionUser(request),
		},
		UserInfo:     &user,
		UserPrograms: &programs,
	})
}

func userListHandler(document http.ResponseWriter, request *http.Request) (err error) {
	return nil
}

func userLoginHandler(document http.ResponseWriter, request *http.Request) (err error) {

	var tmpl templates.Template
	tmpl.Layout = "default.tmpl"
	tmpl.Template = "userLogin.tmpl"

	return tmpl.Render(document, templates.DefaultMember{
		Title:  "ログイン",
		UserID: getSessionUser(request),
	})
}

func userLogoutHandler(document http.ResponseWriter, request *http.Request) (err error) {

	var tmpl templates.Template
	tmpl.Layout = "default.tmpl"
	tmpl.Template = "userLogout.tmpl"

	removeSession(document, request)

	return tmpl.Render(document, templates.DefaultMember{
		Title:  "ログアウト中です...",
		UserID: 0,
	})
}

func userEditHandler(document http.ResponseWriter, request *http.Request) (err error) {
	return nil
}

type userProgramsMember struct {
	*templates.DefaultMember
	Programs     []models.Program
	ProgramCount int
	CurPage      int
	MaxPage      int
	Sort         string
	UserName     string
	UserId       int
}

func userProgramsHandler(document http.ResponseWriter, request *http.Request) (err error) {

	var tmpl templates.Template
	tmpl.Layout = "default.tmpl"
	tmpl.Template = "userPrograms.tmpl"

	rawUserId := request.URL.Query().Get("u")
	userId, err := strconv.Atoi(rawUserId)

	if err != nil {

		log.DebugStr("リクエストが不正 Request:" + rawUserId)

		showError(document, request, "不正なリクエストです。")
		return nil
	}

	sort := request.URL.Query().Get("s")

	var sortKey models.ProgramColumn
	switch sort {
	case "c":
		sortKey = models.ProgramColCreatedAt
	case "g":
		sortKey = models.ProgramColGood
	case "n":
		sortKey = models.ProgramColTitle
	default:
		sortKey = models.ProgramColCreatedAt
	}

	page, err := strconv.Atoi(request.URL.Query().Get("p"))
	if err != nil {
		page = 0
	}

	if !models.ExistsUser(userId) {
		log.Debug(err)

		showError(document, request, "ユーザが存在しません。")
		return
	}

	var programs []models.Program
	i, err := models.GetProgramListByUser(sortKey, &programs, userId, true, page*10, 10)

	if err != nil {
		return errors.New("ユーザプログラム一覧の取得に失敗: \r\n" + err.Error())
	}

	maxPage := i / 10
	if i%10 == 0 {
		maxPage--
	}

	userName, err := models.GetUserName(userId)

	if err != nil {
		return errors.New("ユーザ名の取得に失敗: \r\n" + err.Error())
	}

	return tmpl.Render(document, userProgramsMember{
		DefaultMember: &templates.DefaultMember{
			Title:  userName + " - " + config.SiteTitle,
			UserID: getSessionUser(request),
		},
		Programs:     programs,
		ProgramCount: i,
		CurPage:      page,
		MaxPage:      maxPage,
		Sort:         sort,
		UserName:     userName,
		UserId:       userId,
	})
}

type userSettingsMember struct {
	*templates.DefaultMember
	UserInfo models.User
	Section string
}

func userSettingsHandler(document http.ResponseWriter, request *http.Request) (err error) {

	var tmpl templates.Template
	tmpl.Layout = "default.tmpl"
	tmpl.Template = "userSettings.tmpl"

	userId := getSessionUser(request)

	if userId == 0 {
		http.Redirect(document, request, "/user/login/", http.StatusFound)
		return nil
	}

	var user models.User
	err = user.Load(userId)

	if err != nil {
		return errors.New("ユーザの読み込みに失敗: \r\n" + err.Error())
	}

	section := bluemonday.StrictPolicy().Sanitize(request.URL.Query().Get("s"))

	return tmpl.Render(document, userSettingsMember{
		DefaultMember: &templates.DefaultMember{
			Title:  "管理画面 - " + config.SiteTitle,
			UserID: userId,
		},
		UserInfo: user,
		Section: section,
	})
}

