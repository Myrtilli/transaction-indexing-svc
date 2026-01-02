package config

import (
	"time"

	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Bitcoin interface {
	NodeURL() string
	NodeUser() string
	NodePass() string
	IndexerPollInterval() time.Duration
	StartHeight() int64
}

type bitcoin struct {
	getter kv.Getter
	once   comfig.Once
}

type bitcoinConfig struct {
	URL          string        `figure:"url"`
	User         string        `figure:"user"`
	Pass         string        `figure:"pass"`
	PollInterval time.Duration `figure:"poll_interval"`
	StartHeight  int64         `figure:"start_height"`
}

func NewBitcoin(getter kv.Getter) Bitcoin {
	return &bitcoin{
		getter: getter,
	}
}

func (b *bitcoin) BitcoinConfig() *bitcoinConfig {
	return b.once.Do(func() interface{} {
		var config bitcoinConfig
		raw := kv.MustGetStringMap(b.getter, "bitcoin")
		err := figure.Out(&config).From(raw).Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to get bitcoin config"))
		}

		return &config
	}).(*bitcoinConfig)
}

func (b *bitcoin) NodeURL() string {
	return b.BitcoinConfig().URL
}

func (b *bitcoin) NodeUser() string {
	return b.BitcoinConfig().User
}

func (b *bitcoin) NodePass() string {
	return b.BitcoinConfig().Pass
}

func (b *bitcoin) IndexerPollInterval() time.Duration {
	return b.BitcoinConfig().PollInterval
}

func (b *bitcoin) StartHeight() int64 {
	return b.BitcoinConfig().StartHeight
}
