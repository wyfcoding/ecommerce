package elasticsearch

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"github.com/wyfcoding/pkg/search"
)

type userSearchRepository struct {
	client *search.Client
	index  string
}

// NewUserSearchRepository 创建用户搜索仓储
func NewUserSearchRepository(client *search.Client, index string) domain.UserSearchRepository {
	if client == nil {
		return nil
	}
	if index == "" {
		index = "users"
	}
	return &userSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *userSearchRepository) Index(ctx context.Context, user *domain.User) error {
	if user == nil {
		return nil
	}
	docID := fmt.Sprintf("%d", user.ID)
	return r.client.Index(ctx, r.index, docID, user)
}

func (r *userSearchRepository) Search(ctx context.Context, keyword string, limit, offset int) ([]*domain.User, int64, error) {
	query := map[string]any{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
	}

	if keyword == "" {
		query["query"] = map[string]any{"match_all": map[string]any{}}
	} else {
		query["query"] = map[string]any{
			"multi_match": map[string]any{
				"query":  keyword,
				"fields": []string{"username", "email", "phone", "nickname", "full_name"},
			},
		}
	}

	var searchRes struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.User `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, err
	}

	users := make([]*domain.User, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		user := hit.Source
		users[i] = &user
	}
	return users, searchRes.Hits.Total.Value, nil
}
