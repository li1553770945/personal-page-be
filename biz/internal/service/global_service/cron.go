package global_service

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"os"
)

func (s *GlobalService) DeleteFile() {
	root := "./files"
	files, err := os.ReadDir(root)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fmt.Println(file.Name())
		fileEntity, err := s.Repo.FindFile(file.Name())
		if err != nil {
			panic(err)
		}
		if fileEntity.ID == 0 {
			filePath := fmt.Sprintf("./file/%s", file.Name())
			os.Remove(filePath)
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
}
