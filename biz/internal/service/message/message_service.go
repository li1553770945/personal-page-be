package message

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
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
	u4 := uuid.New()
	entity.UUID = u4.String()
	err = s.Repo.SaveMessage(&entity)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
		return
	}
	U.SendServerMessage(s.Config.HttpConfig.SecretKey, "新留言提醒", fmt.Sprintf("消息id：%d  \n类别：%s  \n署名：%s  \n联系方式：%s  \n内容：%s", entity.ID, entity.Category.Name, entity.Name, entity.Contact, entity.Message))
	c.JSON(consts.StatusOK, utils.H{
		"code": 0,
		"data": entity,
	})
}

func (s *MessageService) GetReply(ctx context.Context, c *app.RequestContext) {
	queryUUID := c.DefaultQuery("uuid", "")
	if queryUUID == "" {
		c.JSON(consts.StatusOK, utils.H{
			"code": 2001,
			"msg":  "参数错误",
		})
		return
	}
	msg, err := s.Repo.FindMessageByUUID(queryUUID)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
		return
	}
	if msg.ID == 0 {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4004,
			"msg":  "未找到相关文章",
		})
		return
	}
	reply, err := s.Repo.FindReplyByMessageID(msg.ID)
	if reply.ID == 0 {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4004,
			"msg":  "该文章暂未回复",
		})
		return
	}
	c.JSON(consts.StatusOK, utils.H{
		"code": 0,
		"data": reply,
	})
}
func (s *MessageService) AddReply(ctx context.Context, c *app.RequestContext) {
	username := ctx.Value("username")
	user, err := s.Repo.FindUser(username.(string))
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
		return
	}
	if user.Role != "admin" {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4003,
			"msg":  "您无权执行此操作",
		})
		return
	}

	var reply domain.ReplyEntity
	err = c.BindAndValidate(&reply)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  "操作失败：" + err.Error(),
		})
		return
	}
	msg, err := s.Repo.FindMessageByID(reply.MessageID)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
		return
	}
	if msg.ID == 0 {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4004,
			"msg":  "未找到相关文章",
		})
		return
	}
	msg.HaveRead = true
	err = s.Repo.SaveMessage(msg)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
		return
	}

	err = s.Repo.SaveReply(&reply)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  "操作失败：" + err.Error(),
		})
		return
	}
	c.JSON(consts.StatusOK, utils.H{
		"code": "0",
	})
	return
}
