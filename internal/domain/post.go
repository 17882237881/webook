package domain

// Post 帖子领域模型
type Post struct {
	Id       int64  // 帖子ID
	Title    string // 标题
	Content  string // 正文内容
	AuthorId int64  // 作者ID
	Status   uint8  // 状态：0-未发布，1-已发布，2-仅自己可见
	Ctime    int64  // 创建时间（毫秒时间戳）
	Utime    int64  // 更新时间（毫秒时间戳）
}

// 帖子状态常量
const (
	PostStatusUnpublished uint8 = iota // 未发布（草稿）
	PostStatusPublished                // 已发布
	PostStatusPrivate                  // 仅自己可见
)
