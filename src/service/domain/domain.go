package domain

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

type ApiDomain struct {
	ID         string `json:"id"`          // 分组id
	Name       string `json:"name"`        // 域名
	Status     int8   `json:"status"`      // 状态（0-禁用，1-启用）
	CreateTime int64  `json:"create_time"` // 创建时间戳
	UpdateTime int64  `json:"update_time"` // 更新时间戳
}

func ListDomain() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)

		groupId := c.Param("groupId")
		if !utils.CheckUuid(groupId) {
			errMsg := "group id is invalid"
			klog.Errorf("%s. [%s]", errMsg, groupId)
			c.JSON(http.StatusBadRequest, answer.ResBody(answer.EcodeInvalidRequestErr, errMsg, nil))
			return
		}

		dao, err := models.NewSystemDBModel()
		if err != nil {
			klog.Errorf("NewDBModel err: %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}

		list, err := dao.Domain.ListByGroupId(groupId)
		if err != nil {
			klog.Error("ListApiDomainByGroupID err: ", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}
		result := make([]*ApiDomain, len(list))
		for k, v := range list {
			item := &ApiDomain{
				ID:         v.ID,
				Name:       v.Name,
				Status:     v.Status,
				CreateTime: v.CreateTime,
				UpdateTime: v.UpdateTime,
			}
			result[k] = item
		}

		payload := map[string]any{
			"items": result,
		}

		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, payload))
	}
}
func BindDomain() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		var req ApiDomain
		if err := c.BindJSON(&req); err != nil {
			klog.Warnf("bind json failed, err: %v", err)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "Failed to parse request data. Check JSON format.", nil))
			return
		}
		groupId := c.Param("groupId")
		if groupId == "" {
			klog.Warnf("bind domain groupId empty")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "groupId cannot be empty", nil))
			return
		}
		if !utils.CheckUuid(groupId) {
			klog.Warnf("bind domain groupId is invalid. [%s]", groupId)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "groupId is invalid", nil))
			return
		}
		dao, err := models.NewSystemDBModel()
		if err != nil {
			klog.Errorf("NewDBModel err: %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}

		list, err := dao.Domain.ListByGroupId(groupId)
		if err != nil {
			klog.Error("ListApiDomainByGroupID err: ", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}

		// 最多绑定8个独立域名
		if len(list) >= 8 {
			klog.Warn("the number of bound domains reaches the upper limit of 8")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "the group can only bind up to 8 independent domains", nil))
			return
		}

		t := common.CreateTimestamp()
		domainID := utils.CreateUuid()
		data := &models.OrmApiDomain{
			ID:         domainID,
			GroupID:    groupId,
			Name:       req.Name,
			Status:     1,
			CreateTime: t,
			UpdateTime: t,
		}
		err = dao.Domain.Create(data)
		if err != nil {
			klog.Errorf("create domain to db error: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "create domain to db error.", nil))
			return
		}

		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, nil))
	}
}

func UnbindDomain() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		groupId := c.Param("groupId")
		if groupId == "" {
			klog.Warnf("unbind domain groupId empty")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "groupId cannot be empty", nil))
			return
		}
		if !utils.CheckUuid(groupId) {
			klog.Warnf("unbind domain groupId is invalid. [%s]", groupId)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "groupId is invalid", nil))
			return
		}

		id := c.Param("id")
		if id == "" {
			klog.Warnf("unbind domain id empty")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "id cannot be empty", nil))
			return
		}
		if !utils.CheckUuid(id) {
			klog.Warnf("domain id is invalid. [%s]", id)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "domain id is invalid", nil))
			return
		}

		dao, err := models.NewSystemDBModel()
		if err != nil {
			klog.Errorf("NewDBModel err: %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}

		rows, err := dao.Domain.Delete(id)
		if err != nil {
			klog.Errorf("DeleteApiDomain database delete failed, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "error deleting domain record from the database.", nil))
			return
		}
		if rows == 0 {
			klog.Warnf("target domain record does not exist, id:[%s]", id)
			c.JSON(404, answer.ResBody(answer.EcodeNotFound, "domain record does not exist.", nil))
			return
		}

		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, nil))
	}
}
