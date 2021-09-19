package storage

import (
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/syaiful6/payung/logger"
	"github.com/syaiful6/payung/packager"
	"github.com/thatique/awan/verr"
)

// S3 - Amazon S3 storage
//
// type: s3
// bucket: gobackup-test
// region: us-east-1
// path: backups
// access_key_id: your-access-key-id
// secret_access_key: your-secret-access-key
// max_retries: 5
// timeout: 300
type S3 struct {
	Base
	bucket string
	path   string
	client *s3manager.Uploader
}

func (ctx *S3) open() (err error) {
	ctx.viper.SetDefault("region", "us-east-1")
	cfg := aws.NewConfig()
	endpoint := ctx.viper.GetString("endpoint")
	if len(endpoint) > 0 {
		cfg.Endpoint = aws.String(endpoint)
	}
	cfg.Credentials = credentials.NewStaticCredentials(
		ctx.viper.GetString("access_key_id"),
		ctx.viper.GetString("secret_access_key"),
		ctx.viper.GetString("token"),
	)
	cfg.Region = aws.String(ctx.viper.GetString("region"))
	cfg.MaxRetries = aws.Int(ctx.viper.GetInt("max_retries"))

	ctx.bucket = ctx.viper.GetString("bucket")
	ctx.path = ctx.viper.GetString("path")

	sess := session.Must(session.NewSession(cfg))
	ctx.client = s3manager.NewUploader(sess)

	return
}

func (ctx *S3) close() {}

func (ctx *S3) upload(backupPackage *packager.Package) (err error) {
	remotePath := ctx.RemotePath(ctx.path, backupPackage)

	logger.Info("-> S3 Uploading...")

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

		f, err := os.Open(src)
		if err != nil {
			return err
		}
		files = append(files, f)

		input := &s3manager.UploadInput{
			Bucket: aws.String(ctx.bucket),
			Key:    aws.String(dest),
			Body:   f,
		}

		result, err := ctx.client.Upload(input)
		if err != nil {
			return err
		}
		logger.Info("=>", result.Location)
	}

	return nil
}

func (ctx *S3) delete(backupPackage *packager.Package) (err error) {
	remotePath := ctx.RemotePath(ctx.path, backupPackage)
	fileNames := backupPackage.FileNames()

	var errlist []error

	for i := range fileNames {
		input := &s3.DeleteObjectInput{
			Bucket: aws.String(ctx.bucket),
			Key:    aws.String(path.Join(remotePath, fileNames[i])),
		}
		_, err = ctx.client.S3.DeleteObject(input)
		if err != nil {
			errlist = append(errlist, err)
		}
	}

	if len(errlist) > 0 {
		return verr.NewAggregate(errlist)
	}
	return nil
}
