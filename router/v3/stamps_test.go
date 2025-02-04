package v3

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/traPtitech/traQ/model"
	"github.com/traPtitech/traQ/repository"
	"github.com/traPtitech/traQ/router/session"
	"github.com/traPtitech/traQ/utils/optional"
)

func stampEquals(t *testing.T, expect *model.Stamp, actual *httpexpect.Object) {
	t.Helper()
	actual.Value("id").String().Equal(expect.ID.String())
	actual.Value("name").String().Equal(expect.Name)
	actual.Value("creatorId").String().Equal(expect.CreatorID.String())
	actual.Value("createdAt").String().NotEmpty()
	actual.Value("updatedAt").String().NotEmpty()
	actual.Value("fileId").String().Equal(expect.FileID.String())
	actual.Value("isUnicode").Boolean().Equal(expect.IsUnicode)
}

func TestHandlers_GetStamps(t *testing.T) {
	t.Parallel()

	path := "/api/v3/stamps"
	env := Setup(t, s1)
	user := env.CreateUser(t, rand)
	stamp := env.CreateStamp(t, user.GetID(), rand)
	s := env.S(t, user.GetID())

	t.Run("not logged in", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		e.GET(path).
			Expect().
			Status(http.StatusUnauthorized)
	})

	t.Run("bad request (invalid include-unicode query)", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		e.GET(path).
			WithQuery("include-unicode", "invalid").
			WithCookie(session.CookieName, s).
			Expect().
			Status(http.StatusBadRequest)
	})

	t.Run("bad request (invalid type query)", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		e.GET(path).
			WithQuery("type", "invalid").
			WithCookie(session.CookieName, s).
			Expect().
			Status(http.StatusBadRequest)
	})

	t.Run("bad request (both query)", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		e.GET(path).
			WithQuery("include-unicode", "true").
			WithQuery("type", "original").
			WithCookie(session.CookieName, s).
			Expect().
			Status(http.StatusBadRequest)
	})

	t.Run("success (no query)", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		obj := e.GET(path).
			WithCookie(session.CookieName, s).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		obj.Length().Equal(1)
		stampEquals(t, stamp, obj.First().Object())
	})

	t.Run("success (type=original)", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		obj := e.GET(path).
			WithQuery("type", "original").
			WithCookie(session.CookieName, s).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		obj.Length().Equal(1)
		stampEquals(t, stamp, obj.First().Object())
	})

	t.Run("success (type=unicode)", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		obj := e.GET(path).
			WithQuery("type", "unicode").
			WithCookie(session.CookieName, s).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		obj.Length().Equal(0)
	})

	t.Run("success (include-unicode=false)", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		obj := e.GET(path).
			WithQuery("include-unicode", "false").
			WithCookie(session.CookieName, s).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		obj.Length().Equal(1)
		stampEquals(t, stamp, obj.First().Object())
	})
}

func TestHandlers_GetStamp(t *testing.T) {
	t.Parallel()

	path := "/api/v3/stamps/{stampId}"
	env := Setup(t, common1)
	user := env.CreateUser(t, rand)
	stamp := env.CreateStamp(t, user.GetID(), rand)
	s := env.S(t, user.GetID())

	t.Run("not logged in", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		e.GET(path, stamp.ID).
			Expect().
			Status(http.StatusUnauthorized)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		e.GET(path, uuid.Must(uuid.NewV4())).
			WithCookie(session.CookieName, s).
			Expect().
			Status(http.StatusNotFound)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		obj := e.GET(path, stamp.ID).
			WithCookie(session.CookieName, s).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		stampEquals(t, stamp, obj)
	})
}

