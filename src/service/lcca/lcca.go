package lcca

import (
	"apigw/src/models"
	"apigw/src/pkg/answer"
	"apigw/src/pkg/common"
	"apigw/src/pkg/utils"
	"apigw/src/slog"
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

type LcCa struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Cert       Cert   `json:"cert"`        // 后端CA证书
	CreateTime int64  `json:"create_time"` // 创建时间戳
	UpdateTime int64  `json:"update_time"` // 更新时间戳
}

type Cert struct {
	CN        string `json:"CN"`                // 证书CN
	NotBefore int64  `json:"not_before"`        // 证书生效起始时间戳
	NotAfter  int64  `json:"not_after"`         // 证书过期结束时间戳
	Content   string `json:"content,omitempty"` // 证书文件base64编码内容
}

func ParseBase64PemCert(ca string) (*x509.Certificate, error) {
	cert, err := utils.Base64ToCert(ca)
	if err != nil {
		return nil, err
	}

	block, rest := pem.Decode(cert)
	if block == nil {
		return nil, errors.New("no valid pem certificate block found, invalid content")
	}

	if len(rest) > 0 {
		return nil, errors.New("extra irrelevant data exists in pem, only single certificate is supported")
	}

	obj, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse x509 certificate: %w", err)
	}

	return obj, nil
}

func CreateLcCa() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		var req LcCa
		if err := c.BindJSON(&req); err != nil {
			klog.Warnf("bind json failed, err: %v", err)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "Failed to parse request data. Check JSON format.", nil))
			return
		}

		dao, err := models.NewSystemDBModel()
		if err != nil {
			klog.Errorf("NewDBModel err: %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}

		obj, err := ParseBase64PemCert(req.Cert.Content)
		if err != nil {
			klog.Errorf("parse cert error: %v", err)
			c.JSON(400, answer.ResBody(answer.EcodeError, "parse cert error.", nil))
			return
		}

		t := common.CreateTimestamp()
		id := utils.CreateUuid()
		data := &models.OrmLcCa{
			ID:         id,
			Name:       req.Name,
			CN:         obj.Subject.CommonName,
			NotBefore:  obj.NotBefore.UnixMilli(),
			NotAfter:   obj.NotAfter.UnixMilli(),
			Cert:       req.Cert.Content,
			CreateTime: t,
			UpdateTime: t,
		}

		err = dao.LcCa.CreateLcCa(data)
		if err != nil {
			klog.Errorf("create lcca to db error: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "create lcca to db error.", nil))
			return
		}

		req.ID = id
		req.Cert.CN = obj.Subject.CommonName
		req.Cert.NotBefore = obj.NotBefore.UnixMilli()
		req.Cert.NotAfter = obj.NotAfter.UnixMilli()
		req.Cert.Content = ""
		req.CreateTime = t
		req.UpdateTime = t

		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, req))

	}
}

