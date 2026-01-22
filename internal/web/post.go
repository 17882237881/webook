package web

import (
	"net/http"
	"strconv"
	"webook/internal/domain"
	"webook/internal/service"
	"webook/internal/web/ginx"

	"github.com/gin-gonic/gin"
)

// PostHandler 帖子相关的 HTTP 请求处理
type PostHandler struct {
	svc service.PostService
}

// NewPostHandler 创建 PostHandler 实例
func NewPostHandler(svc service.PostService) *PostHandler {
	return &PostHandler{svc: svc}
}

// RegisterRoutes 注册路由
func (h *PostHandler) RegisterRoutes(server *gin.Engine) {
	pg := server.Group("/posts")
	{
		// 作者操作（需要登录）
		pg.POST("", h.Save)               // 创建/更新草稿
		pg.POST("/publish", h.Publish)    // 发布帖子
		pg.GET("/author", h.ListByAuthor) // 获取作者的帖子列表
		pg.GET("/draft/:id", h.GetDraft)  // 获取草稿详情

		// 读者操作
		pg.GET("/:id", h.GetPublished) // 获取已发布帖子
		pg.GET("", h.ListPublished)    // 获取已发布帖子列表

		// 删除
		pg.DELETE("/:id", h.Delete)
	}
}

// Save 保存帖子（创建或更新草稿）
// POST /posts
func (h *PostHandler) Save(c *gin.Context) {
	type SaveReq struct {
		Id      int64  `json:"id"` // 0 表示新建
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	var req SaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "参数错误")
		return
	}

	// 从 Context 获取当前登录用户
	authorId := c.GetInt64("userId")
	if authorId == 0 {
		ginx.ErrorWithStatus(c, http.StatusUnauthorized, ginx.CodeUnauthorized, "请先登录")
		return
	}

	id, err := h.svc.Save(c.Request.Context(), domain.Post{
		Id:       req.Id,
		Title:    req.Title,
		Content:  req.Content,
		AuthorId: authorId,
	})
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "保存失败")
		return
	}

	ginx.Success(c, gin.H{"id": id})
}

// Publish 发布帖子
// POST /posts/publish
func (h *PostHandler) Publish(c *gin.Context) {
	type PublishReq struct {
		Id      int64  `json:"id"` // 0 表示新建并发布
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	var req PublishReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "参数错误")
		return
	}

	authorId := c.GetInt64("userId")
	if authorId == 0 {
		ginx.ErrorWithStatus(c, http.StatusUnauthorized, ginx.CodeUnauthorized, "请先登录")
		return
	}

	id, err := h.svc.Publish(c.Request.Context(), domain.Post{
		Id:       req.Id,
		Title:    req.Title,
		Content:  req.Content,
		AuthorId: authorId,
	})
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "发布失败")
		return
	}

	ginx.Success(c, gin.H{"id": id})
}

// GetDraft 获取草稿详情（作者用）
// GET /posts/draft/:id
func (h *PostHandler) GetDraft(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "无效的帖子ID")
		return
	}

	post, err := h.svc.GetById(c.Request.Context(), id)
	if err != nil {
		ginx.Error(c, ginx.CodeNotFound, "帖子不存在")
		return
	}

	// 验证是否是作者本人
	authorId := c.GetInt64("userId")
	if post.AuthorId != authorId {
		ginx.ErrorWithStatus(c, http.StatusForbidden, ginx.CodeForbidden, "无权访问")
		return
	}

	ginx.Success(c, gin.H{
		"id":      post.Id,
		"title":   post.Title,
		"content": post.Content,
		"status":  post.Status,
		"ctime":   post.Ctime,
		"utime":   post.Utime,
	})
}

// GetPublished 获取已发布帖子详情（读者用）
// GET /posts/:id
func (h *PostHandler) GetPublished(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "无效的帖子ID")
		return
	}

	post, err := h.svc.GetPublishedById(c.Request.Context(), id)
	if err != nil {
		ginx.Error(c, ginx.CodeNotFound, "帖子不存在")
		return
	}

	ginx.Success(c, gin.H{
		"id":       post.Id,
		"title":    post.Title,
		"content":  post.Content,
		"authorId": post.AuthorId,
		"ctime":    post.Ctime,
		"utime":    post.Utime,
	})
}

// ListByAuthor 获取作者的帖子列表
// GET /posts/author?page=1&pageSize=10
func (h *PostHandler) ListByAuthor(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	authorId := c.GetInt64("userId")
	if authorId == 0 {
		ginx.ErrorWithStatus(c, http.StatusUnauthorized, ginx.CodeUnauthorized, "请先登录")
		return
	}

	posts, err := h.svc.ListByAuthor(c.Request.Context(), authorId, page, pageSize)
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "获取列表失败")
		return
	}

	ginx.Success(c, gin.H{
		"posts": h.toPostVOs(posts),
	})
}

// ListPublished 获取已发布帖子列表（公开）
// GET /posts?page=1&pageSize=10
func (h *PostHandler) ListPublished(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	posts, err := h.svc.ListPublished(c.Request.Context(), page, pageSize)
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "获取列表失败")
		return
	}

	ginx.Success(c, gin.H{
		"posts": h.toPostVOs(posts),
	})
}

// Delete 删除帖子
// DELETE /posts/:id
func (h *PostHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "无效的帖子ID")
		return
	}

	authorId := c.GetInt64("userId")
	if authorId == 0 {
		ginx.ErrorWithStatus(c, http.StatusUnauthorized, ginx.CodeUnauthorized, "请先登录")
		return
	}

	err = h.svc.Delete(c.Request.Context(), id, authorId)
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "删除失败")
		return
	}

	ginx.SuccessMsg(c, "删除成功")
}

// toPostVOs 转换为视图对象列表
func (h *PostHandler) toPostVOs(posts []domain.Post) []gin.H {
	result := make([]gin.H, len(posts))
	for i, p := range posts {
		result[i] = gin.H{
			"id":       p.Id,
			"title":    p.Title,
			"content":  p.Content,
			"authorId": p.AuthorId,
			"status":   p.Status,
			"ctime":    p.Ctime,
			"utime":    p.Utime,
		}
	}
	return result
}
