package storage

import (
	"os"
	"path"
	"time"

	"github.com/syaiful6/payung/helper"
	"github.com/syaiful6/payung/packager"
	"golang.org/x/crypto/ssh"

	// "crypto/tls"
	"github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"github.com/syaiful6/payung/logger"
)

// SCP storage
//
// type: scp
// host: 192.168.1.2
// port: 22
// username: root
// password:
// timeout: 300
// private_key: ~/.ssh/id_rsa
type SCP struct {
	Base
	path       string
	host       string
	port       string
	privateKey string
	username   string
	password   string
	client     scp.Client
}

func (ctx *SCP) open() (err error) {
	ctx.viper.SetDefault("port", "22")
	ctx.viper.SetDefault("timeout", 300)
	ctx.viper.SetDefault("private_key", "~/.ssh/id_rsa")

	ctx.host = ctx.viper.GetString("host")
	ctx.port = ctx.viper.GetString("port")
	ctx.path = ctx.viper.GetString("path")
	ctx.username = ctx.viper.GetString("username")
	ctx.password = ctx.viper.GetString("password")
	ctx.privateKey = helper.ExplandHome(ctx.viper.GetString("private_key"))
	var clientConfig ssh.ClientConfig
	logger.Info("PrivateKey", ctx.privateKey)
	clientConfig, err = auth.PrivateKey(
		ctx.username,
		ctx.privateKey,
		ssh.InsecureIgnoreHostKey(),
	)
	if err != nil {
		logger.Warn(err)
		logger.Info("PrivateKey fail, Try User@Host with Password")
		clientConfig = ssh.ClientConfig{
			User:            ctx.username,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	}
	clientConfig.Timeout = ctx.viper.GetDuration("timeout") * time.Second
	if len(ctx.password) > 0 {
		clientConfig.Auth = append(clientConfig.Auth, ssh.Password(ctx.password))
	}

	ctx.client = scp.NewClient(ctx.host+":"+ctx.port, &clientConfig)

	err = ctx.client.Connect()
	if err != nil {
		return err
	}
	defer ctx.client.Session.Close()
	ctx.client.Session.Run("mkdir -p " + ctx.path)
	return
}

func (ctx *SCP) close() {}

func (ctx *SCP) upload(backupPackage *packager.Package) (err error) {
	err = ctx.client.Connect()
	if err != nil {
		return err
	}
	defer ctx.client.Session.Close()

	remotePath := ctx.RemotePath(ctx.path, backupPackage)

	fileNames := backupPackage.FileNames()
	// close files
	var files []*os.File
	defer func() {
		for i := range files {
			files[i].Close()
		}
	}()

	for i := range fileNames {
		src := path.Join(ctx.model.TempPath, fileNames[i])
		dest := path.Join(remotePath, fileNames[i])

		file, err := os.Open(src)
		if err != nil {
			return err
		}
		files = append(files, file)

		logger.Info("-> scp", dest)
		ctx.client.CopyFromFile(*file, dest, "0655")
	}

	logger.Info("Store successed")
	return nil
}

func (ctx *SCP) delete(backupPackage *packager.Package) (err error) {
	err = ctx.client.Connect()
	if err != nil {
		return
	}
	defer ctx.client.Session.Close()

	remotePath := ctx.RemotePath(ctx.path, backupPackage)

	logger.Info("-> remove", remotePath)
	err = ctx.client.Session.Run("rm -r " + remotePath)
	return
}