func ListLcCa() func(ctx context.Context, c *app.RequestContext) {
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

		list, count, err := dao.ListLcCa(limit, offset)
		if err != nil {
			klog.Error("FindItemAll err:: ", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}

		result := make([]*LcCa, len(list))
		for k, v := range list {
			cert := Cert{
				CN:        v.CN,
				NotBefore: v.NotBefore,
				NotAfter:  v.NotAfter,
			}
			item := &LcCa{
				ID:         v.ID,
				Name:       v.Name,
				Cert:       cert,
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

func DeleteLcCa() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		id := c.Param("id")
		if id == "" {
			klog.Warnf("delete lcca id empty")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "id cannot be empty", nil))
			return
		}
		if !utils.CheckUuid(id) {
			klog.Warnf("lcca id is invalid. [%s]", id)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "lcca id is invalid", nil))
			return
		}

		dao, err := models.NewDBModel("SYSTEM")
		if err != nil {
			klog.Errorf("NewDBModel error %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}
		lcca, err := dao.GetLcCaByID(id)
		if err != nil {
			klog.Errorf("GetLcCaByID query failed, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}
		if lcca == nil {
			klog.Warnf("lc ca record does not exist. id:[%s]", id)
			c.JSON(404, answer.ResBody(answer.EcodeNotFound, "lc ca record does not exist.", nil))
			return
		}

		lc, err := dao.GetLoadChannelByLcCaID(id)
		if err != nil {
			klog.Errorf("GetLoadChannelByLccaID query failed, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}
		if len(lc) > 0 {
			klog.Warn("The lc ca is associated with load channels and cannot be deleted.")
			c.JSON(409, answer.ResBody(answer.EcodeGroupInUse, "The lc ca is in use and can't be deleted.", nil))
			return
		}

		if err := dao.DeleteLcCa(id); err != nil {
			klog.Errorf("DeleteLcCa database delete failed, err: %v", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "error deleting lc ca record from the database.", nil))
			return
		}

		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, nil))
	}
}

func UpdateLcCa() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		id := c.Param("id")
		if !utils.CheckUuid(id) {
			errMsg := "lc ca id is invalid"
			klog.Errorf("%s. [%s]", errMsg, id)
			c.JSON(http.StatusBadRequest, answer.ResBody(answer.EcodeInvalidApiId, errMsg, nil))
			return
		}
		var req LcCa
		if err := c.BindJSON(&req); err != nil {
			klog.Warnf("bind json failed, err: %v", err)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidRequestErr, "Failed to parse request data. Check JSON format.", nil))
			return
		}
		updateData := make(map[string]any)
		if req.Name != "" {
			updateData["name"] = req.Name
		}

		if req.Cert.Content != "" {
			obj, err := ParseBase64PemCert(req.Cert.Content)
			if err != nil {
				klog.Errorf("parse certificate content error: %v", err)
				c.JSON(400, answer.ResBody(answer.EcodeError, "Failed to parse certificate content.", nil))
				return
			}
			updateData["CN"] = obj.Subject.CommonName
			updateData["NotBefore"] = obj.NotBefore.UnixMilli()
			updateData["NotAfter"] = obj.NotAfter.UnixMilli()
			updateData["Cert"] = req.Cert.Content
		}

		dao, err := models.NewDBModel("SYSTEM")
		if err != nil {
			klog.Errorf("NewDBModel err: %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}

		if err := dao.UpdateLcCa(id, updateData); err != nil {
			klog.Errorf("UpdateLcCa database update failed, err: %v", err)
			c.JSON(http.StatusInternalServerError, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}

		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, nil))
	}
}

func GetLcCaDetail() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		id := c.Param("id")
		if id == "" {
			klog.Warnf("delete lcca id empty")
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "id cannot be empty", nil))
			return
		}
		if !utils.CheckUuid(id) {
			klog.Warnf("lcca id is invalid. [%s]", id)
			c.JSON(400, answer.ResBody(answer.EcodeInvalidApiId, "lcca id is invalid", nil))
			return
		}

		dao, err := models.NewDBModel("SYSTEM")
		if err != nil {
			klog.Errorf("NewDBModel err: %s", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, err.Error(), nil))
			return
		}

		cert, err := dao.GetLcCaByID(id)
		if err != nil {
			klog.Error("FindItemAll err:: ", err)
			c.JSON(500, answer.ResBody(answer.EcodeError, "Internal server error.", nil))
			return
		}
		result := &LcCa{
			ID:   cert.ID,
			Name: cert.Name,
			Cert: Cert{
				CN:        cert.CN,
				NotBefore: cert.NotBefore,
				NotAfter:  cert.NotAfter,
				Content:   cert.Cert,
			},
			CreateTime: cert.CreateTime,
			UpdateTime: cert.UpdateTime,
		}
		c.JSON(http.StatusOK, answer.ResBody(answer.EcodeOK, nil, result))
	}
}
