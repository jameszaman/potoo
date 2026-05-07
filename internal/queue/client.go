package queue

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
)

type Client struct {
	c *asynq.Client
}

func NewClient(redisAddr string) *Client {
	return &Client{
		c: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}),
	}
}

func (c *Client) Enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) error {
	_, err := c.c.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return fmt.Errorf("enqueue: %w", err)
	}
	return nil
}

func (c *Client) Close() error {
	return c.c.Close()
}
