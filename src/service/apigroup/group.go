package apigroup

import (
	"apigw/src/models"
	"apigw/src/pkg/answer"
	"apigw/src/pkg/common"
	"apigw/src/pkg/utils"
	"apigw/src/slog"
	"context"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/cloudwego/hertz/pkg/app"
)

type ApiGroup struct {
	ID         string  `json:"id"`          // uuid
	Name       string  `json:"name"`        // 分组名称
	Remark     *string `json:"remark"`      // 分组描述
	CreateTime int64   `json:"create_time"` // 创建时间戳
	UpdateTime int64   `json:"update_time"` // 更新时间戳
}

// Check if it's an invalid name
func checkApiGroupNameNotValid(ctx context.Context, c *app.RequestContext, name string) bool {
	klog := slog.FromCtx(ctx)
	s := strings.TrimSpace(name)
	if strings.Contains(name, " ") {
		klog.Warnf("group name contains space: %s", name)
		c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "The name cannot contain any spaces.", nil))
		return true
	}
	if s == "" {
		klog.Warn("create api group name is empty")
		c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "Name can't be empty.", nil))
		return true
	}
	charLen := utf8.RuneCountInString(s)
	if charLen < 3 || charLen > 20 {
		klog.Warnf("group name length illegal, length: %d", len(s))
		c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "Group name length range [3,20]", nil))
		return true
	}
	return false
}

func GroupList() func(ctx context.Context, c *app.RequestContext) {
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

		list, count, err := dao.ListApiGroup(limit, offset)
		if err != nil {
			klog.Error("FindItemAll err:: ", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}

		result := make([]*ApiGroup, len(list))
		for k, v := range list {
			name := v.Name
			remark := v.Remark
			item := &ApiGroup{
				ID:         v.ID,
				Name:       name,
				Remark:     &remark,
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

func GetApiGroupDetail() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		id := c.Param("id")
		if id == "" {
			klog.Warn("query api group detail id is empty")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "api group id cannot be empty", nil))
			return
		}
		if !utils.CheckUuid(id) {
			klog.Warnf("api group id is invalid. [%s]", id)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "api group id is invalid", nil))
			return
		}

		dao, err := models.NewDBModel("SYSTEM")
		if err != nil {
			klog.Errorf("NewDBModel error %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}

		group, err := dao.GetApiGroupByID(id)
		if err != nil {
			klog.Errorf("GetApiGroupByID, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}
		if group == nil {
			klog.Warnf("api group does not exist. [%s]", id)
			c.JSON(400, answer.ResBody(answer.EcodeNotFound, "api group does not exist.", nil))
			return
		}

		result := &ApiGroup{
			ID:         group.ID,
			Name:       group.Name,
			Remark:     &group.Remark,
			CreateTime: group.CreateTime,
			UpdateTime: group.UpdateTime,
		}

		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, result))
	}
}

