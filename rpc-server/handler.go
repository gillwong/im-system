package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gillwong/im-system/rpc-server/kitex_gen/rpc"
)

// IMServiceImpl implements the last service interface defined in the IDL.
type IMServiceImpl struct{}

func (s *IMServiceImpl) Send(ctx context.Context, req *rpc.SendRequest) (*rpc.SendResponse, error) {
	if err := validateSendReq(req); err != nil {
		return nil, err
	}

	timestamp := time.Now().Unix()
	msg := &MessageJson{
		Message:   req.Message.GetText(),
		Sender:    req.Message.GetSender(),
		Timestamp: timestamp,
	}

	id, err := getId(req.Message.GetChat())
	if err != nil {
		return nil, err
	}

	if err := redisDb.SaveMsg(ctx, id, msg); err != nil {
		return nil, err
	}

	resp := rpc.NewSendResponse()

	resp.Code = 0
	resp.Msg = "successfully created message"

	return resp, nil
}

func (s *IMServiceImpl) Pull(ctx context.Context, req *rpc.PullRequest) (*rpc.PullResponse, error) {
	id, err := getId(req.GetChat())
	if err != nil {
		return nil, err
	}

	start := req.GetCursor()
	end := start + int64(req.GetLimit())

	messages, err := redisDb.getMsg(ctx, id, start, end, req.GetReverse())
	if err != nil {
		return nil, err
	}

	respMsg := make([]*rpc.Message, 0)
	var counter int32 = 0
	var nextCursor int64 = 0
	hasMore := false

	for _, msg := range messages {
		if counter+1 > req.GetLimit() {
			hasMore = true
			nextCursor = end
			break
		}
		temp := &rpc.Message{
			Chat:   req.GetChat(),
			Text:   msg.Message,
			Sender: msg.Sender,
		}
		respMsg = append(respMsg, temp)
		counter += 1
	}

	resp := rpc.NewPullResponse()
	resp.Messages = respMsg
	resp.Code = 0
	resp.Msg = "pull success"
	resp.HasMore = &hasMore
	resp.NextCursor = &nextCursor

	return resp, nil
}

func getId(chat string) (string, error) {
	var id string

	senders := strings.Split(strings.ToLower(chat), ":")
	if len(senders) != 2 {
		return "", fmt.Errorf("invalid chat:", chat, "Must follow format user1:user2")
	}

	sender1, sender2 := senders[0], senders[1]
	if strings.Compare(sender1, sender2) == 1 {
		id = fmt.Sprintf("%s:%s", sender2, sender1)
	} else {
		id = fmt.Sprintf("%s:%s", sender1, sender2)
	}

	return id, nil

}

func validateSendReq(req *rpc.SendRequest) error {
	senders := strings.Split(req.Message.Chat, ":")
	if len(senders) != 2 {
		return fmt.Errorf("invalid chat:", req.Message.GetChat(), "Must follow format user1:user2")
	}

	sender1, sender2 := senders[0], senders[1]
	sender := req.Message.GetSender()
	if sender != sender1 && sender != sender2 {
		return fmt.Errorf("sender %s not in chat", sender)
	}

	return nil
}
