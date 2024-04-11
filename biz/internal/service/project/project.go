package project

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"personal-page-be/biz/internal/domain"
	"strconv"
)

func (s *ProjectService) AddProject(ctx context.Context, c *app.RequestContext) {
	var project domain.ProjectEntity
	err := c.BindAndValidate(&project)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4001,
			"msg":  err.Error(),
		})
		return
	}
	err = s.Repo.SaveProject(&project)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
		return
	}
	c.JSON(consts.StatusOK, utils.H{
		"code": 0,
		"msg":  "添加成功",
	})
}

func (s *ProjectService) RemoveProject(ctx context.Context, c *app.RequestContext) {
	projectID := c.Param("id")
	projectIDInt, err := strconv.ParseUint(projectID, 10, 64)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 2001,
			"msg":  "参数错误",
		})
		return
	}

	project, err := s.Repo.GetProject(uint(projectIDInt))
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
	}

	err = s.Repo.RemoveProject(project.ID)
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

func (s *ProjectService) GetPages(ctx context.Context, c *app.RequestContext) {

	c.JSON(consts.StatusOK, utils.H{
		"code": 0,
		"msg":  "获取成功",
	})
}

func (s *ProjectService) GetProjects(ctx context.Context, c *app.RequestContext) {

	c.JSON(consts.StatusOK, utils.H{
		"code": 0,
		"msg":  "获取成功",
	})
}
