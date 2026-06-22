package message

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/google/uuid"
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/response"
	U "personal-page-be/biz/internal/utils"
)

type feedbackReq struct {
	CategoryID int    `json:"categoryId"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Name       string `json:"name"`
	Contact    string `json:"contact"`
}

type feedbackReplyReq struct {
	FeedbackID uint   `json:"feedbackId"`
	MessageID  uint   `json:"message_id"`
	Content    string `json:"content"`
}

func (s *MessageService) FindAllFeedbackCategories(ctx context.Context, c *app.RequestContext) {
	categories, err := s.Repo.FindAllFeedbackCategory()
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	data := make([]utils.H, 0, len(*categories))
	for _, category := range *categories {
		data = append(data, utils.H{
			"id":   category.ID,
			"name": category.Name,
		})
	}
	response.OK(c, data, "ok")
}

func (s *MessageService) SaveFeedback(ctx context.Context, c *app.RequestContext) {
	var req feedbackReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	entity := domain.FeedbackEntity{
		CategoryID: req.CategoryID,
		Title:      req.Title,
		Content:    req.Content,
		Name:       req.Name,
		Contact:    req.Contact,
		HaveRead:   false,
		UUID:       uuid.New().String(),
	}
	if err := s.Repo.SaveFeedback(&entity); err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if key := s.Config.EffectiveNotifyKey(); key != "" {
		msg := fmt.Sprintf("反馈 id: %d  \n类别: %s  \n署名: %s  \n联系方式: %s  \n内容: %s", entity.ID, entity.Category.Name, entity.Name, entity.Contact, entity.Content)
		U.SendServerMessage(key, "新反馈提醒", msg, s.Logger)
	}
	response.OK(c, utils.H{"uuid": entity.UUID}, "ok")
}

func (s *MessageService) GetFeedback(ctx context.Context, c *app.RequestContext) {
	queryUUID := c.DefaultQuery("uuid", "")
	if queryUUID == "" {
		response.Error(c, 2001, "缺少反馈 UUID")
		return
	}
	entity, err := s.Repo.FindFeedbackByUUID(queryUUID)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if entity.ID == 0 {
		response.Error(c, 4004, "未找到对应反馈")
		return
	}
	response.OK(c, feedbackDTO(entity), "ok")
}

func (s *MessageService) GetFeedbackReply(ctx context.Context, c *app.RequestContext) {
	queryUUID := c.DefaultQuery("feedbackUuid", "")
	if queryUUID == "" {
		queryUUID = c.DefaultQuery("uuid", "")
	}
	if queryUUID == "" {
		response.Error(c, 2001, "缺少反馈 UUID")
		return
	}
	entity, err := s.Repo.FindFeedbackByUUID(queryUUID)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if entity.ID == 0 {
		response.Error(c, 4004, "未找到对应反馈")
		return
	}
	reply, err := s.Repo.FindReplyByFeedbackID(entity.ID)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if reply.ID == 0 {
		response.Error(c, 4004, "该反馈暂无回复")
		return
	}
	reply.HaveRead = true
	_ = s.Repo.SaveFeedbackReply(reply)
	response.OK(c, replyDTO(reply), "ok")
}

func (s *MessageService) AddFeedbackReply(ctx context.Context, c *app.RequestContext) {
	if !s.ensureAdmin(ctx, c) {
		return
	}
	var req feedbackReplyReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, 5001, "操作失败: "+err.Error())
		return
	}
	if req.FeedbackID == 0 {
		req.FeedbackID = req.MessageID
	}
	if req.FeedbackID == 0 {
		response.Error(c, 2001, "缺少反馈 ID")
		return
	}
	entity, err := s.Repo.FindFeedbackByID(req.FeedbackID)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if entity.ID == 0 {
		response.Error(c, 4004, "未找到对应反馈")
		return
	}
	entity.HaveRead = true
	if err = s.Repo.SaveFeedback(entity); err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	reply := domain.FeedbackReplyEntity{
		Content:   req.Content,
		MessageID: entity.ID,
		HaveRead:  false,
	}
	if err = s.Repo.SaveFeedbackReply(&reply); err != nil {
		response.Error(c, 5001, "操作失败: "+err.Error())
		return
	}
	response.OK(c, nil, "ok")
}

func (s *MessageService) GetUnreadFeedback(ctx context.Context, c *app.RequestContext) {
	if !s.ensureAdmin(ctx, c) {
		return
	}
	entities, err := s.Repo.GetUnreadFeedback()
	if err != nil {
		response.Error(c, 5001, "操作失败: "+err.Error())
		return
	}
	data := make([]utils.H, 0, len(*entities))
	for _, entity := range *entities {
		data = append(data, feedbackDTO(&entity))
	}
	response.OK(c, data, "ok")
}

func (s *MessageService) ensureAdmin(ctx context.Context, c *app.RequestContext) bool {
	username, _ := ctx.Value("username").(string)
	user, err := s.Repo.FindUser(username)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return false
	}
	if !domain.IsAdminRole(user.Role) {
		response.Error(c, 4003, "无权执行此操作")
		return false
	}
	return true
}

func feedbackDTO(entity *domain.FeedbackEntity) utils.H {
	return utils.H{
		"id":         entity.ID,
		"uuid":       entity.UUID,
		"categoryId": entity.CategoryID,
		"category":   entity.Category.Name,
		"title":      entity.Title,
		"content":    entity.Content,
		"name":       entity.Name,
		"contact":    entity.Contact,
		"haveRead":   entity.HaveRead,
		"createdAt":  entity.CreatedAt.Unix(),
	}
}

func replyDTO(entity *domain.FeedbackReplyEntity) utils.H {
	return utils.H{
		"id":         entity.ID,
		"feedbackId": entity.MessageID,
		"messageId":  entity.MessageID,
		"content":    entity.Content,
		"haveRead":   entity.HaveRead,
		"createdAt":  entity.CreatedAt.Unix(),
	}
}
