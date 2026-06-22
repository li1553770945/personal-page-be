package file

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/google/uuid"
	"github.com/tencentyun/cos-go-sdk-v5"
	"gorm.io/gorm"
	"personal-page-be/biz/internal/constant"
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/dto"
	"personal-page-be/biz/internal/response"
	U "personal-page-be/biz/internal/utils"
)

func GenerateFileKey() string {
	return U.RandSeq(8)
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func (s *FileService) UploadFile(ctx context.Context, c *app.RequestContext) {
	if fileHeader, err := c.FormFile("file"); err == nil && fileHeader != nil {
		s.uploadLocalFile(ctx, c, fileHeader)
		return
	}
	s.uploadSignedFile(ctx, c)
}

func (s *FileService) uploadSignedFile(ctx context.Context, c *app.RequestContext) {
	var req dto.SignedUploadFileReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, 2001, "参数错误: "+err.Error())
		return
	}
	if req.Name == "" {
		response.Error(c, 2001, "文件名不能为空")
		return
	}
	if req.Key != "" && !U.IsAlphanumeric(req.Key) {
		response.Error(c, 2001, "文件 key 只能包含大小写字母和数字")
		return
	}
	if req.Key == "" {
		for {
			req.Key = GenerateFileKey()
			exists, err := s.Repo.FindFileByKey(req.Key)
			if err != nil {
				response.Error(c, 5001, err.Error())
				return
			}
			if exists.ID == 0 {
				break
			}
		}
	} else {
		exists, err := s.Repo.FindFileByKey(req.Key)
		if err != nil {
			response.Error(c, 5001, err.Error())
			return
		}
		if exists.ID != 0 {
			response.Error(c, 5002, "文件 key 已经存在，请更换")
			return
		}
	}

	user, ok := s.requireCurrentUser(ctx, c)
	if !ok {
		return
	}
	entity := &domain.FileEntity{
		User:          *user,
		UserID:        int(user.ID),
		Name:          req.Name,
		Key:           req.Key,
		MaxDownload:   req.MaxDownload,
		DownloadCount: 0,
		OSSPath:       s.generateFilePath(req.Name),
	}
	if err := s.Repo.SaveFile(entity); err != nil {
		response.Error(c, 5001, "保存文件数据失败: "+err.Error())
		return
	}

	signedURL, err := s.getSignedURL(ctx, http.MethodPut, entity.OSSPath, entity.Name)
	if err != nil {
		response.Error(c, 5001, "生成上传 URL 失败: "+err.Error())
		return
	}
	response.OK(c, utils.H{
		"signedUrl": signedURL,
		"key":       entity.Key,
	}, "ok")
}

func (s *FileService) uploadLocalFile(ctx context.Context, c *app.RequestContext, fileHeader *multipart.FileHeader) {
	count, err := strconv.Atoi(string(c.FormValue("count")))
	if err != nil || count == 0 {
		count = -1
	}

	fileKey := string(c.FormValue("file_key"))
	if fileKey == "" {
		fileKey = GenerateFileKey()
	}
	if !U.IsAlphanumeric(fileKey) {
		response.Error(c, 2001, "文件 key 只能包含大小写字母和数字")
		return
	}

	findFileEntity, err := s.Repo.FindFileByFileKey(fileKey)
	if err != nil {
		response.Error(c, 5001, "查询已有文件失败: "+err.Error())
		return
	}
	if findFileEntity.ID != 0 {
		response.Error(c, 5002, "文件 key 已经存在，请更换")
		return
	}

	user, ok := s.requireCurrentUser(ctx, c)
	if !ok {
		return
	}

	saveName := fmt.Sprintf("%s_%s_%s", time.Now().Format("2006-01-02_15_04_05"), fileKey, uuid.New().String())
	fileEntity := domain.FileEntity{
		User:     *user,
		UserID:   int(user.ID),
		FileKey:  fileKey,
		FileName: fileHeader.Filename,
		Count:    count,
		SaveName: saveName,
	}
	if err = c.SaveUploadedFile(fileHeader, fmt.Sprintf("./%s/%s", constant.FileBasePath, fileEntity.SaveName)); err != nil {
		response.Error(c, 5001, "保存文件失败: "+err.Error())
		return
	}
	if err = s.Repo.SaveFile(&fileEntity); err != nil {
		response.Error(c, 5001, "保存文件数据失败: "+err.Error())
		return
	}
	response.OK(c, fileEntity, "上传成功")
}

