package channel

import (
	"apigw/src/models"
	"apigw/src/pkg/answer"
	"apigw/src/pkg/common"
	"apigw/src/pkg/utils"
	"apigw/src/slog"
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

type CaCert struct {
	ID string `json:"id,omitempty"`
}

// R: The field can be used for updates
type LoadChannel struct {
	ID         string `json:"id"`          // 负载通道id
	Name       string `json:"name"`        // R 通道名称
	Backend    string `json:"backend"`     // R 后端地址列表（逗号分隔）
	Timeout    int    `json:"timeout"`     // R 后端超时时间（毫秒）
	HcInterval int    `json:"hcinterval"`  // R 健康检查间隔（毫秒）
	Status     int8   `json:"status"`      // 状态
	CaCert     CaCert `json:"ca_cert"`     // R 后端CA证书
	CreateTime int64  `json:"create_time"` // 创建时间戳
	UpdateTime int64  `json:"update_time"` // 更新时间戳
}

func ListChannel() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		limit, offset, err := common.GetPagingQuery(c)
		if err != nil {
			klog.Errorf("parse pagination parameters failed, err: %s", err)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, err.Error(), nil))
			return
		}

		dao, err := models.NewDBModel("SYSTEM")
		if err != nil {
			klog.Errorf("create db model failed, err: %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}
		list, count, err := dao.ListLoadChannel(limit, offset)
		if err != nil {
			klog.Errorf("query load channel list failed, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}
		result := make([]*LoadChannel, len(list))
		for k, v := range list {
			item := &LoadChannel{
				ID:         v.ID,
				Name:       v.Name,
				Backend:    v.Backend,
				Timeout:    v.Timeout,
				HcInterval: v.HcInterval,
				Status:     v.Status,
				CaCert:     CaCert{ID: v.CaCert},
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

func CreateChannel() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		var req LoadChannel
		if err := c.BindJSON(&req); err != nil {
			klog.Warnf("bind json failed, err: %v", err)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "Failed to parse request data. Check JSON format.", nil))
			return
		}

		if req.CaCert.ID != "" && !utils.CheckUuid(req.CaCert.ID) {
			klog.Warnf("ca cert id format is invalid. [%s]", req.CaCert.ID)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "ca cert id is invalid", nil))
			return
		}

		dao, err := models.NewSystemDBModel()
		if err != nil {
			klog.Errorf("create db model failed, err: %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}

		t := common.CreateTimestamp()
		id := utils.CreateUuid()
		cate := &models.OrmLoadChannel{
			ID:         id,
			Name:       req.Name,
			Backend:    req.Backend,
			Timeout:    req.Timeout,
			HcInterval: req.HcInterval,
			CaCert:     req.CaCert.ID,
			Status:     1,
			CreateTime: t,
			UpdateTime: t,
		}
		err = dao.LoadChannel.CreateLoadChannel(cate)
		if err != nil {
			klog.Errorf("create load channel record to db failed, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "failed to create load channel record in db.", nil))
			return
		}

		go func() {
			err := refreshLcCache(id)
			if err != nil {
				klog.Error("Cache refresh failed")
			}
		}()

		req.ID = id
		req.CreateTime = t
		req.UpdateTime = t
		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, req))
	}
}

func DeleteChannel() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		id := c.Param("id")
		if id == "" {
			klog.Warnf("delete load channel request, id is empty")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "id cannot be empty", nil))
			return
		}
		if !utils.CheckUuid(id) {
			klog.Warnf("load channel id format is invalid. [%s]", id)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "load channel id is invalid", nil))
			return
		}

		dao, err := models.NewDBModel("SYSTEM")
		if err != nil {
			klog.Errorf("create db model failed, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}
		channel, err := dao.GetLoadChannelByID(id)
		if err != nil {
			klog.Errorf("query load channel by uuid failed, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}
		if channel == nil {
			klog.Warnf("target load channel record does not exist, id:[%s]", id)
			c.JSON(404, answer.ResBody(answer.EcodeNotFound, "load channel record does not exist.", nil))
			return
		}

		apis, err := dao.ListApiInterfaceByLcID(id)
		if err != nil {
			klog.Errorf("query associated api interfaces by channel id failed, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}
		if len(apis) > 0 {
			klog.Warn("The load channel is bound to api interfaces and cannot be deleted.")
			c.JSON(409, answer.ResBody(answer.EcodeGroupInUse, "The load channel is in use and can't be deleted.", nil))
			return
		}

		if err := dao.DeleteLoadChannel(id); err != nil {
			klog.Errorf("execute load channel delete statement failed, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "failed to delete load channel record from the database.", nil))
			return
		}

		go func() {
			deleteLcCache(id)
		}()
		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, nil))
	}
}

func UpdateChannel() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		id := c.Param("id")
		if id == "" {
			klog.Warnf("update load channel request, id is empty")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "id cannot be empty", nil))
			return
		}
		if !utils.CheckUuid(id) {
			klog.Warnf("load channel id format is invalid. [%s]", id)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "load channel id is invalid", nil))
			return
		}
		var req LoadChannel
		if err := c.BindJSON(&req); err != nil {
			klog.Warnf("bind json failed, err: %v", err)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "Failed to parse request data. Check JSON format.", nil))
			return
		}

		if req.CaCert.ID != "" && !utils.CheckUuid(req.CaCert.ID) {
			klog.Warnf("ca cert id format is invalid. [%s]", req.CaCert.ID)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "ca cert id is invalid", nil))
			return
		}

		updateData := make(map[string]any)
		if req.Name != "" {
			updateData["name"] = req.Name
		}
		if req.Backend != "" {
			updateData["backend"] = req.Backend
		}
		if req.Timeout > 0 {
			updateData["timeout"] = req.Timeout
		}
		if req.HcInterval > 0 {
			updateData["hcinterval"] = req.HcInterval
		}
		if req.CaCert.ID != "" {
			updateData["ca_cert"] = req.CaCert.ID
		}

		if len(updateData) == 0 {
			klog.Warnf("update channel no valid fields to update, id: %s", id)
			c.JSON(http.StatusBadRequest, answer.ResBody(answer.EcodeInvalidRequestErr, "no valid fields to update", nil))
			return
		}

		dao, err := models.NewDBModel("SYSTEM")
		if err != nil {
			klog.Errorf("create db model failed, err: %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}

		// Check if the CA certificate exists
		lcca, err := dao.GetLcCaByID(id)
		if err != nil {
			klog.Errorf("GetLcCaByID query failed, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}
		if lcca == nil {
			klog.Warnf("lc ca record does not exist. id:[%s]", id)
			c.JSON(404, answer.ResBody(answer.EcodeNotFound, "`ca_cert.id` does not exist.", nil))
			return
		}

		lc, err := dao.GetLoadChannelByID(id)
		if err != nil {
			klog.Errorf("query load channel by id failed, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}
		if lc == nil {
			klog.Warnf("target load channel record does not exist, id:[%s]", id)
			c.JSON(404, answer.ResBody(answer.EcodeNotFound, "load channel record does not exist.", nil))
			return
		}
		if err := dao.UpdateLoadChannel(id, updateData); err != nil {
			klog.Errorf("UpdateLoadChannel db update failed, err: %v", err)
			c.JSON(http.StatusInternalServerError, answer.ResBody(answer.EcodeError, "failed to update load channel record in db.", nil))
			return
		}

		go func() {
			err := refreshLcCache(id)
			if err != nil {
				klog.Error("Cache refresh failed")
			}
		}()
		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, nil))
	}
}

func GetChannelDetail() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		id := c.Param("id")
		if id == "" {
			klog.Warnf("delete load channel request, id is empty")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "id cannot be empty", nil))
			return
		}
		if !utils.CheckUuid(id) {
			klog.Warnf("load channel id format is invalid. [%s]", id)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "load channel id is invalid", nil))
			return
		}

		dao, err := models.NewDBModel("SYSTEM")
		if err != nil {
			klog.Errorf("create db model failed, err: %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}

		lc, err := dao.GetLoadChannelByID(id)
		if err != nil {
			klog.Errorf("query load channel list failed, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}
		if lc == nil {
			klog.Warnf("target load channel record does not exist, id:[%s]", id)
			c.JSON(404, answer.ResBody(answer.EcodeNotFound, "lc does not exist.", nil))
			return
		}
		result := &LoadChannel{
			ID:         lc.ID,
			Name:       lc.Name,
			Backend:    lc.Backend,
			Timeout:    lc.Timeout,
			HcInterval: lc.HcInterval,
			Status:     lc.Status,
			CaCert:     CaCert{ID: lc.CaCert},
			CreateTime: lc.CreateTime,
			UpdateTime: lc.UpdateTime,
		}

		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, result))
	}
}

