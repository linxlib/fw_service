package fw_service

import (
	"fmt"
	"github.com/kardianos/service"
	"os"
	"path/filepath"
	"runtime"
)

// RunAsService 以服务形式运行
// 封装 github.com/kardianos/service
// 提供 install, uninstall, start, stop 命令, 用于管理服务
// 只需在 main 函数中调用 RunAsService
// start传入的函数会在服务启动时执行(即你真实的启动代码)
// 支持windows,linux, 参考 github.com/kardianos/service
func RunAsService(name string, displayName string, description string, start func(), stop func()) {
	if service.Interactive() && len(os.Args) <= 1 {
		start()
		return
	}
	options := make(service.KeyValue)
	options["Restart"] = "on-failure"
	options["Type"] = "simple"
	options["StartLimitInterval"] = 3
	options["StartLimitBurst"] = 100
	execFile, _ := os.Executable()
	serviceConfig := &service.Config{
		Name:             name,
		DisplayName:      displayName,
		Description:      description,
		WorkingDirectory: filepath.Dir(execFile),
		UserName:         "root",
		Option:           options,
	}
	s := &FWService{start: start, stop: stop}
	s3, err := service.New(s, serviceConfig)
	if err != nil {
		fmt.Println(err)
		return
	}
	s.logger, _ = s3.Logger(nil)
	if len(os.Args) > 1 {
		serviceAction := os.Args[1]
		switch serviceAction {
		case "install":
			err := s3.Install()
			if err != nil {
				fmt.Printf("安装服务[%s]失败: %s\n", displayName, err.Error())
			} else {
				fmt.Printf("安装服务[%s]成功\n", displayName)
			}
			return
		case "uninstall":
			err := s3.Uninstall()
			if err != nil {
				fmt.Printf("卸载服务[%s]失败: %s\n", displayName, err.Error())
			} else {
				fmt.Printf("卸载服务[%s]成功\n", displayName)
			}
			return
		case "start":
			err := s3.Start()
			if err != nil {
				fmt.Printf("运行服务[%s]失败: %s\n", displayName, err.Error())
			} else {
				fmt.Printf("运行服务[%s]成功\n", displayName)
			}
			return
		case "stop":
			err := s3.Stop()
			if err != nil {
				fmt.Printf("停止服务[%s]失败: %s\n", displayName, err.Error())
			} else {
				fmt.Printf("停止服务[%s]成功\n", displayName)
			}
			return
		default:
			return
		}

	} else {
		s3.Run()
	}
}

type FWService struct {
	logger service.Logger
	start  func()
	stop   func()
}

func (F *FWService) Start(s service.Service) error {
	if service.Interactive() {
		F.logger.Info("Running in terminal.")
	} else {
		F.logger.Info("Running under service manager.")
		F.changeDir()
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				F.logger.Error(err)
			}
		}()
		F.start()
	}()
	return nil
}
func (F *FWService) changeDir() {
	// for windows not have a `WorkingDirectory` option
	if runtime.GOOS == "windows" {
		execFile, _ := os.Executable()
		os.Chdir(filepath.Dir(execFile))
	}
}

func (F *FWService) Stop(s service.Service) error {
	fmt.Println(s.String() + "服务正在停止...")
	F.stop()
	return nil
}
