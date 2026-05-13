package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"stream_hub/pkg/model/api"
	"stream_hub/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type UserApi struct {
	userServiceAddr string
	httpClient      *http.Client
}

func NewUserApi(userServiceAddr string) *UserApi {
	return &UserApi{
		userServiceAddr: userServiceAddr,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (u *UserApi) forwardGet(ctx *gin.Context, path string) (json.RawMessage, error) {
	url := fmt.Sprintf("http://%s%s", u.userServiceAddr, path)

	httpReq, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, err
	}

	token := ctx.GetHeader("Authorization")
	if token != "" {
		httpReq.Header.Set("Authorization", token)
	}

	resp, err := u.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user service returned status %d: %s", resp.StatusCode, string(body))
	}

	var upstream struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &upstream); err != nil {
		return nil, err
	}

	return upstream.Data, nil
}

func (u *UserApi) GetUserInfo(ctx *gin.Context) {
	var req api.GetUserInfoReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	data, err := u.forwardGet(ctx, fmt.Sprintf("/user/get_user_profile/%s", req.UserID))
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	utils.StatusOK(ctx, data, "User info retrieved successfully")
}

func (u *UserApi) GetUserProfile(ctx *gin.Context) {
	uid := ctx.GetString("user_id")

	data, err := u.forwardGet(ctx, fmt.Sprintf("/user/get_user_profile/%s", uid))
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	utils.StatusOK(ctx, data, "User profile retrieved successfully")
}
