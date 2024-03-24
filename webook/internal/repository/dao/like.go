package dao

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
)

const LikeIndexName = "like_index"

type Like struct {
	Uid    int64
	Biz    string
	BizId  int64
	Status uint8
}

type LikeElasticDAO struct {
	client *elastic.Client
}

func (l *LikeElasticDAO) Search(ctx context.Context, uid int64, biz string) ([]int64, error) {
	query := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("uid", uid),
		elastic.NewTermQuery("biz", biz))
	resp, err := l.client.Search(LikeIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]int64, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var ele Like
		err = json.Unmarshal(hit.Source, &ele)
		if err != nil {
			return nil, err
		}
		res = append(res, ele.BizId)
	}
	return res, nil
}

func (l *LikeElasticDAO) InputLike(ctx context.Context, like Like) error {
	_, err := l.client.Index().
		Index(LikeIndexName).
		BodyJson(like).Do(ctx)
	return err
}

func NewLikeElasticDAO(client *elastic.Client) LikeDAO {
	return &LikeElasticDAO{client: client}
}
