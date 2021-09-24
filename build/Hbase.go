/**
 @author: ZHC
 @date: 2021-09-23 17:12:22
 @description: Hbase服务启动类
**/
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	log   *zap.Logger
	level zapcore.Level

	master       = flag.Bool("Master", false, "是否启动Master服务")
	regionServer = flag.Bool("RegionServer", false, "是否启动RegionServer服务")
	thrift       = flag.Bool("Thrift", false, "是否启动Thrift服务")
	thrift2      = flag.Bool("Thrift2", false, "是否启动Thrift2服务")
)

func main() {
	log = Zap()
	ctx := context.Background()
	// Parse command line into the defined flags
	flag.Parse()

	fmt.Println(`
 ██      ██          ██  ██            ██      ██ ██████                            
░██     ░██         ░██ ░██           ░██     ░██░█░░░░██                           
░██     ░██  █████  ░██ ░██  ██████   ░██     ░██░█   ░██   ██████    ██████  █████ 
░██████████ ██░░░██ ░██ ░██ ██░░░░██  ░██████████░██████   ░░░░░░██  ██░░░░  ██░░░██
░██░░░░░░██░███████ ░██ ░██░██   ░██  ░██░░░░░░██░█░░░░ ██  ███████ ░░█████ ░███████
░██     ░██░██░░░░  ░██ ░██░██   ░██  ░██     ░██░█    ░██ ██░░░░██  ░░░░░██░██░░░░ 
░██     ░██░░██████ ███ ███░░██████   ░██     ░██░███████ ░░████████ ██████ ░░██████
░░      ░░  ░░░░░░ ░░░ ░░░  ░░░░░░    ░░      ░░ ░░░░░░░   ░░░░░░░░ ░░░░░░   ░░░░░░
	`)

	if *master {
		err := CommandContext(ctx, os.ExpandEnv("$HBASE_HOME")+"/bin/hbase", "master", "start")
		if err != nil {
			log.Error("Master启动失败!", zap.Error(err))
			os.Exit(1)
		}
	}

	if *regionServer {
		err := CommandContext(ctx, os.ExpandEnv("$HBASE_HOME")+"/bin/hbase", "regionserver", "start")
		if err != nil {
			log.Error("RegionServer启动失败!", zap.Error(err))
			os.Exit(1)
		}
	}

	if *thrift {
		err := CommandContext(ctx, os.ExpandEnv("$HBASE_HOME")+"/bin/hbase", "thrift", "start")
		if err != nil {
			log.Error("Thrift启动失败!", zap.Error(err))
			os.Exit(1)
		}
	}

	if *thrift2 {
		err := CommandContext(ctx, os.ExpandEnv("$HBASE_HOME")+"/bin/hbase", "thrift2", "start")
		if err != nil {
			log.Error("Thrift2启动失败!", zap.Error(err))
			os.Exit(1)
		}
	}

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		fmt.Println("Bye: context cancelled")
	case <-sigterm:
		fmt.Println("Bye: signal cancelled")
	}
}

// CommandContext 执行shell实时输出日志
func CommandContext(ctx context.Context, name string, cmd ...string) error {
	c := exec.CommandContext(ctx, name, cmd...)
	stdout, err := c.StderrPipe()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		reader := bufio.NewReader(stdout)
		for {
			readString, err := reader.ReadString('\n')
			if err != nil || err == io.EOF {
				panic(err)
			}
			fmt.Print(readString)
		}
	}(&wg)
	err = c.Start()
	wg.Wait()
	return err
}

// Zap 初始日志zap
func Zap() (logger *zap.Logger) {
	level = zap.InfoLevel
	logger = zap.New(getEncoderCore())
	logger = logger.WithOptions(zap.AddCaller())
	return logger
}

// getEncoderConfig 获取zapcore.EncoderConfig
func getEncoderConfig() (config zapcore.EncoderConfig) {
	config = zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     CustomTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}
	config.EncodeLevel = zapcore.LowercaseLevelEncoder
	return config
}

// getEncoder 获取zapcore.Encoder
func getEncoder() zapcore.Encoder {
	return zapcore.NewConsoleEncoder(getEncoderConfig())
}

// getEncoderCore 获取Encoder的zapcore.Core
func getEncoderCore() (core zapcore.Core) {
	writer := zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout))
	return zapcore.NewCore(getEncoder(), writer, level)
}

// 自定义日志输出时间格式
func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("[HBase] " + "2006-01-02 15:04:05.000"))
}
