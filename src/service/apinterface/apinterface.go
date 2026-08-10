package apinterface

import (
	"apigw/src/models"
	"apigw/src/pkg/answer"
	"apigw/src/pkg/common"
	"apigw/src/pkg/utils"
	"apigw/src/slog"
	"context"
	"fmt"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

type ApiInterface struct {
	ID         string `json:"id"`          // API ID
	GroupID    string `json:"group_id"`    // 所属分组ID
	Name       string `json:"name"`        // 接口名称
	Protocol   string `json:"protocol"`    // R 后端协议
	Method     string `json:"method"`      // R HTTP方法
	ReqUri     string `json:"req_uri"`     // R API路径
	BackendUri string `json:"backend_uri"` // R 后端api
	Auth       string `json:"auth"`        // R 认证类型
	Mode       string `json:"mode"`        // R API的匹配方式prefix：前缀匹配,exact：精确匹配
	LcID       string `json:"lc_id"`       // R 关联负载通道ID
	RateLimit  int    `json:"rate_limit"`  // R 接口限流（QPS，0-不限流）
	Status     int8   `json:"status"`      // 状态（0-禁用，1-启用）
	Publish    int8   `json:"publish"`     // 发布状态（0-未发布，1-测试中，2-已发布，3-已下线）
	CreateTime int64  `json:"create_time"` // 创建时间戳
	UpdateTime int64  `json:"update_time"` // 更新时间戳
}

type ApiLifeSrt struct {
	Publish int8 `json:"publish"`
}

type ApiStatusSut struct {
	Status *int8 `json:"status"`
}

const (
	StatusDisable     = models.StatusDisable     // 禁用
	StatusEnable      = models.StatusEnable      // 启用
	PublishUnReleased = models.PublishUnReleased // 未发布
	PublishTesting    = models.PublishTesting    // 测试中
	PublishReleased   = models.PublishReleased   // 已发布
	PublishOffline    = models.PublishOffline    // 已下线
)

// ListApi GET /group1/api?scope=all 逻辑：scope=all 忽略 groupId，查全量；否则按 groupId 查询。
func ListApi() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		limit, offset, err := common.GetPagingQuery(c)
		if err != nil {
			klog.Errorf("GetPagingQuery err: %s", err)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, err.Error(), nil))
			return
		}

		dao, err := models.NewDBModel("SYSTEM")
		if err != nil {
			klog.Errorf("NewDBModel err: %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}
		list, count, err := dao.ListApiInterface(limit, offset)
		if err != nil {
			klog.Error("ListApiInterface err:: ", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}

		result := make([]*ApiInterface, len(list))
		for k, v := range list {
			item := &ApiInterface{
				ID:         v.ID,
				GroupID:    v.GroupID,
				Name:       v.Name,
				Protocol:   v.Protocol,
				Method:     v.Method,
				ReqUri:     v.ReqUri,
				BackendUri: v.BackendUri,
				Auth:       v.Auth,
				Mode:       v.Mode,
				LcID:       v.LcID,
				RateLimit:  v.RateLimit,
				Status:     v.Status,
				Publish:    v.Publish,
				CreateTime: v.CreateTime,
				UpdateTime: v.UpdateTime,
			}
			result[k] = item
		}

		pageInfo := answer.SetPageInfo(limit, offset, count)
		payload := map[string]any{
			"items":     result,
			"page_info": pageInfo,
		}

		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, payload))
	}
}