func (s *FileService) FileInfo(ctx context.Context, c *app.RequestContext) {
	key := queryKey(c)
	if key == "" {
		response.Error(c, 2001, "缺少文件 key")
		return
	}

	entity, err := s.findFileByAnyKey(key)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if entity.ID == 0 {
		response.Error(c, 4004, "未找到对应文件")
		return
	}
	name := entity.Name
	if name == "" {
		name = entity.FileName
	}
	response.OK(c, utils.H{
		"name":      name,
		"key":       key,
		"size":      0,
		"createdAt": entity.CreatedAt.Unix(),
	}, "ok")
}

func (s *FileService) DownloadSignedFile(ctx context.Context, c *app.RequestContext) {
	key := queryKey(c)
	if key == "" {
		response.Error(c, 2001, "缺少文件 key")
		return
	}

	entity, err := s.Repo.FindFileByKey(key)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if entity.ID == 0 {
		response.Error(c, 4004, "未找到对应文件")
		return
	}

	signedURL, err := s.getSignedURL(ctx, http.MethodGet, entity.OSSPath, entity.Name)
	if err != nil {
		response.Error(c, 5001, "生成下载 URL 失败: "+err.Error())
		return
	}
	if entity.MaxDownload != 0 {
		entity.DownloadCount++
		if entity.DownloadCount >= entity.MaxDownload {
			entity.ExpiredTime = gorm.DeletedAt{Time: time.Now(), Valid: true}
			_ = s.deleteCOSObject(ctx, entity.OSSPath)
			_ = s.Repo.RemoveFileByKey(entity.Key)
		} else if err = s.Repo.SaveFile(entity); err != nil {
			response.Error(c, 5001, "更新下载次数失败: "+err.Error())
			return
		}
	}
	response.OK(c, utils.H{
		"signedUrl": signedURL,
		"name":      entity.Name,
	}, "ok")
}

func (s *FileService) DownloadFile(ctx context.Context, c *app.RequestContext) {
	key := queryKey(c)
	if key == "" {
		response.Error(c, 2001, "缺少文件 key")
		return
	}
	entity, err := s.findFileByAnyKey(key)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if entity.ID == 0 || entity.SaveName == "" {
		response.Error(c, 4004, "未找到可直出的本地文件")
		return
	}
	filePath := fmt.Sprintf("./%s/%s", constant.FileBasePath, entity.SaveName)
	if !Exists(filePath) {
		response.Error(c, 4004, "文件记录存在，但文件数据已被删除")
		return
	}
	c.FileAttachment(filePath, entity.FileName)
	if entity.Count > 0 {
		entity.Count--
		if err = s.Repo.SaveFile(entity); err != nil {
			s.Logger.Error("update file count failed: " + err.Error())
		}
	}
}

func (s *FileService) ListMyFiles(ctx context.Context, c *app.RequestContext) {
	user, ok := s.requireCurrentUser(ctx, c)
	if !ok {
		return
	}
	userID := user.ID
	files, err := s.Repo.ListFiles(&userID)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	response.OK(c, fileEntitiesToDTO(files), "ok")
}

func (s *FileService) ListAllFiles(ctx context.Context, c *app.RequestContext) {
	user, ok := s.requireCurrentUser(ctx, c)
	if !ok {
		return
	}
	if !domain.IsSuperAdminRole(user.Role) {
		response.Error(c, 4003, "无权执行此操作")
		return
	}
	files, err := s.Repo.ListFiles(nil)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	response.OK(c, fileEntitiesToDTO(files), "ok")
}

func (s *FileService) DeleteFile(ctx context.Context, c *app.RequestContext) {
	if id := c.Param("id"); id != "" {
		s.deleteFileByID(ctx, c, id)
		return
	}

	var req dto.DeleteFileReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, 2001, "参数错误: "+err.Error())
		return
	}
	if req.Key == "" {
		response.Error(c, 2001, "缺少文件 key")
		return
	}
	entity, err := s.findFileByAnyKey(req.Key)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if entity.ID == 0 {
		response.Error(c, 4004, "未找到对应文件")
		return
	}
	user, ok := s.requireCurrentUser(ctx, c)
	if !ok {
		return
	}
	if !canManageFile(user, entity) {
		response.Error(c, 4003, "无权执行此操作")
		return
	}
	if entity.OSSPath != "" {
		if err = s.deleteCOSObject(ctx, entity.OSSPath); err != nil {
			response.Error(c, 5001, "删除 COS 文件失败: "+err.Error())
			return
		}
		entity.DeleteOnOssTime = gorm.DeletedAt{Time: time.Now(), Valid: true}
		_ = s.Repo.SaveFile(entity)
	}
	if err = s.Repo.RemoveFile(entity.ID); err != nil {
		response.Error(c, 5001, "删除文件数据失败: "+err.Error())
		return
	}
	response.OK(c, nil, "删除成功")
}