func TestHandlers_EditStamp(t *testing.T) {
	t.Parallel()

	path := "/api/v3/stamps/{stampId}"
	env := Setup(t, common1)
	user := env.CreateUser(t, rand)
	user2 := env.CreateUser(t, rand)
	stamp := env.CreateStamp(t, user.GetID(), rand)
	stamp2 := env.CreateStamp(t, user2.GetID(), rand)
	stamp3 := env.CreateStamp(t, user.GetID(), rand)
	env.CreateStamp(t, user2.GetID(), "409_conflict")
	s := env.S(t, user.GetID())

	t.Run("not logged in", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		e.PATCH(path, stamp3.ID).
			WithJSON(&PatchStampRequest{Name: optional.StringFrom("test123")}).
			Expect().
			Status(http.StatusUnauthorized)
	})

	t.Run("bad request (empty name)", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		e.PATCH(path, stamp3.ID).
			WithCookie(session.CookieName, s).
			WithJSON(&PatchStampRequest{Name: optional.StringFrom("")}).
			Expect().
			Status(http.StatusBadRequest)
	})

	t.Run("bad request (nil creator id)", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		e.PATCH(path, stamp3.ID).
			WithCookie(session.CookieName, s).
			WithJSON(&PatchStampRequest{CreatorID: optional.UUIDFrom(uuid.Nil)}).
			Expect().
			Status(http.StatusBadRequest)
	})

	t.Run("bad request (invalid creator id)", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		e.PATCH(path, stamp3.ID).
			WithCookie(session.CookieName, s).
			WithJSON(&PatchStampRequest{CreatorID: optional.UUIDFrom(uuid.Must(uuid.NewV4()))}).
			Expect().
			Status(http.StatusBadRequest)
	})

	t.Run("forbidden", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		e.PATCH(path, stamp2.ID).
			WithCookie(session.CookieName, s).
			WithJSON(&PatchStampRequest{Name: optional.StringFrom("test123")}).
			Expect().
			Status(http.StatusForbidden)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		e.PATCH(path, uuid.Must(uuid.NewV4())).
			WithCookie(session.CookieName, s).
			WithJSON(&PatchStampRequest{Name: optional.StringFrom("test123")}).
			Expect().
			Status(http.StatusNotFound)
	})

	t.Run("conflict", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		e.PATCH(path, stamp3.ID).
			WithCookie(session.CookieName, s).
			WithJSON(&PatchStampRequest{Name: optional.StringFrom("409_conflict")}).
			Expect().
			Status(http.StatusConflict)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		e.PATCH(path, stamp.ID).
			WithCookie(session.CookieName, s).
			WithJSON(&PatchStampRequest{Name: optional.StringFrom("test123123"), CreatorID: optional.UUIDFrom(user2.GetID())}).
			Expect().
			Status(http.StatusNoContent)

		stamp, err := env.Repository.GetStamp(stamp.ID)
		require.NoError(t, err)
		assert.EqualValues(t, user2.GetID().String(), stamp.CreatorID.String())
		assert.EqualValues(t, "test123123", stamp.Name)
	})
}

func TestHandlers_DeleteStamp(t *testing.T) {
	t.Parallel()

	path := "/api/v3/stamps/{stampId}"
	env := Setup(t, common1)
	user := env.CreateUser(t, rand)
	admin := env.CreateAdmin(t, rand)
	stamp := env.CreateStamp(t, user.GetID(), rand)
	stamp2 := env.CreateStamp(t, user.GetID(), rand)
	userSession := env.S(t, user.GetID())
	adminSession := env.S(t, admin.GetID())

	t.Run("not logged in", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		e.DELETE(path, stamp2.ID).
			Expect().
			Status(http.StatusUnauthorized)
	})

	t.Run("forbidden", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		e.DELETE(path, stamp2.ID).
			WithCookie(session.CookieName, userSession).
			Expect().
			Status(http.StatusForbidden)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		e.DELETE(path, uuid.Must(uuid.NewV4())).
			WithCookie(session.CookieName, adminSession).
			Expect().
			Status(http.StatusNotFound)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		e.DELETE(path, stamp.ID).
			WithCookie(session.CookieName, adminSession).
			Expect().
			Status(http.StatusNoContent)

		_, err := env.Repository.GetStamp(stamp.ID)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})
}

func TestHandlers_GetStampStats(t *testing.T) {
	t.Parallel()

	path := "/api/v3/stamps/{stampId}/stats"
	env := Setup(t, common1)
	user := env.CreateUser(t, rand)
	stamp := env.CreateStamp(t, user.GetID(), rand)
	ch := env.CreateChannel(t, rand)
	m := env.CreateMessage(t, user.GetID(), ch.ID, rand)
	env.AddStampToMessage(t, m.GetID(), stamp.ID, user.GetID())
	env.AddStampToMessage(t, m.GetID(), stamp.ID, user.GetID())
	s := env.S(t, user.GetID())

	t.Run("not logged in", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		e.GET(path, stamp.ID).
			Expect().
			Status(http.StatusUnauthorized)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		e := env.R(t)
		obj := e.GET(path, stamp.ID).
			WithCookie(session.CookieName, s).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()

		obj.Value("count").Number().Equal(1)
		obj.Value("totalCount").Number().Equal(2)
	})
}
