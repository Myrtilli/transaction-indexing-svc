package config

import (
	"time"

	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type JWT interface {
	JWTKey() string
}

type jwt struct {
	getter kv.Getter
	once   comfig.Once
}

type jwtConfig struct {
	Key string `figure:"key"`
}

func NewJWT(getter kv.Getter) JWT {
	return &jwt{
		getter: getter,
	}
}

func (j *jwt) JWTKey() string {
	return j.once.Do(func() interface{} {
		var config jwtConfig
		raw := kv.MustGetStringMap(j.getter, "jwt_key")
		err := figure.Out(&config).From(raw).Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to get jwt key from config"))
		}

		return config.Key
	}).(string)
}

type Bitcoin interface {
	NodeURL() string
	NodeUser() string
	NodePass() string
	IndexerPollInterval() time.Duration
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
