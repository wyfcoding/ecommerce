// 变更说明：
// 1. 【图谱模型】使用 GraphDB 模式建立经典的“关注(Following)/粉丝(Follower)”体系。
// 2. 【Feed 流】实现社交分享 Feed，支持“推拉结合(Push-Pull) Inbox”模型。
// 3. 【圈层分享】加入基于图关系的内容触达。
package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrAlreadyFollowing = errors.New("already following the user")
	ErrNotFollowing     = errors.New("not following the user")
	ErrCannotFollowSelf = errors.New("cannot follow yourself")
)

type GraphRelationType string

const (
	RelFollow GraphRelationType = "FOLLOWS"
	RelBlock  GraphRelationType = "BLOCKS"
	RelMute   GraphRelationType = "MUTES"
)

// UserNode 表示社交网络中的一个用户节点
type UserNode struct {
	UserID         uint64    `json:"user_id"`
	FollowersCount int64     `json:"followers_count"`
	FollowingCount int64     `json:"following_count"`
	PostsCount     int64     `json:"posts_count"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// GraphRelationship 社交网络中的单向边关系
type GraphRelationship struct {
	SourceID  uint64            `json:"source_id"` // 发起关注的人
	TargetID  uint64            `json:"target_id"` // 被关注的大V/朋友
	RelType   GraphRelationType `json:"rel_type"`  // 当前关系 (Follow/Block)
	CreatedAt time.Time         `json:"created_at"`
}

// InteractivePost 一条可互动的分享帖子/动态 (例如带货动态或晒单)
type InteractivePost struct {
	PostID     uint64    `json:"post_id"`
	AuthorID   uint64    `json:"author_id"`
	Content    string    `json:"content"`
	MediaURLs  []string  `json:"media_urls"`  // 图片/视频
	ProductIDs []uint64  `json:"product_ids"` // 关联的带货商品
	IsPublic   bool      `json:"is_public"`   // 仅粉丝可见，或公开广场
	LikeCount  int64     `json:"like_count"`
	ShareCount int64     `json:"share_count"`
	CreatedAt  time.Time `json:"created_at"`
}

// PostFeedItem 用于 Inbox/Outbox 流分发的数据载体（缩略快照）
type PostFeedItem struct {
	PostID    uint64 `json:"post_id"`
	AuthorID  uint64 `json:"author_id"`
	Timestamp int64  `json:"timestamp"` // 纳秒
	IsRead    bool   `json:"is_read"`
}

// FeedService Feed 流读写分离策略（收件箱模型 Inbox Model）
type FeedService interface {
	// PublishPost 将作者发送的内容推进大 V 的发件箱 (Outbox)，并依据粉丝量决定是否立刻 Push 到粉丝收件箱 (Inbox)
	PublishPost(ctx context.Context, post *InteractivePost, followers []uint64) error

	// GetTimeline 粉丝打开 APP 时，合并 (Merge) 他们关注大V的 Outbox 进自己的 Inbox 展示
	GetTimeline(ctx context.Context, viewerID uint64, cursor int64, limit int) ([]*PostFeedItem, error)

	// LikePost 点赞互动
	LikePost(ctx context.Context, postID, userID uint64) error
}

// SocialGraphRepository 图数据库仓储接口，处理 Neo4j / 专门的关系存储。
type SocialGraphRepository interface {
	// Graph Operations
	CreateUserNode(ctx context.Context, userID uint64) error
	AddRelationship(ctx context.Context, rel *GraphRelationship) error
	RemoveRelationship(ctx context.Context, sourceID, targetID uint64, relType GraphRelationType) error
	CheckRelationship(ctx context.Context, sourceID, targetID uint64) (GraphRelationType, error)

	// Queries for Feed Fan-out
	GetFollowers(ctx context.Context, targetID uint64, limit int, offset int) ([]uint64, error)
	GetFollowing(ctx context.Context, sourceID uint64, limit int, offset int) ([]uint64, error)

	// Count
	GetFollowersCount(ctx context.Context, userID uint64) (int64, error)
	GetFollowingCount(ctx context.Context, userID uint64) (int64, error)
}

// SocialRelation 轻量关系聚合，兼容邀请绑定链路。
type SocialRelation struct {
	ID           uint64    `json:"id"`
	UserID       string    `json:"user_id"`
	ParentID     string    `json:"parent_id"`
	InvitationID string    `json:"invitation_id"`
	Level        int       `json:"level"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func NewSocialRelation(userID, parentID, invitationID string, level int) (*SocialRelation, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}
	if invitationID == "" {
		return nil, errors.New("invitation id is required")
	}
	now := time.Now()
	return &SocialRelation{
		UserID:       userID,
		ParentID:     parentID,
		InvitationID: invitationID,
		Level:        level,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// Repository 兼容现有 social application 的持久化接口。
type Repository interface {
	Save(ctx context.Context, rel *SocialRelation) error
	FindByUserID(ctx context.Context, userID string) (*SocialRelation, error)
	FindChildrenByParentID(ctx context.Context, parentID string) ([]*SocialRelation, error)
}

// SocialRepository 兼容旧版 service.go 的引用类型。
type SocialRepository = Repository
