package dao

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
)

const CollectIndexName = "collect_index"

type Collect struct {
	Uid   int64
	Biz   string
	BizId int64
}

type CollectElasticDAO struct {
	client *elastic.Client
}

func (c *CollectElasticDAO) Search(ctx context.Context, uid int64, biz string) ([]int64, error) {
	query := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("uid", uid),
		elastic.NewTermQuery("biz", biz))
	resp, err := c.client.Search(CollectIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]int64, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var ele Collect
		err = json.Unmarshal(hit.Source, &ele)
		if err != nil {
			return nil, err
		}
		res = append(res, ele.BizId)
	}
	return res, nil
}

func (c *CollectElasticDAO) InputCollect(ctx context.Context, collect Collect) error {
	_, err := c.client.Index().
		Index(CollectIndexName).
		BodyJson(collect).Do(ctx)
	return err
}

func NewCollectElasticDAO(client *elastic.Client) CollectDAO {
	return &CollectElasticDAO{client: client}
}
