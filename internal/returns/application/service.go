package application

import (
	"context"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/returns/domain"
	"github.com/wyfcoding/pkg/messagequeue"
	"github.com/wyfcoding/pkg/xerrors"
)

// 生成摘要：售后退货应用逻辑。

type ReturnService struct {
	repo domain.Repository
	mq   messagequeue.EventPublisher
}

func NewReturnService(repo domain.Repository, mq messagequeue.EventPublisher) *ReturnService {
	return &ReturnService{repo: repo, mq: mq}
}

func (s *ReturnService) ApproveReturn(ctx context.Context, returnID uint) error {
	req, err := s.repo.FindByID(returnID)
	if err != nil {
		return xerrors.NotFound("return request not found")
	}

	if err := req.Approve(); err != nil {
		return xerrors.InvalidArg(err.Error())
	}

	if err := s.repo.Save(req); err != nil {
		return err
	}

	// 异步发送物流预约事件
	_ = s.mq.Publish(ctx, "logistic.return_booking", strconv.FormatUint(uint64(req.ID), 10), req)
	return nil
}
