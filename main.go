package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"jingxi.cn/voip_register/cmd"
	"jingxi.cn/voip_register/conf"
	"jingxi.cn/voip_register/utils"
	"os"
	"path/filepath"
	"time"
)

type CmdLine struct {
	isDebug       bool
	httpAddr      string
	freeswitchDir string
	container     string
	confDir       string
	logDir        string
}

var cmdline CmdLine

func init() {
	flag.BoolVar(&cmdline.isDebug, "debug", false, "true")
	flag.StringVar(&cmdline.httpAddr, "http", "0.0.0.0:9989", "E.g: 0.0.0.0:9989")
	flag.StringVar(&cmdline.container, "container", "voip", "freeswitch container name")
	flag.StringVar(&cmdline.freeswitchDir, "freeswitch", "/data/freeswitch/directory/default", "E.g: /data/freeswitch/directory/default")
	flag.StringVar(&cmdline.confDir, "conf", "", "/app/conf")
	flag.StringVar(&cmdline.logDir, "log", "", "/app/log")
}

func initLog(dir string) {
	path := filepath.Join(dir, "voip_regisger.log")
	writer, _ := rotatelogs.New(
		path+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(path),
		rotatelogs.WithMaxAge(time.Duration(24)*time.Hour),
		rotatelogs.WithRotationTime(time.Duration(8)*time.Hour))

	if cmdline.isDebug {
		writers := []io.Writer{
			writer,
			os.Stdout,
		}
		fileAndStdoutWriter := io.MultiWriter(writers...)
		logrus.SetOutput(fileAndStdoutWriter)
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetOutput(writer)
		logrus.SetLevel(logrus.ErrorLevel)
	}
	logrus.SetReportCaller(true)
}

func prepareEnv(serverConf *conf.ServerConfig) error {
	if len(serverConf.ConfDir) < 1 {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		exePath := filepath.Dir(exe)
		serverConf.LogDir = filepath.Join(exePath, "log")
		serverConf.ConfDir = filepath.Join(exePath, "conf")
	}
	data, err := ioutil.ReadFile(filepath.Join(serverConf.ConfDir, "1000.xml"))
	if err != nil {
		return err
	}
	serverConf.SourceXML = string(data)
	return nil
}

func main() {
	flag.Parse()

	var serverConfig conf.ServerConfig
	serverConfig.LogDir = cmdline.logDir
	serverConfig.ConfDir = cmdline.confDir
	serverConfig.FreeswitchDir = cmdline.freeswitchDir
	serverConfig.ContainerName = cmdline.container
	initLog(serverConfig.LogDir)

	err := prepareEnv(&serverConfig)
	if err != nil {
		logrus.Fatalf("start server error: %+v\n", err)
		return
	}
	if len(serverConfig.SourceXML) < 1 {
		logrus.Fatalf("not found 1000.xml or file empty\n")
		return
	}

	if len(cmdline.httpAddr) < 1 {
		logrus.Fatalf("http listen address empty")
		return
	}
	if len(cmdline.freeswitchDir) < 1 {
		logrus.Fatalf("not assign freeswitch dir")
		return
	}
	if len(cmdline.container) < 1 {
		logrus.Fatalf("not assign voip container name")
		return
	}

	_ = utils.SaveAppStartTime(serverConfig.LogDir)
	if !cmdline.isDebug {
		gin.SetMode(gin.ReleaseMode)
	}
	app := cmd.NewApp()
	app.Run(cmdline.httpAddr, &serverConfig)
}
