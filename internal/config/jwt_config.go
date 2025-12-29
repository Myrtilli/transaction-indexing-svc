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
	JWTExpiration() time.Duration
}

type jwt struct {
	getter kv.Getter
	once   comfig.Once
}

type jwtConfig struct {
	Key        string        `figure:"key"`
	Expiration time.Duration `figure:"exp"`
}

func NewJWT(getter kv.Getter) JWT {
	return &jwt{
		getter: getter,
	}
}

func (j *jwt) JWTConfig() *jwtConfig {
	return j.once.Do(func() interface{} {
		var config jwtConfig
		raw := kv.MustGetStringMap(j.getter, "jwt_key")
		err := figure.Out(&config).From(raw).Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to get jwt config"))
		}

		return &config
	}).(*jwtConfig)
}

func (j *jwt) JWTKey() string {
	return j.JWTConfig().Key
}

func (j *jwt) JWTExpiration() time.Duration {
	return j.JWTConfig().Expiration
}
