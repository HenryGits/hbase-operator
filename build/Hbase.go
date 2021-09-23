/**
 @author: ZHC
 @date: 2021-09-23 17:12:22
 @description: Hbase服务启动类
**/
package main

import (
	"context"
	"flag"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var (
	log   *zap.Logger
	level zapcore.Level

	nnDir     = flag.String("NameNodeDir", "/usr/local/hadoop/nn", "NameNode Directory")
	master = flag.Bool("master", false, "是否启动NameNode服务")
	regionserver = flag.Bool("regionserver", false, "是否启动DataNode服务")
	thrift = flag.Bool("thrift", false, "是否启动ResourceManager服务")
	thrift2 = flag.Bool("thrift2", false, "是否启动NodeManager服务")
)

func main() {
	log = Zap()
	ctx := context.Background()
	// Parse command line into the defined flags
	flag.Parse()

		path, _ := filepath.Abs(*nnDir)
		// 判断NameNode是否有效
		files, err := ioutil.ReadDir(path)
		if err != nil {
			log.Error("NameNode路径不存在!", zap.Error(err))
			os.Exit(1)
		}

		// 判断NameNode是否已被初始化过
		if len(files) <= 0 {
			cmd := exec.CommandContext(ctx, os.ExpandEnv("$HADOOP_HOME")+"/bin/hdfs", "namenode", "-format")
			out, err := cmd.CombinedOutput()
			log.Info(fmt.Sprintf("===NameNode Init...=== \n%s\n", string(out)))

			if err != nil {
				log.Error("NameNode初始化失败!", zap.Error(err))
				os.Exit(1)
			}
		}


	if *dnService {
		cmd := exec.CommandContext(ctx, os.ExpandEnv("$HADOOP_HOME")+"/bin/hdfs", "--daemon", "start", "datanode")
		out, err := cmd.CombinedOutput()
		log.Info(fmt.Sprintf("===DataNode Service...=== \n%s\n", string(out)))

		if err != nil {
			log.Error("DataNode启动失败!", zap.Error(err))
			os.Exit(1)
		}
	}

	if *rmService {
		cmd := exec.CommandContext(ctx, os.ExpandEnv("$HADOOP_HOME")+"/bin/yarn", "--daemon", "start", "resourcemanager")
		out, err := cmd.CombinedOutput()
		log.Info(fmt.Sprintf("===ResourceManager Service...=== \n%s\n", string(out)))

		if err != nil {
			log.Error("ResourceManager启动失败!", zap.Error(err))
			os.Exit(1)
		}
	}

	if *nmService {
		cmd := exec.CommandContext(ctx, os.ExpandEnv("$HADOOP_HOME")+"/bin/yarn", "--daemon", "start", "nodemanager")
		out, err := cmd.CombinedOutput()
		log.Info(fmt.Sprintf("===NodeManager Service...=== \n%s\n", string(out)))

		if err != nil {
			log.Error("NodeManager启动失败!", zap.Error(err))
			os.Exit(1)
		}
	}

	if *hsService {
		cmd := exec.CommandContext(ctx, os.ExpandEnv("$HADOOP_HOME")+"/bin/mapred", "--daemon", "start", "historyserver")
		out, err := cmd.CombinedOutput()
		log.Info(fmt.Sprintf("===HistoryServer Service...=== \n%s\n", string(out)))

		if err != nil {
			log.Error("HistoryServer启动失败!", zap.Error(err))
			os.Exit(1)
		}
	}

	fmt.Println(`
 ██      ██          ██  ██            ██      ██                ██                          
░██     ░██         ░██ ░██           ░██     ░██               ░██                   ██████ 
░██     ░██  █████  ░██ ░██  ██████   ░██     ░██  ██████       ░██  ██████   ██████ ░██░░░██
░██████████ ██░░░██ ░██ ░██ ██░░░░██  ░██████████ ░░░░░░██   ██████ ██░░░░██ ██░░░░██░██  ░██
░██░░░░░░██░███████ ░██ ░██░██   ░██  ░██░░░░░░██  ███████  ██░░░██░██   ░██░██   ░██░██████ 
░██     ░██░██░░░░  ░██ ░██░██   ░██  ░██     ░██ ██░░░░██ ░██  ░██░██   ░██░██   ░██░██░░░  
░██     ░██░░██████ ███ ███░░██████   ░██     ░██░░████████░░██████░░██████ ░░██████ ░██     
░░      ░░  ░░░░░░ ░░░ ░░░  ░░░░░░    ░░      ░░  ░░░░░░░░  ░░░░░░  ░░░░░░   ░░░░░░  ░░
	`)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		fmt.Println("Bye: context cancelled")
	case <-sigterm:
		fmt.Println("Bye: signal cancelled")
	}
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
	enc.AppendString(t.Format("[Hadoop] " + "2006-01-02 15:04:05.000"))
}
