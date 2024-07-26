package storage

import (
	"math"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/syaiful6/payung/helper"
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
	bucket       string
	path         string
	storageClass string
	client       *s3manager.Uploader
}

func (ctx *S3) open() (err error) {
	ctx.viper.SetDefault("region", "us-east-1")
	ctx.viper.SetDefault("max_retries", 5)
	ctx.viper.SetDefault("storage_class", "STANDARD")
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
	ctx.storageClass = ctx.viper.GetString("storage_class")

	sess := session.Must(session.NewSession(cfg))
	ctx.client = s3manager.NewUploader(sess)

	return
}

func (ctx *S3) close() {}

func (ctx *S3) upload(backupPackage *packager.Package) (err error) {
	remotePath := ctx.RemotePath(ctx.path, backupPackage)
	uploadLogger := logger.Tag("Storage S3")
	uploadLogger.Info("-> S3 Uploading...")

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

		progress := helper.NewProgressBar(uploadLogger, f)
		input := &s3manager.UploadInput{
			Bucket:       aws.String(ctx.bucket),
			Key:          aws.String(dest),
			StorageClass: aws.String(ctx.storageClass),
			Body:         progress.Reader,
		}

		result, err := ctx.client.Upload(input, func(uploader *s3manager.Uploader) {
			// set the part size as low as possible to avoid timeouts and aborts
			// also set concurrency to 1 for the same reason
			var partSize int64 = 64 * 1024 * 1024 // 64MiB
			maxParts := progress.FileLength / partSize

			// 10000 parts is the limit for AWS S3. If the resulting number of parts would exceed that limit, increase the
			// part size as much as needed but as little possible
			if maxParts > 10000 {
				partSize = int64(math.Ceil(float64(progress.FileLength) / 10000))
			}

			uploader.Concurrency = 1
			uploader.LeavePartsOnError = false
			uploader.PartSize = partSize
		})
		if err != nil {
			return progress.Errorf("%v", err)
		}
		progress.Done(result.Location)
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