func CreateApi() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		var req ApiInterface
		if err := c.BindJSON(&req); err != nil {
			klog.Warnf("bind json failed, err: %v", err)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "Failed to parse request data. Check JSON format.", nil))
			return
		}

		groupId := c.Param("groupId")
		if groupId == "" {
			klog.Warnf("create api groupId empty")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "groupId cannot be empty", nil))
			return
		}
		if !utils.CheckUuid(groupId) {
			klog.Warnf("create api groupId is invalid. [%s]", groupId)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "groupId is invalid", nil))
			return
		}

		// Validation of all mandatory fields
		var missingFields []string
		if req.Name == "" {
			missingFields = append(missingFields, "name")
		}
		if req.Protocol == "" {
			missingFields = append(missingFields, "protocol")
		}
		if req.Method == "" {
			missingFields = append(missingFields, "method")
		}
		if req.ReqUri == "" {
			missingFields = append(missingFields, "req_uri")
		}
		if req.BackendUri == "" {
			missingFields = append(missingFields, "backend_uri")
		}
		if req.Auth == "" {
			missingFields = append(missingFields, "auth")
		}
		if req.Mode == "" {
			missingFields = append(missingFields, "mode")
		}
		if req.LcID == "" {
			missingFields = append(missingFields, "lc_id")
		}

		if len(missingFields) > 0 {
			msg := fmt.Sprintf("required fields cannot be empty: %v", missingFields)
			klog.Warnf("create api missing required fields: %v", missingFields)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, msg, nil))
			return
		}

		if !utils.CheckUuid(req.LcID) {
			klog.Warnf("create api lcID invalid. [%s]", req.LcID)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "lcID is invalid uuid", nil))
			return
		}

		dao, err := models.NewDBModel("SYSTEM")
		if err != nil {
			klog.Errorf("NewDBModel err: %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}

		t := common.CreateTimestamp()
		apiID := utils.CreateUuid()

		data := &models.OrmApiInterface{
			ID:         apiID,
			GroupID:    groupId,
			Name:       req.Name,
			Protocol:   req.Protocol,
			Method:     req.Method,
			ReqUri:     req.ReqUri,
			BackendUri: req.BackendUri,
			Auth:       req.Auth,
			Mode:       req.Mode,
			LcID:       req.LcID,
			RateLimit:  req.RateLimit,
			Status:     StatusEnable,
			Publish:    PublishUnReleased,
			CreateTime: t,
			UpdateTime: t,
		}
		err = dao.CreateApiInterface(data)
		if err != nil {
			klog.Errorf("create api to db error: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "create api to db error.", nil))
			return
		}

		req.ID = apiID
		req.GroupID = groupId
		req.Name = data.Name
		req.Status = data.Status
		req.Publish = data.Publish
		req.CreateTime = data.CreateTime
		req.UpdateTime = data.UpdateTime
		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, req))
	}
}

func DeleteApi() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		id := c.Param("id")
		if id == "" {
			klog.Warnf("delete api id empty")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "id cannot be empty", nil))
			return
		}
		if !utils.CheckUuid(id) {
			klog.Warnf("api id is invalid. [%s]", id)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "api id is invalid", nil))
			return
		}

		dao, err := models.NewDBModel("SYSTEM")
		if err != nil {
			klog.Errorf("NewDBModel error %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}

		api, err := dao.GetApiById(id)
		if err != nil {
			klog.Errorf("UpdateApiInterface database query failed, err: %v", err)
			c.JSON(http.StatusInternalServerError, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}
		if api == nil {
			klog.Errorf("UpdateApiInterface target api data not found")
			c.JSON(http.StatusNotFound, answer.ResBody(answer.EcodeStatusNotFound, "target api does not exist.", nil))
			return
		}

		// The release interface doesn't allow to delete
		if api.Publish != PublishOffline && api.Publish != PublishUnReleased {
			klog.Errorf("UpdateApiInterface published api cannot be edited")
			c.JSON(http.StatusBadRequest, answer.ResBody(answer.EcodeInvalidRequestErr, "published api cannot be edited.", nil))
			return

		}

		if err := dao.DeleteApiInterface(id); err != nil {
			klog.Errorf("DeleteApiInterface database delete failed, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "error deleting api record from the database.", nil))
			return
		}

		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, nil))

	}
}

func UpdateApi() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		id := c.Param("id")
		if !utils.CheckUuid(id) {
			errMsg := "api id is invalid"
			klog.Errorf("%s. [%s]", errMsg, id)
			c.JSON(http.StatusBadRequest, answer.ResBody(answer.EcodeInvalidApiId, errMsg, nil))
			return
		}
		var req ApiInterface
		if err := c.BindJSON(&req); err != nil {
			klog.Warnf("bind json failed, err: %v", err)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "Failed to parse request data. Check JSON format.", nil))
			return
		}
		updateData := make(map[string]any)
		if req.Name != "" {
			updateData["name"] = req.Name
		}
		if req.Protocol != "" {
			updateData["protocol"] = req.Protocol
		}
		if req.Method != "" {
			updateData["method"] = req.Method
		}
		if req.ReqUri != "" {
			updateData["req_uri"] = req.ReqUri
		}
		if req.BackendUri != "" {
			updateData["backend_uri"] = req.BackendUri
		}
		if req.Auth != "" {
			updateData["auth"] = req.Auth
		}
		if req.Mode != "" {
			updateData["mode"] = req.Mode
		}
		if req.LcID != "" {
			updateData["lc_id"] = req.LcID
		}
		if req.RateLimit != 0 {
			updateData["rate_limit"] = req.RateLimit
		}

		dao, err := models.NewDBModel("SYSTEM")
		if err != nil {
			klog.Errorf("NewDBModel err: %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}

		api, err := dao.GetApiById(id)
		if err != nil {
			klog.Errorf("UpdateApiInterface database query failed, err: %v", err)
			c.JSON(http.StatusInternalServerError, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}
		if api == nil {
			klog.Errorf("UpdateApiInterface target api data not found")
			c.JSON(http.StatusNotFound, answer.ResBody(answer.EcodeStatusNotFound, "target api does not exist.", nil))
			return
		}

		// The release interface doesn't allow editing
		if api.Publish != PublishOffline && api.Publish != PublishUnReleased {
			klog.Errorf("UpdateApiInterface published api cannot be edited")
			c.JSON(http.StatusBadRequest, answer.ResBody(answer.EcodeInvalidRequestErr, "published api cannot be edited.", nil))
			return

		}

		if err := dao.UpdateApiInterface(id, updateData); err != nil {
			klog.Errorf("UpdateApiInterface database update failed, err: %v", err)
			c.JSON(http.StatusInternalServerError, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}

		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, nil))
	}
}

