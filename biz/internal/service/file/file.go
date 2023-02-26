package file

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"os"
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/repo"
	U "personal-page-be/biz/internal/utils"
	"strconv"
)

type FileService struct {
	Repo repo.IRepository
}

func GenerateFileKey() string {
	return U.RandSeq(4)

}
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
func (s *FileService) UploadFile(ctx context.Context, c *app.RequestContext) {
	// single file
	count, err := strconv.Atoi(string(c.FormValue("count")))
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{"code": 2001, "msg": "count序列化失败"})
		return
	}
	if count == 0 {
		count = -1
	}

	fileKey := string(c.FormValue("file_key"))
	if fileKey == "" {
		fileKey = GenerateFileKey()
	}
	// 是否已经有文件

	findFileEntity, err := s.Repo.FindFile(fileKey)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{"code": 5001, "msg": "查询已有文件失败：" + err.Error()})
		return
	}
	if findFileEntity.ID != 0 {
		c.JSON(consts.StatusOK, utils.H{"code": 5002, "msg": "文件key已经存在，请更换"})
		return
	}

	file, _ := c.FormFile("file")
	if file == nil {
		c.JSON(consts.StatusOK, utils.H{"code": 2001, "msg": "无文件"})
		return
	}

	username := ctx.Value("username")
	user, err := s.Repo.FindUser(username.(string))
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
		return
	}

	//生成文件持久化数据

	fileEntity := domain.FileEntity{
		User:     *user,
		UserID:   int(user.ID),
		FileKey:  fileKey,
		FileName: file.Filename,
		Count:    count,
	}

	if err != nil {
		c.JSON(200, utils.H{"code": 2001, "msg": err.Error()})
		return
	}

	//保存文件
	err = c.SaveUploadedFile(file, fmt.Sprintf("./file/%s", fileEntity.FileKey))
	if err != nil {
		c.JSON(200, utils.H{"code": 5001, "msg": err.Error()})
		return
	}

	//保存文件模型
	err = s.Repo.SaveFile(&fileEntity)
	if err != nil {
		c.JSON(200, utils.H{"code": 5001, "msg": err.Error()})
		return
	}
	c.JSON(200, utils.H{"code": 0, "msg": "上传成功", "data": fileEntity})
	return
}

func (s *FileService) getFileInfo(fileKey string) (int, string, *domain.FileEntity) {
	if fileKey == "" {
		return 2001, "参数不合法", nil
	}

	fileEntity, err := s.Repo.FindFile(fileKey)
	if err != nil {
		return 5001, err.Error(), nil
	}
	if fileEntity.ID == 0 {
		return 4004, "未找到相关key的文件", nil
	}

	return 0, "", fileEntity
}
func (s *FileService) FileInfo(ctx context.Context, c *app.RequestContext) {
	code, msg, data := s.getFileInfo(c.DefaultQuery("file_key", ""))
	if code == 0 {
		c.JSON(consts.StatusOK, utils.H{
			"code": code,
			"data": data,
		})
	} else {
		c.JSON(consts.StatusOK, utils.H{
			"code": code,
			"msg":  msg,
		})
	}

}

func (s *FileService) DownloadFile(ctx context.Context, c *app.RequestContext) {
	code, msg, data := s.getFileInfo(c.DefaultQuery("file_key", ""))
	if code == 0 {

		filePath := fmt.Sprintf("./file/%s", data.FileKey)
		if !Exists(fmt.Sprintf(filePath)) {
			c.JSON(consts.StatusOK, utils.H{
				"code": 4004,
				"msg":  "文件记录存在，但文件数据已被删除",
			})
			return
		}
		c.FileAttachment(filePath, data.FileName)

		if data.Count > 0 {
			data.Count--
			if data.Count == 0 {
				s.removeFile(data)
			} else {
				s.Repo.SaveFile(data)
			}

		}

	} else {
		c.JSON(consts.StatusOK, utils.H{
			"code": code,
			"msg":  msg,
		})
	}
}

func (s *FileService) DeleteFile(ctx context.Context, c *app.RequestContext) {
	code, msg, data := s.getFileInfo(c.DefaultQuery("file_key", ""))

	if code == 0 {
		if data.UserID != s.getUserId(ctx) {
			c.JSON(consts.StatusOK, utils.H{
				"code": 4003,
				"msg":  "无权执行操作",
			})
			return
		}
		err := s.removeFile(data)
		if err != nil {
			c.JSON(consts.StatusOK, utils.H{
				"code": 5001,
				"msg":  "删除失败：" + err.Error(),
			})
			return
		}
		c.JSON(consts.StatusOK, utils.H{
			"code": 0,
			"msg":  "删除成功",
		})
		return
	} else {
		c.JSON(consts.StatusOK, utils.H{
			"code": code,
			"msg":  msg,
		})
	}
}
func (s *FileService) MyFiles(ctx context.Context, c *app.RequestContext) {
	c.JSON(consts.StatusOK, utils.H{
		"code": 0,
	})
}

func (s *FileService) getUserId(ctx context.Context) (uid int) {
	username := ctx.Value("username")
	user, err := s.Repo.FindUser(username.(string))
	if err != nil {

		return 0
	}
	return int(user.ID)
}
func (s *FileService) removeFile(entity *domain.FileEntity) error {

	err := s.Repo.RemoveFile(int(entity.ID))
	if err != nil {
		return err
	}
	if !Exists(fmt.Sprintf("./file/%s", entity.FileKey)) {
		return nil
	}
	filePath := fmt.Sprintf("./file/%s", entity.FileKey)
	return os.Remove(filePath)
}
