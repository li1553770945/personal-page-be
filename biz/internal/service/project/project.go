package project

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"personal-page-be/biz/internal/constant"
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/service/error_type"
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
			"code": error_type.ErrInternal.Code,
			"msg":  err.Error(),
		})
		return
	}
	c.JSON(consts.StatusOK, utils.H{
		"code": 0,
		"data": project.ID,
	})
}

func (s *ProjectService) RemoveProject(ctx context.Context, c *app.RequestContext) {
	err0, _ := s.GlobalService.CheckLogin(c, true)
	if err0 != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": err0.Code,
			"msg":  err0.Message,
		})
		return
	}

	projectID := c.Param("id")
	projectIDInt, err := strconv.ParseUint(projectID, 10, 64)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": error_type.ErrBadRequest.Code,
			"msg":  error_type.ErrBadRequest.Message,
		})
		return
	}

	project, err := s.Repo.GetProject(uint(projectIDInt))

	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
		return
	}
	if project.ID == 0 {
		c.JSON(consts.StatusOK, utils.H{
			"code": error_type.ErrNotFound.Code,
			"msg":  error_type.ErrNotFound.Message,
		})
		return
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

func (s *ProjectService) GetNum(ctx context.Context, c *app.RequestContext) {
	count, err := s.Repo.GetProjectsNum()
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
	}
	c.JSON(consts.StatusOK, utils.H{
		"code": 0,
		"data": count,
	})
}

func (s *ProjectService) GetProjects(ctx context.Context, c *app.RequestContext) {
	sort := c.DefaultQuery("sort", "default")
	sortFields := map[string]string{
		"default":             "start_date desc",
		"work_of_volume_asc":  "volume_of_work asc",
		"work_of_volume_desc": "volume_of_work desc",
		"start_date_asc":      "start_date asc",
		"start_date_desc":     "start_date desc",
		"difficulty_asc":      "difficulty asc",
		"difficulty_desc":     "difficulty desc",
	}

	order, ok := sortFields[sort]
	if !ok {
		c.JSON(consts.StatusOK, utils.H{
			"code": error_type.ErrBadRequest,
			"msg":  "排序参数错误",
		})
	}

	status := c.DefaultQuery("status", "0")
	statusInt, err := strconv.Atoi(status)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": error_type.ErrBadRequest,
			"msg":  "排序参数错误",
		})
	}

	start := c.DefaultQuery("start", "0")
	startInt, err := strconv.Atoi(start)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4001,
			"msg":  err.Error(),
		})
	}

	end := c.DefaultQuery("end", strconv.Itoa(startInt+10))
	endInt, err := strconv.Atoi(end)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4001,
			"msg":  err.Error(),
		})
	}

	if endInt-startInt > constant.MaxProjectsNum {
		endInt = startInt + constant.MaxProjectsNum + 1
	}

	projects, err := s.Repo.GetProjects(startInt, endInt, order, statusInt)

	c.JSON(consts.StatusOK, utils.H{
		"code": 0,
		"data": projects,
	})
}
