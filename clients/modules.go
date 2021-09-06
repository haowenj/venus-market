package clients

import (
	"context"
	"github.com/filecoin-project/venus-auth/log"
	"github.com/filecoin-project/venus-market/builder"
	types2 "github.com/filecoin-project/venus-messager/types"
	"github.com/filecoin-project/venus/app/client"
	"github.com/filecoin-project/venus/app/client/apiface"
	"github.com/filecoin-project/venus/pkg/types"
	"golang.org/x/xerrors"
	"time"
)

const (
	ReplaceMpoolMethod builder.Invoke = 6
)

func ConvertMpoolToMessager(fullNode apiface.FullNode, messager IMessager) error {
	fullNodeStruct := fullNode.(*client.FullNodeStruct)
	fullNodeStruct.IMessagePoolStruct.Internal.MpoolPushMessage = func(ctx context.Context, p1 *types.UnsignedMessage, p2 *types.MessageSendSpec) (*types.SignedMessage, error) {
		uid, err := messager.PushMessage(ctx, p1, nil)
		if err != nil {
			return nil, err
		}
		for {
			msgDetail, err := messager.GetMessageByUid(ctx, uid)
			if err != nil {
				log.Errorf("get message detail from messager %w", err)
				return nil, err
			}
			switch msgDetail.State {
			case types2.UnFillMsg:
				time.Sleep(time.Second * 10)
				continue
			case types2.FailedMsg:
				return nil, xerrors.Errorf("push message %w", err)
			default:
				return &types.SignedMessage{
					Message:   msgDetail.UnsignedMessage,
					Signature: *msgDetail.Signature,
				}, nil
			}
		}
	}
	return nil
}

var ClientsOpts = builder.Options(
	builder.Override(ReplaceMpoolMethod, ConvertMpoolToMessager),
)
