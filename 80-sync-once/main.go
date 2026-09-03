package main

import (
	"sync"
	"sync/atomic"
)

func main() {}

type Config struct {
	db string
}

var (
	config    *Config
	syncOnce  sync.Once
	initCalls atomic.Int32
)

func Get() *Config {
	syncOnce.Do(func() {
		initCalls.Add(1)
		config = &Config{db: ".........."}
	})
	return config
}                      // lazy-loads via a package-level sync.Once
func InitCalls() int32 { return initCalls.Load() }

type LazyConfig struct {
	once sync.Once
	load func() *Config
	cfg  *Config
}

func NewLazyConfig(load func() *Config) *LazyConfig {
	return &LazyConfig{load: load}
}
func (l *LazyConfig) Get() *Config {
	l.once.Do(func() { l.cfg = l.load() })
	return l.cfg
}
func SameOnceTwice(first, second func()) {
	var o sync.Once
	o.Do(first)
	o.Do(second)
}

var racyCfg *Config

func RacyGet() *Config {
	if racyCfg == nil {
		racyCfg = &Config{db: "..."}
	}
	return racyCfg
}
