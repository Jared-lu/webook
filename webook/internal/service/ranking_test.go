package service

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
	"webook/webook/internal/domain"
)

func TestBatchRankingService_TopN(t *testing.T) {
	now := time.Now()
	fmt.Println(now)
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (ArticleService, InteractiveService)
		wantErr  error
		wantArts []domain.Article
	}{
		{},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			artSvc, intrSvc := tc.mock(ctrl)
			svc := NewBatchRankingService(artSvc, intrSvc)
			arts, err := svc.topN(context.Background())
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantArts, arts)
		})
	}
}
