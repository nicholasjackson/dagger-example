package main

import (
	"context"
	"fmt"
	"time"

	"dagger.io/dagger"
	"github.com/hashicorp/go-hclog"
)

const DefaultTimeout = 100 * time.Second

type Build struct {
	client    *dagger.Client
	ctx       context.Context
	cancel    context.CancelFunc
	logger    hclog.Logger
	lastError error
}

func NewBuild() (*Build, error) {
	ctx, cancel := context.WithCancel(context.Background())

	opts := hclog.LoggerOptions{
		Color: hclog.AutoColor,
		Level: hclog.Info,
	}
	logger := hclog.New(&opts)

	client, err := dagger.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to Dagger Engine: %s", err)
	}

	return &Build{
		client: client,
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
	}, nil
}

func (b *Build) ContextWithTimeout(d time.Duration) context.Context {
	ctx, _ := context.WithTimeout(b.ctx, d)
	return ctx
}

func (b *Build) LogError(err error, params ...interface{}) {
	b.lastError = err
	b.logger.Error(err.Error(), params)
	b.cancel()
}

func (b *Build) LogStart(message string, params ...interface{}) func() {
	b.LogInfo(fmt.Sprintf("Starting %s", message), params)

	return func() {
		b.LogInfo(fmt.Sprintf("Finished %s", message), params)
	}
}

func (b *Build) LogInfo(message string, params ...interface{}) {
	b.logger.Info(message, params...)
}

func (b *Build) LastError() error {
	return b.lastError
}

func (b *Build) HasError() bool {
	return b.LastError() == nil
}

func (b *Build) Cancelled() bool {
	return b.ctx.Err() != nil
}
