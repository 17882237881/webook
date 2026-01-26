package web

import (
	"net/http"
	"strconv"
	"webook/internal/adapters/inbound/http/ginx"
	"webook/internal/domain"
	service "webook/internal/ports/input"

	"github.com/gin-gonic/gin"
)

// PostHandler 帖子相关的 HTTP 请求处理
type PostHandler struct {
	svc      service.PostService
	statsSvc service.PostInteractionService
}

// NewPostHandler 创建 PostHandler 实例
func NewPostHandler(svc service.PostService, statsSvc service.PostInteractionService) *PostHandler {
	return &PostHandler{
		svc:      svc,
		statsSvc: statsSvc,
	}
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

		// Reader actions (published posts)
		pg.GET("/:id", h.GetPublished) // Get published post
		pg.GET("", h.ListPublished)    // List published posts
		pg.POST("/:id/like", h.Like)
		pg.POST("/:id/unlike", h.Unlike)
		pg.POST("/:id/collect", h.Collect)
		pg.POST("/:id/uncollect", h.Uncollect)
		pg.POST("/:id/read", h.Read)

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
		Id      int64  `json:"id"`
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
		ginx.Error(c, ginx.CodeNotFound, "post not found")
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
		ginx.Error(c, ginx.CodeNotFound, "post not found")
		return
	}

	userId := c.GetInt64("userId")
	_ = h.statsSvc.Read(c.Request.Context(), id, userId, c.ClientIP(), c.Request.UserAgent())
	stats, userStats, _ := h.statsSvc.GetStats(c.Request.Context(), id, userId)
	ginx.Success(c, gin.H{
		"id":         post.Id,
		"title":      post.Title,
		"content":    post.Content,
		"authorId":   post.AuthorId,
		"ctime":      post.Ctime,
		"utime":      post.Utime,
		"likeCnt":    stats.LikeCnt,
		"collectCnt": stats.CollectCnt,
		"readCnt":    stats.ReadCnt,
		"liked":      userStats.Liked,
		"collected":  userStats.Collected,
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

	posts, total, err := h.svc.ListByAuthor(c.Request.Context(), authorId, page, pageSize)
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "获取列表失败")
		return
	}

	postIds := make([]int64, 0, len(posts))
	for _, p := range posts {
		postIds = append(postIds, p.Id)
	}
	statsMap, userStats, _ := h.statsSvc.GetStatsBatch(c.Request.Context(), postIds, authorId)
	ginx.Success(c, gin.H{
		"posts":    h.toPostVOs(posts, statsMap, userStats),
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
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

	posts, total, err := h.svc.ListPublished(c.Request.Context(), page, pageSize)
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "获取列表失败")
		return
	}

	postIds := make([]int64, 0, len(posts))
	for _, p := range posts {
		postIds = append(postIds, p.Id)
	}
	userId := c.GetInt64("userId")
	statsMap, userStats, _ := h.statsSvc.GetStatsBatch(c.Request.Context(), postIds, userId)
	ginx.Success(c, gin.H{
		"posts":    h.toPostVOs(posts, statsMap, userStats),
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
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
func (h *PostHandler) toPostVOs(posts []domain.Post, stats map[int64]domain.PostStats, userStats map[int64]domain.PostUserStats) []gin.H {
	result := make([]gin.H, len(posts))
	for i, p := range posts {
		st := stats[p.Id]
		us := userStats[p.Id]
		result[i] = gin.H{
			"id":         p.Id,
			"title":      p.Title,
			"content":    p.Content,
			"authorId":   p.AuthorId,
			"status":     p.Status,
			"ctime":      p.Ctime,
			"utime":      p.Utime,
			"likeCnt":    st.LikeCnt,
			"collectCnt": st.CollectCnt,
			"readCnt":    st.ReadCnt,
			"liked":      us.Liked,
			"collected":  us.Collected,
		}
	}
	return result
}

func (h *PostHandler) Like(c *gin.Context) {
	id, ok := h.getPostIdParam(c)
	if !ok {
		return
	}
	userId := c.GetInt64("userId")
	if userId == 0 {
		ginx.ErrorWithStatus(c, http.StatusUnauthorized, ginx.CodeUnauthorized, "请先登录")
		return
	}
	if err := h.statsSvc.Like(c.Request.Context(), id, userId); err != nil {
		ginx.Error(c, ginx.CodeInternalError, "点赞失败")
		return
	}
	ginx.SuccessMsg(c, "ok")
}

func (h *PostHandler) Unlike(c *gin.Context) {
	id, ok := h.getPostIdParam(c)
	if !ok {
		return
	}
	userId := c.GetInt64("userId")
	if userId == 0 {
		ginx.ErrorWithStatus(c, http.StatusUnauthorized, ginx.CodeUnauthorized, "请先登录")
		return
	}
	if err := h.statsSvc.Unlike(c.Request.Context(), id, userId); err != nil {
		ginx.Error(c, ginx.CodeInternalError, "取消点赞失败")
		return
	}
	ginx.SuccessMsg(c, "ok")
}

func (h *PostHandler) Collect(c *gin.Context) {
	id, ok := h.getPostIdParam(c)
	if !ok {
		return
	}
	userId := c.GetInt64("userId")
	if userId == 0 {
		ginx.ErrorWithStatus(c, http.StatusUnauthorized, ginx.CodeUnauthorized, "请先登录")
		return
	}
	if err := h.statsSvc.Collect(c.Request.Context(), id, userId); err != nil {
		ginx.Error(c, ginx.CodeInternalError, "收藏失败")
		return
	}
	ginx.SuccessMsg(c, "ok")
}

func (h *PostHandler) Uncollect(c *gin.Context) {
	id, ok := h.getPostIdParam(c)
	if !ok {
		return
	}
	userId := c.GetInt64("userId")
	if userId == 0 {
		ginx.ErrorWithStatus(c, http.StatusUnauthorized, ginx.CodeUnauthorized, "请先登录")
		return
	}
	if err := h.statsSvc.Uncollect(c.Request.Context(), id, userId); err != nil {
		ginx.Error(c, ginx.CodeInternalError, "取消收藏失败")
		return
	}
	ginx.SuccessMsg(c, "ok")
}

func (h *PostHandler) Read(c *gin.Context) {
	id, ok := h.getPostIdParam(c)
	if !ok {
		return
	}
	userId := c.GetInt64("userId")
	if err := h.statsSvc.Read(c.Request.Context(), id, userId, c.ClientIP(), c.Request.UserAgent()); err != nil {
		ginx.Error(c, ginx.CodeInternalError, "记录阅读失败")
		return
	}
	ginx.SuccessMsg(c, "ok")
}

func (h *PostHandler) getPostIdParam(c *gin.Context) (int64, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "无效的帖子ID")
		return 0, false
	}
	return id, true
}
