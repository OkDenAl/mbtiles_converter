package logging

import (
	"encoding/json"
	"go.uber.org/zap"
)

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

func New(config json.RawMessage) (Logger, error) {
	l, err := newZapLogger(config)
	if err != nil {
		return nil, err
	}
	return l, nil
}

type logger struct {
	*zap.SugaredLogger
}

func newZapLogger(config json.RawMessage) (Logger, error) {
	zapLogger, err := initLoggerFromConfig(config)
	if err != nil {
		return nil, err
	}
	return zapLogger, nil
}

func initLoggerFromConfig(jsonCfg json.RawMessage) (*logger, error) {
	cfg := zap.Config{}
	err := json.Unmarshal(jsonCfg, &cfg)
	if err != nil {
		return nil, err
	}

	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return &logger{l.Sugar()}, nil
}