const (
	LcStateDisable int8 = 0
	LcStateEnable  int8 = 1
)

type LcStatusReq struct {
	Status *int8 `json:"status"`
}

func SetChannelStatus() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		id := c.Param("id")
		if id == "" {
			klog.Warnf("set load channel status request, id is empty")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "id cannot be empty", nil))
			return
		}
		if !utils.CheckUuid(id) {
			klog.Warnf("load channel id format is invalid. [%s]", id)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "load channel id is invalid", nil))
			return
		}
		var req LcStatusReq
		if err := c.Bind(&req); err != nil {
			klog.Warnf("bind request body failed, err: %v", err)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "Failed to parse request data. Check JSON format.", nil))
			return
		}
		if req.Status == nil {
			klog.Warn("request status field is empty")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "status can't be empty.", nil))
			return
		}

		status := *req.Status
		if status != LcStateDisable && status != LcStateEnable {
			klog.Warnf("invalid status value: [%d]", status)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "Wrong status value.", nil))
			return
		}

		dao, err := models.NewDBModel("SYSTEM")
		if err != nil {
			klog.Errorf("create db model failed, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}

		channel, err := dao.GetLoadChannelByID(id)
		if err != nil {
			klog.Errorf("query load channel by id failed, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}
		if channel == nil {
			klog.Warnf("target load channel record does not exist, id:[%s]", id)
			c.JSON(404, answer.ResBody(answer.EcodeNotFound, "load channel record does not exist.", nil))
			return
		}

		// The target value stays the same as the original value, not updating.
		if channel.Status == status {
			klog.Infof("channel %s already in target status %d, skip update", id, status)
			c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, "status no change", nil))
			return
		}

		rows, err := dao.SetLoadChannelStatusById(id, status)
		if err != nil {
			klog.Errorf("update load channel status in db failed, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "failed to update load channel status in db.", nil))
			return
		}
		if rows == 0 {
			klog.Warnf("target load channel record does not exist, id:[%s]", id)
			c.JSON(404, answer.ResBody(answer.EcodeNotFound, "load channel record does not exist.", nil))
			return
		}

		go func() {
			err := refreshLcCache(id)
			if err != nil {
				klog.Error("Cache refresh failed")
			}
		}()
		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, nil))
	}
}