func GetApiDetail() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		id := c.Param("id")
		if id == "" {
			klog.Warnf("get api detail id empty")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "id cannot be empty", nil))
			return
		}
		if !utils.CheckUuid(id) {
			klog.Warnf("api id is invalid. [%s]", id)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "api id is invalid", nil))
			return
		}

		dao, err := models.NewDBModel("SYSTEM")
		if err != nil {
			klog.Errorf("NewDBModel err: %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}
		api, err := dao.GetApiById(id)
		if err != nil {
			klog.Error("GetApiById err: ", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}

		result := &ApiInterface{
			ID:         api.ID,
			GroupID:    api.GroupID,
			Name:       api.Name,
			Protocol:   api.Protocol,
			Method:     api.Method,
			ReqUri:     api.ReqUri,
			BackendUri: api.BackendUri,
			Auth:       api.Auth,
			Mode:       api.Mode,
			LcID:       api.LcID,
			RateLimit:  api.RateLimit,
			Status:     api.Status,
			Publish:    api.Publish,
			CreateTime: api.CreateTime,
			UpdateTime: api.UpdateTime,
		}

		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, result))
	}
}

func ApiLifecycle() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		var req ApiLifeSrt
		if err := c.BindJSON(&req); err != nil {
			klog.Warnf("bind json failed, err: %v", err)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "Failed to parse request data. Check JSON format.", nil))
			return
		}

		ApiID := c.Param("id")
		if ApiID == "" {
			klog.Warnf("api lifecycle operate id empty")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "id cannot be empty", nil))
			return
		}
		if !utils.CheckUuid(ApiID) {
			klog.Warnf("api id is invalid. [%s]", ApiID)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "api id is invalid", nil))
			return
		}

		publish := req.Publish
		if publish != PublishTesting && publish != PublishReleased && publish != PublishOffline {
			klog.Warn("publish status error.")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "invalid publish status", nil))
			return
		}

		dao, err := models.NewSystemDBModel()
		if err != nil {
			klog.Errorf("NewDBModel err: %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}
		// offline operation
		if publish == PublishOffline {
			if err := dao.ApiInterface.OfflineApiById(ApiID); err != nil {
				klog.Errorf("OfflineApiById err: %s", err)
				c.JSON(500, answer.ResBody(answer.EcodeError, "offline operation failed", nil))
				return
			}
			if err := OfflineApi(ApiID); err != nil {
				klog.Errorf("offline api, refresh cache failed, err: %v", err)
			}
		}
		// publish online operation
		if publish == PublishReleased {
			if err := dao.ApiInterface.OnlineApiById(ApiID); err != nil {
				klog.Errorf("OnlineApiById err: %s", err)
				c.JSON(500, answer.ResBody(answer.EcodeError, "publish operation failed", nil))
				return
			}
			if err := OnlineApi(ApiID); err != nil {
				klog.Errorf("online api, refresh cache failed, err: %v", err)
			}
		}

		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, nil))
	}
}

func ApiStatus() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		var req ApiStatusSut
		if err := c.BindJSON(&req); err != nil {
			klog.Warnf("bind json failed, err: %v", err)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "Failed to parse request data. Check JSON format.", nil))
			return
		}

		if req.Status == nil {
			klog.Warn("api status field is required")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "status cannot be empty", nil))
			return
		}
		status := *req.Status
		apiID := c.Param("id")
		if apiID == "" {
			klog.Warnf("api status operate id empty")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "id cannot be empty", nil))
			return
		}
		if !utils.CheckUuid(apiID) {
			klog.Warnf("api id is invalid. [%s]", apiID)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "api id is invalid", nil))
			return
		}

		if status != StatusEnable && status != StatusDisable {
			klog.Warn("api status value error.")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "invalid api status", nil))
			return
		}

		dao, err := models.NewSystemDBModel()
		if err != nil {
			klog.Errorf("NewDBModel err: %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}

		err = dao.ApiInterface.UpdateApiStatusById(apiID, status)
		if err != nil {
			klog.Errorf("UpdateApiStatusById err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "update api status failed", nil))
			return
		}

		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, nil))
	}
}
