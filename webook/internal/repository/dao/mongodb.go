package dao

import "context"

type MongoDBArticleDAO struct {
}

func (m MongoDBArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (m MongoDBArticleDAO) UpdateById(ctx context.Context, art Article) error {
	//TODO implement me
	panic("implement me")
}

func (m MongoDBArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (m MongoDBArticleDAO) SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error {
	//TODO implement me
	panic("implement me")
}
