package v3

import (
	"net/http"
	"time"

	vd "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"

	"github.com/livekit/protocol/auth"
	"github.com/traPtitech/traQ/service/webrtcv3"
)

// PostWebRTCAuthenticateRequest POST /webrtc/authenticate リクエストボディ
type PostWebRTCAuthenticateRequest struct {
	PeerID string `json:"peerId"`
	RoomID string `json:"roomId"`
}

func (r PostWebRTCAuthenticateRequest) Validate() error {
	return vd.ValidateStruct(&r,
		vd.Field(&r.PeerID, vd.Required),
	)
}

// PostWebRTCAuthenticate POST /webrtc/authenticate
func (h *Handlers) PostWebRTCAuthenticate(c echo.Context) error {
	if len(h.WebRTCSecretKey) == 0 {
		return echo.NewHTTPError(http.StatusServiceUnavailable)
	}

	var req PostWebRTCAuthenticateRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	at := auth.NewAccessToken(h.WebRTCAPIKey, h.WebRTCSecretKey) // accesstokenを生成

	canpublish, cansubscribe := true, true

	// 権限を付与
	grant := auth.VideoGrant{
		RoomCreate: true,
		RoomList:   true,
		RoomRecord: true,

		RoomJoin: true,
		Room:     req.RoomID,

		CanPublish:   &canpublish,
		CanSubscribe: &cansubscribe,
	}

	// 有効期間を設定
	vf := time.Duration(10) * time.Hour

	at.AddGrant(&grant).
		SetIdentity(req.PeerID).
		SetValidFor(vf)

	ts := time.Now().Unix()
	ttl := vf.Seconds()

	token, err := at.ToJWT() // jwtに変換
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"peerId":    req.PeerID,
		"timestamp": ts,
		"ttl":       ttl,
		"authToken": token,
	})
}

// GetWebRTCState GET /webrtc/state
func (h *Handlers) GetWebRTCState(c echo.Context) error {
	type StateSession struct {
		State     string `json:"state"`
		SessionID string `json:"sessionId"`
	}
	type WebRTCUserState struct {
		UserID    uuid.UUID      `json:"userId"`
		ChannelID uuid.UUID      `json:"channelId"`
		Sessions  []StateSession `json:"sessions"`
	}

	res := make([]WebRTCUserState, 0)
	h.WebRTC.IterateStates(func(state webrtcv3.ChannelState) {
		for _, userState := range state.Users() {
			var sessions []StateSession
			for sessionID, state := range userState.Sessions() {
				sessions = append(sessions, StateSession{State: state, SessionID: sessionID})
			}
			res = append(res, WebRTCUserState{
				UserID:    userState.UserID(),
				ChannelID: userState.ChannelID(),
				Sessions:  sessions,
			})
		}
	})

	return c.JSON(http.StatusOK, res)
}