func CreateApiGroup() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		var req ApiGroup
		if err := c.BindJSON(&req); err != nil {
			klog.Warnf("bind json failed, err: %v", err)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "Failed to parse request data. Check JSON format.", nil))
			return
		}

		if checkApiGroupNameNotValid(ctx, c, req.Name) {
			return
		}

		if req.Remark == nil {
			s := ""
			req.Remark = &s
		}

		dao, err := models.NewSystemDBModel()
		if err != nil {
			klog.Errorf("NewDBModel err: %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}

		group, err := dao.GetApiGroupByName(req.Name)
		if err != nil {
			klog.Errorf("GetApiGroupByName err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}
		if group != nil {
			klog.Warnf("The group name already exists，id: %s", group.ID)
			c.JSON(400, answer.ResBody(answer.EcodeApiGroupDuplicate, "Group name is duplicated.", nil))
			return
		}

		t := common.CreateTimestamp()
		id := utils.CreateUuid()
		cate := &models.OrmApiGroup{
			ID:         id,
			Name:       req.Name,
			Remark:     *req.Remark,
			Status:     1,
			CreateTime: t,
			UpdateTime: t,
		}
		err = dao.ApiGroup.Create(cate)
		if err != nil {
			klog.Errorf("create group to db error: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "create group to db error.", nil))
			return
		}

		req.ID = id
		req.CreateTime = t
		req.UpdateTime = t
		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, req))
	}
}

func UpdateApiGroup() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if err := recover(); err != nil {
				klog := slog.FromCtx(ctx)
				klog.Errorf("UpdateApiGroup panic recover, err: %v", err)
				c.JSON(http.StatusInternalServerError, answer.ResBody(answer.EcodeError, "server internal panic", nil))
			}
		}()

		klog := slog.FromCtx(ctx)
		id := c.Param("id")
		if !utils.CheckUuid(id) {
			errMsg := "api group id is invalid"
			klog.Errorf("%s. [%s]", errMsg, id)
			c.JSON(http.StatusBadRequest, answer.ResBody(answer.EcodeInvalidApiId, errMsg, nil))
			return
		}

		var req ApiGroup
		if err := c.BindJSON(&req); err != nil {
			klog.Errorf("bind json failed. err: %v", err)
			errMsg := "Failed to parse request data. Check JSON format."
			c.JSON(http.StatusBadRequest, answer.ResBody(answer.EcodeInvalidRequestErr, errMsg, nil))
			return
		}

		updateData := make(map[string]any)
		// Skip updating the name field if it's empty
		if req.Name != "" {
			if checkApiGroupNameNotValid(ctx, c, req.Name) {
				return
			}
			updateData["name"] = req.Name
		}
		if req.Remark != nil {
			updateData["remark"] = *req.Remark
		}

		if len(updateData) == 0 {
			c.JSON(http.StatusBadRequest, answer.ResBody(answer.EcodeInvalidRequestErr, "no field needs to be updated", nil))
			return
		}

		dao, err := models.NewDBModel("SYSTEM")
		if err != nil {
			klog.Errorf("NewDBModel err: %s", err)
			c.JSON(http.StatusInternalServerError, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}

		targetGroup, err := dao.GetApiGroupByID(id)
		if err != nil {
			klog.Errorf("GetApiGroupByID err: %v", err)
			c.JSON(http.StatusInternalServerError, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}
		if targetGroup == nil {
			errMsg := "api group does not exist"
			klog.Errorf("%s. [%s]", errMsg, id)
			c.JSON(http.StatusBadRequest, answer.ResBody(answer.EcodeNotFound, errMsg, nil))
			return
		}

		if req.Name != "" {
			existGroup, err := dao.GetApiGroupByName(req.Name)
			if err != nil {
				klog.Errorf("GetApiGroupByName err: %v", err)
				c.JSON(http.StatusInternalServerError, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
				return
			}
			if existGroup != nil && existGroup.ID != targetGroup.ID {
				errMsg := "Group name is duplicated"
				klog.Errorf("%s, exist id: %s, current id: %s", errMsg, existGroup.ID, targetGroup.ID)
				c.JSON(http.StatusBadRequest, answer.ResBody(answer.EcodeApiGroupDuplicate, errMsg, nil))
				return
			}
		}

		if err := dao.UpdateApiGroup(id, updateData); err != nil {
			klog.Errorf("UpdateApiGroup err: %v", err)
			c.JSON(http.StatusInternalServerError, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}

		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, nil))
	}
}

func DeleteApiGroup() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		id := c.Param("id")
		if id == "" {
			klog.Warnf("delete api group empty id")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "id cannot be empty", nil))
			return
		}
		if !utils.CheckUuid(id) {
			klog.Warnf("api group id is invalid. [%s]", id)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "api group id is invalid", nil))
			return
		}

		dao, err := models.NewDBModel("SYSTEM")
		if err != nil {
			klog.Errorf("NewDBModel error %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}

		// Check if the API group exists
		group, err := dao.GetApiGroupByID(id)
		if err != nil {
			klog.Errorf("GetApiGroupByID, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}
		if group == nil {
			klog.Warnf("api group does not exist. [%s]", id)
			c.JSON(404, answer.ResBody(answer.EcodeNotFound, "api group does not exist.", nil))
			return
		}

		// Check if the API group is already used
		api, err := dao.ListApiInterfaceByGroupID(id)
		if err != nil {
			klog.Errorf("ListApiInterfaceByGroupID, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}
		if len(api) > 0 {
			klog.Warn("The API group is in use and can't be deleted.")
			c.JSON(409, answer.ResBody(answer.EcodeGroupInUse, "The API group is in use and can't be deleted.", nil))
			return
		}

		if err := dao.DeleteApiGroup(id); err != nil {
			klog.Errorf("Error deleting api group, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "error deleting api group from the database.", nil))
			return
		}

		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, nil))
	}
}
