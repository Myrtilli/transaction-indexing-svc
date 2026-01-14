package config

import (
	"time"

	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Bitcoin interface {
	UseNodeRPC() bool
	UrlRpc() string
	UrlP2p() string
	NodeUser() string
	NodePass() string
	IndexerPollInterval() time.Duration
	StartHeight() int64
	ABC() string
}

type bitcoin struct {
	getter kv.Getter
	once   comfig.Once
}

type bitcoinConfig struct {
	UseNodeRPC   bool          `figure:"node_rpc"`
	UrlRpc       string        `figure:"url_rpc"`
	UrlP2p       string        `figure:"url_p2p"`
	User         string        `figure:"user"`
	Pass         string        `figure:"pass"`
	PollInterval time.Duration `figure:"poll_interval"`
	StartHeight  int64         `figure:"start_height"`
	ABC          string        `figure:"abc"`
}

func NewBitcoin(getter kv.Getter) Bitcoin {
	return &bitcoin{
		getter: getter,
	}
}

func (b *bitcoin) BitcoinConfig() *bitcoinConfig {
	return b.once.Do(func() any {
		var config bitcoinConfig
		raw := kv.MustGetStringMap(b.getter, "bitcoin")
		err := figure.Out(&config).From(raw).Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to get bitcoin config"))
		}

		return &config
	}).(*bitcoinConfig)
}

func (b *bitcoin) UrlRpc() string {
	return b.BitcoinConfig().UrlRpc
}

func (b *bitcoin) UrlP2p() string {
	return b.BitcoinConfig().UrlP2p
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

func (b *bitcoin) UseNodeRPC() bool {
	return b.BitcoinConfig().UseNodeRPC
}

func (b *bitcoin) ABC() string {
	return b.BitcoinConfig().ABC
}
