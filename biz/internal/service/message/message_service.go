package message

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"personal-page-be/biz/internal/assembler"
	"personal-page-be/biz/internal/domain"
	U "personal-page-be/biz/internal/utils"
)

func (s *MessageService) FindAllMessageCategories(ctx context.Context, c *app.RequestContext) {
	categories, err := s.Repo.FindAllMessageCategory()
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
	} else {
		c.JSON(consts.StatusOK, utils.H{
			"code": 0,
			"data": assembler.MessageCategoriesEntityToDTO(categories),
		})
	}
}

func (s *MessageService) SaveMessage(ctx context.Context, c *app.RequestContext) {
	var entity domain.MessageEntity
	err := c.BindAndValidate(&entity)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
		return
	}
	entity.HaveRead = false
	err = s.Repo.SaveMessage(&entity)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
		return
	}
	U.SendServerMessage(s.Config.HttpConfig.SecretKey, "新消息提醒", fmt.Sprintf("消息id：%d  \n类别：%s  \n署名：%s  \n联系方式：%s  \n内容：%s", entity.ID, entity.Category.Name, entity.Name, entity.Contact, entity.Message))
	c.JSON(consts.StatusOK, utils.H{
		"code": 0,
	})
}
