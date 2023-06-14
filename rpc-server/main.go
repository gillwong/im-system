package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	rpc "github.com/gillwong/im-system/rpc-server/kitex_gen/rpc/imservice"
	etcd "github.com/kitex-contrib/registry-etcd"
	"github.com/redis/go-redis/v9"
)

type RedisCli struct {
	cli *redis.Client
}

type MessageJson struct {
	Sender    string `json:"sender"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

var redisDb = &RedisCli{}

func main() {
	ctx := context.Background()

	err := redisDb.InitCli(ctx, "redis:6379", "")
	if err != nil {
		log.Fatal(err)
	}

	r, err := etcd.NewEtcdRegistry([]string{"etcd:2379"}) // r should not be reused.
	if err != nil {
		log.Fatal(err)
	}

	svr := rpc.NewServer(new(IMServiceImpl), server.WithRegistry(r), server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
		ServiceName: "demo.rpc.server",
	}))

	err = svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}

func (c *RedisCli) InitCli(ctx context.Context, address, pw string) error {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: pw,
		DB:       0,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return err
	}

	c.cli = client
	return nil
}

func (c *RedisCli) getMsg(ctx context.Context, id string, start, end int64, reverse bool) ([]*MessageJson, error) {
	var (
		rawMsg   []string
		messages []*MessageJson
		err      error
	)

	if reverse {
		rawMsg, err = c.cli.ZRevRange(ctx, id, start, end).Result()
		if err != nil {
			return nil, err
		}
	} else {
		rawMsg, err = c.cli.ZRange(ctx, id, start, end).Result()
		if err != nil {
			return nil, err
		}
	}

	for _, msg := range rawMsg {
		temp := &MessageJson{}
		err := json.Unmarshal([]byte(msg), temp)
		if err != nil {
			panic(err)
		}
		messages = append(messages, temp)
	}

	return messages, nil
}

func (c *RedisCli) SaveMsg(ctx context.Context, id string, msg *MessageJson) error {
	text, err1 := json.Marshal(msg)
	if err1 != nil {
		return err1
	}

	member := &redis.Z{
		Score:  float64(msg.Timestamp),
		Member: text,
	}

	_, err2 := c.cli.ZAdd(ctx, id, *member).Result()
	if err2 != nil {
		return err2
	}

	return nil
}
