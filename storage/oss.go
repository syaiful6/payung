package storage

import (
	"path"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/syaiful6/payung/logger"
	"github.com/syaiful6/payung/packager"
	"github.com/thatique/awan/verr"
)

// OSS - Aliyun OSS storage
//
// type: oss
// bucket: gobackup-test
// endpoint: oss-cn-beijing.aliyuncs.com
// path: /
// access_key_id: your-access-key-id
// access_key_secret: your-access-key-secret
// max_retries: 5
// timeout: 300
// threads: 1 (1 .. 100)
type OSS struct {
	Base
	endpoint        string
	bucket          string
	accessKeyID     string
	accessKeySecret string
	path            string
	maxRetries      int
	timeout         int
	client          *oss.Bucket
	threads         int
}

var (
	// 1 Mb one part
	ossPartSize int64 = 1024 * 1024
)

func (ctx *OSS) open() (err error) {
	ctx.viper.SetDefault("endpoint", "oss-cn-beijing.aliyuncs.com")
	ctx.viper.SetDefault("max_retries", 3)
	ctx.viper.SetDefault("path", "/")
	ctx.viper.SetDefault("timeout", 300)
	ctx.viper.SetDefault("threads", 1)

	ctx.endpoint = ctx.viper.GetString("endpoint")
	ctx.bucket = ctx.viper.GetString("bucket")
	ctx.accessKeyID = ctx.viper.GetString("access_key_id")
	ctx.accessKeySecret = ctx.viper.GetString("access_key_secret")
	ctx.path = ctx.viper.GetString("path")
	ctx.maxRetries = ctx.viper.GetInt("max_retries")
	ctx.timeout = ctx.viper.GetInt("timeout")
	ctx.threads = ctx.viper.GetInt("threads")

	// limit thread in 1..100
	if ctx.threads < 1 {
		ctx.threads = 1
	}
	if ctx.threads > 100 {
		ctx.threads = 100
	}

	logger.Info("endpoint:", ctx.endpoint)
	logger.Info("bucket:", ctx.bucket)

	ossClient, err := oss.New(ctx.endpoint, ctx.accessKeyID, ctx.accessKeySecret)
	if err != nil {
		return err
	}
	ossClient.Config.Timeout = uint(ctx.timeout)
	ossClient.Config.RetryTimes = uint(ctx.maxRetries)

	ctx.client, err = ossClient.Bucket(ctx.bucket)
	if err != nil {
		return err
	}

	return
}

func (ctx *OSS) close() {
}

func (ctx *OSS) upload(backupPackage *packager.Package) (err error) {
	remotePath := ctx.RemotePath(ctx.path, backupPackage)

	logger.Info("-> Uploading OSS...")

	fileNames := backupPackage.FileNames()
	for i := range fileNames {
		src := path.Join(ctx.model.TempPath, fileNames[i])
		dest := path.Join(remotePath, fileNames[i])
		if err = ctx.client.UploadFile(dest, src, ossPartSize, oss.Routines(ctx.threads)); err != nil {
			return err
		}
	}

	logger.Info("Success")

	return nil
}

func (ctx *OSS) delete(backupPackage *packager.Package) (err error) {
	remotePath := ctx.RemotePath(ctx.path, backupPackage)
	fileNames := backupPackage.FileNames()

	var errlist []error

	for i := range fileNames {
		err := ctx.client.DeleteObject(path.Join(remotePath, fileNames[i]))
		if err != nil {
			errlist = append(errlist, err)
		}
	}

	if len(errlist) > 0 {
		return verr.NewAggregate(errlist)
	}
	return nil
}
