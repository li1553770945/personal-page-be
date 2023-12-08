package file

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"os"
	"personal-page-be/biz/internal/constant"
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/repo"
	U "personal-page-be/biz/internal/utils"
	"strconv"
	"time"
)

type FileService struct {
	Repo   repo.IRepository
	Logger *logrus.Logger
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
	count, err := strconv.Atoi(string(c.FormValue("count")))
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{"code": 2001, "msg": "count序列化失败"})
		return
	}
	if count == 0 {
		count = -1
	}

	fileKey := string(c.FormValue("file_key"))
	if !U.IsAlphanumeric(fileKey) {
		c.JSON(consts.StatusOK, utils.H{"code": 2001, "msg": "文件key只能包含大小写字母和数字"})
		return
	}
	if fileKey == "" {
		fileKey = GenerateFileKey()
	}
	// 是否已经有文件

	findFileEntity, err := s.Repo.FindFileByFileKey(fileKey)
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
	u4 := uuid.New()
	currentTime := time.Now()
	timeString := currentTime.Format("2006-01-02_15_04_05")

	fileEntity := domain.FileEntity{
		User:     *user,
		UserID:   int(user.ID),
		FileKey:  fileKey,
		FileName: file.Filename,
		Count:    count,
		SaveName: fmt.Sprintf("%s_%s_%s", timeString, fileKey, u4.String()),
	}

	if err != nil {
		c.JSON(200, utils.H{"code": 2001, "msg": err.Error()})
		return
	}

	//保存文件
	err = c.SaveUploadedFile(file, fmt.Sprintf("./%s/%s", constant.FileBasePath, fileEntity.SaveName))
	if err != nil {
		c.JSON(200, utils.H{"code": 5001, "msg": "保存文件失败:" + err.Error()})
		return
	}

	//保存文件模型
	err = s.Repo.SaveFile(&fileEntity)
	if err != nil {
		c.JSON(200, utils.H{"code": 5001, "msg": "保存文件数据失败" + err.Error()})
		return
	}
	c.JSON(200, utils.H{"code": 0, "msg": "上传成功", "data": fileEntity})
	return
}

func (s *FileService) getFileInfo(fileKey string) (int, string, *domain.FileEntity) {
	if fileKey == "" {
		return 2001, "参数不合法，未提交file-key", nil
	}

	fileEntity, err := s.Repo.FindFileByFileKey(fileKey)
	if err != nil {
		return 5001, err.Error(), nil
	}
	if fileEntity.ID == 0 {
		return 4004, "未找到相关key的文件", nil
	}

	return 0, "", fileEntity
}
func (s *FileService) FileInfo(ctx context.Context, c *app.RequestContext) {
	code, msg, data := s.getFileInfo(c.DefaultQuery("file-key", ""))
	c.JSON(consts.StatusOK, utils.H{
		"code": code,
		"msg":  msg,
		"data": data,
	})

}

func (s *FileService) DownloadFile(ctx context.Context, c *app.RequestContext) {
	code, msg, data := s.getFileInfo(c.DefaultQuery("file-key", ""))
	if code == 0 {
		filePath := fmt.Sprintf("./%s/%s", constant.FileBasePath, data.SaveName)
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
			err := s.Repo.SaveFile(data)
			if err != nil {
				s.Logger.Error("下载文件后，更新文件信息失败:" + err.Error())
			}
			return
		}

	} else {
		c.JSON(consts.StatusOK, utils.H{
			"code": code,
			"msg":  msg,
		})
	}
}

func (s *FileService) DeleteFile(ctx context.Context, c *app.RequestContext) {
	fileID := c.Param("id")
	fileIDInt, err := strconv.ParseUint(fileID, 10, 64)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 2001,
			"msg":  "参数错误",
		})
		return
	}

	file, err := s.Repo.FindFileByID(uint(fileIDInt))
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
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

	if uint(file.UserID) != user.ID && user.Role != "admin" {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4003,
			"msg":  "无权执行操作",
		})
		return
	}

	err = s.Repo.RemoveFile(file.ID)
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

}

func (s *FileService) getUserId(ctx context.Context) (uid int) {
	username := ctx.Value("username")
	user, err := s.Repo.FindUser(username.(string))
	if err != nil {

		return 0
	}
	return int(user.ID)
}
