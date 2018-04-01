package router

import (
	"bytes"
	"fmt"
	"github.com/traPtitech/traQ/external/storage"
	"github.com/traPtitech/traQ/utils/validator"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/srinathgs/mysqlstore"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/traQ/model"
	"github.com/traPtitech/traQ/rbac/role"
)

var (
	testUser = &model.User{
		Name:  "testUser",
		Email: "example@trap.jp",
		Icon:  "empty",
		Role:  role.User.ID(),
	}
	engine *xorm.Engine
)

func TestMain(m *testing.M) {
	time.Local = time.UTC

	user := os.Getenv("MARIADB_USERNAME")
	if user == "" {
		user = "root"
	}

	pass := os.Getenv("MARIADB_PASSWORD")
	if pass == "" {
		pass = "password"
	}

	host := os.Getenv("MARIADB_HOSTNAME")
	if host == "" {
		host = "127.0.0.1"
	}

	port := os.Getenv("MARIADB_PORT")
	if port == "" {
		port = "3306"
	}

	dbname := "traq-test-router"

	var err error
	engine, err = xorm.NewEngine("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true", user, pass, host, port, dbname))
	if err != nil {
		panic(err)
	}
	defer engine.Close()
	engine.ShowSQL(false)
	engine.SetMapper(core.GonicMapper{})
	model.SetXORMEngine(engine)

	// テストで作成されたfileは全てメモリ上に乗ります。容量注意
	model.SetFileManager("", storage.NewInMemoryFileManager())

	if err := model.SyncSchema(); err != nil {
		panic(err)
	}

	code := m.Run()

	os.Exit(code)
}

func beforeTest(t *testing.T) (*echo.Echo, *http.Cookie, echo.MiddlewareFunc, *assert.Assertions, *require.Assertions) {
	require := require.New(t)

	require.NoError(model.DropTables())
	require.NoError(model.SyncSchema())
	e := echo.New()
	e.Validator = validator.New()

	store, err := mysqlstore.NewMySQLStoreFromConnection(engine.DB().DB, "sessions", "/", 60*60*24*14, []byte("secret"))
	require.NoError(err)

	require.NoError(testUser.SetPassword("test"))
	require.NoError(testUser.Create())

	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	sess, err := store.New(req, "sessions")
	require.NoError(err)

	sess.Values["userID"] = testUser.ID
	require.NoError(sess.Save(req, rec))

	cookie := parseCookies(rec.Header().Get("Set-Cookie"))["sessions"]
	mw := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return session.Middleware(store)(UserAuthenticate(nil)(next))(c)
		}

	}
	return e, cookie, mw, assert.New(t), require
}

func beforeLoginTest(t *testing.T) (*echo.Echo, echo.MiddlewareFunc) {
	require := require.New(t)

	engine.DropTables("sessions", "messages", "users_private_channels", "channels", "users", "clips", "stars")
	require.NoError(model.SyncSchema())
	e := echo.New()

	store, err := mysqlstore.NewMySQLStoreFromConnection(engine.DB().DB, "sessions", "/", 60*60*24*14, []byte("secret"))
	require.NoError(err)

	req := httptest.NewRequest(echo.GET, "/", nil)
	_, err = store.New(req, "sessions")
	require.NoError(err)

	mw := session.Middleware(store)
	return e, mw
}

func parseCookies(value string) map[string]*http.Cookie {
	m := map[string]*http.Cookie{}
	for _, c := range (&http.Request{Header: http.Header{"Cookie": {value}}}).Cookies() {
		m[c.Name] = c
	}
	return m
}

func requestWithContext(t *testing.T, handler echo.HandlerFunc, c echo.Context) {
	err := handler(c)

	if err != nil {
		t.Fatal(err)
	}
}

func request(e *echo.Echo, t *testing.T, handler echo.HandlerFunc, cookie *http.Cookie, req *http.Request) *httptest.ResponseRecorder {
	if req == nil {
		req = httptest.NewRequest("GET", "http://test", nil)
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if cookie != nil {
		req.Header.Add("Cookie", fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handler(c)

	if err != nil {
		t.Fatal(err)
	}

	return rec
}

func getContext(e *echo.Echo, t *testing.T, cookie *http.Cookie, req *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	if req == nil {
		req = httptest.NewRequest("GET", "http://test", nil)
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if cookie != nil {
		req.Header.Add("Cookie", fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func mustMakeChannel(t *testing.T, userID, name string, isPublic bool) *model.Channel {
	channel := &model.Channel{
		CreatorID: userID,
		Name:      name,
		IsPublic:  isPublic,
	}
	require.NoError(t, channel.Create())
	return channel
}

func mustMakeInvisibleChannel(t *testing.T, userID, name string, isPublic bool) *model.UserInvisibleChannel {
	c := mustMakeChannel(t, testUser.ID, name, isPublic)
	i := &model.UserInvisibleChannel{
		UserID:    userID,
		ChannelID: c.ID,
	}
	require.NoError(t, i.Create())
	return i

}

func mustMakeMessage(t *testing.T, userID, channelID string) *model.Message {
	message := &model.Message{
		UserID:    userID,
		ChannelID: channelID,
		Text:      "popopo",
	}
	require.NoError(t, message.Create())
	return message
}

func mustMakeTag(t *testing.T, userID, tagText string) *model.UsersTag {
	tag := &model.UsersTag{
		UserID: userID,
	}
	require.NoError(t, tag.Create(tagText))
	return tag
}

func mustMakeUnread(t *testing.T, userID, messageID string) *model.Unread {
	unread := &model.Unread{
		UserID:    userID,
		MessageID: messageID,
	}
	require.NoError(t, unread.Create())
	return unread
}

func mustClipMessage(t *testing.T, userID, messageID string) *model.Clip {
	clip := &model.Clip{
		UserID:    userID,
		MessageID: messageID,
	}
	require.NoError(t, clip.Create())
	return clip
}

func mustStarChannel(t *testing.T, userID, channelID string) *model.Star {
	star := &model.Star{
		UserID:    userID,
		ChannelID: channelID,
	}
	require.NoError(t, star.Create())
	return star
}

func mustMakePin(t *testing.T, channelID, userID, messageID string) *model.Pin {
	pin := &model.Pin{
		ChannelID: channelID,
		UserID:    userID,
		MessageID: messageID,
	}

	require.NoError(t, pin.Create())
	return pin
}

func mustCreateUser(t *testing.T, name string) *model.User {
	user := &model.User{
		Name:  name,
		Email: "example@trap.jp",
		Icon:  "empty",
		Role:  role.User.ID(),
	}
	require.NoError(t, user.SetPassword("test"))
	require.NoError(t, user.Create())
	return user
}

func mustMakeFile(t *testing.T) *model.File {
	file := &model.File{
		Name:      "test.txt",
		Size:      90,
		CreatorID: testUser.ID,
	}
	require.NoError(t, file.Create(bytes.NewBufferString("test message")))
	return file
}
