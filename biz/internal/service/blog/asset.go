package blog

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	"github.com/tencentyun/cos-go-sdk-v5"
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/dto"
	"personal-page-be/biz/internal/response"
)

const maxBlogAssetSize = 15 * 1024 * 1024

func (s *BlogService) SignAssetUpload(ctx context.Context, c *app.RequestContext) {
	user, ok := s.requireSuperAdmin(ctx, c)
	if !ok {
		return
	}
	var req dto.SignBlogAssetReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, 2001, "参数错误: "+err.Error())
		return
	}
	req.FileName = strings.TrimSpace(req.FileName)
	req.MimeType = strings.ToLower(strings.TrimSpace(req.MimeType))
	if req.FileName == "" || !strings.HasPrefix(req.MimeType, "image/") {
		response.Error(c, 2001, "只允许上传图片")
		return
	}
	if req.Size <= 0 || req.Size > maxBlogAssetSize {
		response.Error(c, 2001, "图片大小必须在 15 MiB 以内")
		return
	}
	ext := strings.ToLower(filepath.Ext(req.FileName))
	if ext == "" || len(ext) > 10 {
		ext = ".bin"
	}
	now := time.Now()
	objectPath := fmt.Sprintf("blog/%04d/%02d/%s%s", now.Year(), int(now.Month()), uuid.NewString(), ext)
	asset := &domain.BlogAssetEntity{
		PostID: req.PostID, ObjectPath: objectPath, FileName: req.FileName,
		MimeType: req.MimeType, Size: req.Size, CreatedBy: user.ID,
	}
	if err := s.Repo.SaveBlogAsset(asset); err != nil {
		response.Error(c, 5001, "保存图片记录失败: "+err.Error())
		return
	}
	signedURL, err := s.signCOSObjectURL(ctx, http.MethodPut, objectPath)
	if err != nil {
		response.Error(c, 5001, "生成上传地址失败: "+err.Error())
		return
	}
	result := blogAssetToDTO(asset)
	result.SignedURL = signedURL
	response.OK(c, result, "ok")
}

func (s *BlogService) ConfirmAssetUpload(ctx context.Context, c *app.RequestContext) {
	if _, ok := s.requireSuperAdmin(ctx, c); !ok {
		return
	}
	assetID, ok := parseIDParam(c, "assetId")
	if !ok {
		return
	}
	asset, err := s.Repo.FindBlogAssetByID(assetID)
	if err != nil || asset.ID == 0 {
		response.Error(c, 4004, "图片记录不存在")
		return
	}
	var req dto.ConfirmBlogAssetReq
	if err = c.BindAndValidate(&req); err != nil {
		response.Error(c, 2001, "参数错误: "+err.Error())
		return
	}
	client, err := s.cosClient()
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	resp, err := client.Object.Head(ctx, asset.ObjectPath, nil)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		response.Error(c, 4004, "COS 中未找到已上传图片")
		return
	}
	asset.Width = req.Width
	asset.Height = req.Height
	asset.Alt = strings.TrimSpace(req.Alt)
	asset.Ready = true
	if err = s.Repo.SaveBlogAsset(asset); err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	response.OK(c, blogAssetToDTO(asset), "上传完成")
}

func (s *BlogService) ServeAsset(ctx context.Context, c *app.RequestContext) {
	value, err := strconv.ParseUint(c.Param("assetId"), 10, 64)
	if err != nil || value == 0 {
		c.Data(consts.StatusNotFound, "text/plain; charset=utf-8", []byte("asset not found"))
		return
	}
	asset, err := s.Repo.FindBlogAssetByID(uint(value))
	if err != nil || asset.ID == 0 || !asset.Ready {
		c.Data(consts.StatusNotFound, "text/plain; charset=utf-8", []byte("asset not found"))
		return
	}
	signedURL, err := s.signCOSObjectURL(ctx, http.MethodGet, asset.ObjectPath)
	if err != nil {
		c.Data(consts.StatusInternalServerError, "text/plain; charset=utf-8", []byte("asset unavailable"))
		return
	}
	c.Redirect(consts.StatusFound, []byte(signedURL))
	c.Response.Header.Set("Cache-Control", "public, max-age=300")
}

func blogAssetToDTO(asset *domain.BlogAssetEntity) *dto.BlogAssetDTO {
	return &dto.BlogAssetDTO{
		ID: asset.ID, PostID: asset.PostID, ObjectPath: asset.ObjectPath,
		FileName: asset.FileName, MimeType: asset.MimeType, Size: asset.Size,
		Width: asset.Width, Height: asset.Height, Alt: asset.Alt,
		URL: fmt.Sprintf("/api/blog/assets/%d", asset.ID), Ready: asset.Ready,
	}
}

func (s *BlogService) signCOSObjectURL(ctx context.Context, method, objectPath string) (string, error) {
	client, err := s.cosClient()
	if err != nil {
		return "", err
	}
	signedURL, err := client.Object.GetPresignedURL(ctx, method, objectPath, s.Config.EffectiveCOSAk(), s.Config.EffectiveCOSSk(), time.Hour, &cos.PresignedURLOptions{
		Query: &url.Values{}, Header: &http.Header{},
	})
	if err != nil {
		return "", err
	}
	return signedURL.String(), nil
}

func (s *BlogService) cosClient() (*cos.Client, error) {
	endpoint := s.Config.EffectiveCOSEndpoint()
	ak := s.Config.EffectiveCOSAk()
	sk := s.Config.EffectiveCOSSk()
	if endpoint == "" || ak == "" || sk == "" {
		return nil, fmt.Errorf("COS config is incomplete")
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	return cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{Transport: &cos.AuthorizationTransport{SecretID: ak, SecretKey: sk}}), nil
}