func (s *FileService) deleteFileByID(ctx context.Context, c *app.RequestContext, fileID string) {
	fileIDInt, err := strconv.ParseUint(fileID, 10, 64)
	if err != nil {
		response.Error(c, 2001, "参数错误")
		return
	}
	file, err := s.Repo.FindFileByID(uint(fileIDInt))
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if file.ID == 0 {
		response.Error(c, 4004, "未找到对应文件")
		return
	}
	user, ok := s.requireCurrentUser(ctx, c)
	if !ok {
		return
	}
	if !canManageFile(user, file) {
		response.Error(c, 4003, "无权执行此操作")
		return
	}
	if file.OSSPath != "" {
		_ = s.deleteCOSObject(ctx, file.OSSPath)
		file.DeleteOnOssTime = gorm.DeletedAt{Time: time.Now(), Valid: true}
		_ = s.Repo.SaveFile(file)
	}
	if err = s.Repo.RemoveFile(file.ID); err != nil {
		response.Error(c, 5001, "删除失败: "+err.Error())
		return
	}
	response.OK(c, nil, "删除成功")
}

func (s *FileService) findFileByAnyKey(key string) (*domain.FileEntity, error) {
	entity, err := s.Repo.FindFileByKey(key)
	if err != nil || entity.ID != 0 {
		return entity, err
	}
	return s.Repo.FindFileByFileKey(key)
}

func queryKey(c *app.RequestContext) string {
	if key := c.DefaultQuery("key", ""); key != "" {
		return key
	}
	return c.DefaultQuery("file-key", "")
}

func (s *FileService) generateFilePath(fileName string) string {
	now := time.Now()
	return fmt.Sprintf("%d/%d/%d/%s%s", now.Year(), now.Month(), now.Day(), uuid.New().String(), filepath.Ext(fileName))
}

func (s *FileService) getSignedURL(ctx context.Context, method, ossPath, name string) (string, error) {
	client, err := s.cosClient()
	if err != nil {
		return "", err
	}
	opt := &cos.PresignedURLOptions{
		Query:  &url.Values{},
		Header: &http.Header{},
	}
	if method == http.MethodGet {
		opt.Query.Add("response-content-disposition", fmt.Sprintf("attachment; filename=%s", name))
	}
	signedURL, err := client.Object.GetPresignedURL(ctx, method, ossPath, s.Config.EffectiveCOSAk(), s.Config.EffectiveCOSSk(), time.Hour, opt)
	if err != nil {
		return "", err
	}
	return signedURL.String(), nil
}

func (s *FileService) deleteCOSObject(ctx context.Context, ossPath string) error {
	if ossPath == "" {
		return nil
	}
	client, err := s.cosClient()
	if err != nil {
		return err
	}
	_, err = client.Object.Delete(ctx, ossPath)
	return err
}

func (s *FileService) cosClient() (*cos.Client, error) {
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
	return cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  ak,
			SecretKey: sk,
		},
	}), nil
}

func (s *FileService) getUserId(ctx context.Context) (uid int) {
	username, _ := ctx.Value("username").(string)
	user, err := s.Repo.FindUser(username)
	if err != nil {
		return 0
	}
	return int(user.ID)
}

func (s *FileService) requireCurrentUser(ctx context.Context, c *app.RequestContext) (*domain.UserEntity, bool) {
	username, _ := ctx.Value("username").(string)
	user, err := s.Repo.FindUser(username)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return nil, false
	}
	if user.ID == 0 || !user.CanUse {
		response.Error(c, 4003, "未登录")
		return nil, false
	}
	user.Role = domain.NormalizeRole(user.Role)
	return user, true
}

func canManageFile(user *domain.UserEntity, file *domain.FileEntity) bool {
	return uint(file.UserID) == user.ID || domain.IsAdminRole(user.Role)
}

func fileEntitiesToDTO(files *[]domain.FileEntity) []*dto.FileDTO {
	result := make([]*dto.FileDTO, 0, len(*files))
	for i := range *files {
		file := &(*files)[i]
		name := file.Name
		if name == "" {
			name = file.FileName
		}
		key := file.Key
		kind := "object"
		if key == "" {
			key = file.FileKey
			kind = "local"
		}
		result = append(result, &dto.FileDTO{
			ID:            file.ID,
			UserID:        file.UserID,
			Username:      file.User.Username,
			Nickname:      file.User.Nickname,
			Name:          name,
			Key:           key,
			Kind:          kind,
			Count:         file.Count,
			MaxDownload:   file.MaxDownload,
			DownloadCount: file.DownloadCount,
			CreatedAt:     file.CreatedAt.Unix(),
			UpdatedAt:     file.UpdatedAt.Unix(),
		})
	}
	return result
}
