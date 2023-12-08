package global_service

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"os"
	"personal-page-be/biz/internal/constant"
)

func (s *GlobalService) DeleteFile() {
	root := constant.FileBasePath
	files, err := os.ReadDir(root)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileEntity, err := s.Repo.FindFileBySaveName(file.Name())
		if err != nil {
			panic(err)
		}
		if fileEntity.ID == 0 || fileEntity.Count == 0 {
			filePath := fmt.Sprintf("./%s/%s", constant.FileBasePath, file.Name())
			err = os.Remove(filePath)
			if err != nil {
				s.Logger.Error("删除文件失败" + err.Error())
			} else {
				s.Logger.Info(fmt.Sprintf("删除文件%s成功", file.Name()))
			}
		}
	}
}
func (s *GlobalService) StartCronDeleteFile() {
	crontab := cron.New(cron.WithSeconds()) //精确到秒
	task := s.DeleteFile
	spec := "0 5 5 1/1 * ? " //每天5:05
	// 添加定时任务,
	_, err := crontab.AddFunc(spec, task)
	if err != nil {
		panic(err)
	}
	crontab.Start()
	s.Logger.Info("启动定时删除文件任务成功")
}
