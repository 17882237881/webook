package domain

// PostStats holds aggregate counters for a post.
type PostStats struct {
	PostId     int64
	LikeCnt    int64
	CollectCnt int64
	ReadCnt    int64
}

// PostUserStats holds user-specific flags for a post.
type PostUserStats struct {
	Liked     bool
	Collected bool
}
