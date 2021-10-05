package storage

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/auth"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/syaiful6/payung/logger"
	"github.com/syaiful6/payung/packager"
)

const singleShotUploadSizeCutoff int64 = 32 * (1 << 20)

type uploadChunk struct {
	data   []byte
	offset uint64
	close  bool
}

type Dropbox struct {
	Base
	workers int
	path    string
	client  files.Client
}

func (ctx *Dropbox) open() (err error) {
	config := dropbox.Config{
		Token: ctx.viper.GetString("access_token"),
	}
	ctx.viper.SetDefault("workers", 4)
	ctx.viper.SetDefault("path", "/")
	ctx.client = files.New(config)
	ctx.path = ctx.viper.GetString("path")
	if ctx.path[0] != '/' {
		ctx.path = "/" + ctx.path
	}
	ctx.workers = ctx.viper.GetInt("workers")

	return nil
}

func (ctx *Dropbox) close() {}

func (ctx *Dropbox) upload(backupPackage *packager.Package) (err error) {
	remotePath := ctx.RemotePath(ctx.path, backupPackage)

	logger.Info("-> Dropbox Uploading...")

	fileNames := backupPackage.FileNames()

	for i := range fileNames {
		ctx.uploadFile(remotePath, fileNames[i])
	}

	logger.Info("Success")
	return nil
}

func (ctx *Dropbox) delete(backupPackage *packager.Package) error {
	remotePath := ctx.RemotePath(ctx.path, backupPackage)
	fileNames := backupPackage.FileNames()

	var entries []*files.DeleteArg

	for i := range fileNames {
		entries = append(entries, files.NewDeleteArg(path.Join(remotePath, fileNames[i])))
	}

	if _, err := ctx.client.DeleteBatch(files.NewDeleteBatchArg(entries)); err != nil {
		return err
	}

	return nil
}

func (ctx *Dropbox) uploadFile(remotePath string, packageFile string) error {
	src := path.Join(ctx.model.TempPath, packageFile)
	dst := path.Join(remotePath, packageFile)

	commitInfo := files.NewCommitInfo(dst)
	commitInfo.Mode.Tag = "overwrite"
	ts := time.Now().UTC().Round(time.Second)
	commitInfo.ClientModified = &ts

	contents, err := os.Open(src)
	if err != nil {
		return err
	}
	defer contents.Close()

	contentsInfo, err := contents.Stat()
	if err != nil {
		return err
	}

	if contentsInfo.Size() > singleShotUploadSizeCutoff {
		return ctx.uploadChunk(contents, commitInfo, contentsInfo.Size())
	}

	if _, err := ctx.client.Upload(commitInfo, contents); err != nil {
		return err
	}

	return nil
}

func (ctx *Dropbox) uploadChunk(r io.Reader, commitInfo *files.CommitInfo, sizeTotal int64) error {
	startArgs := files.NewUploadSessionStartArg()
	startArgs.SessionType = &files.UploadSessionType{}
	startArgs.SessionType.Tag = files.UploadSessionTypeConcurrent
	res, err := ctx.client.UploadSessionStart(startArgs, nil)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	workCh := make(chan uploadChunk, ctx.workers)
	errCh := make(chan error, 1)
	for i := 0; i < ctx.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chunk := range workCh {
				cursor := files.NewUploadSessionCursor(res.SessionId, chunk.offset)
				args := files.NewUploadSessionAppendArg(cursor)
				args.Close = chunk.close

				if err := uploadOneChunk(ctx.client, args, chunk.data); err != nil {
					errCh <- err
				}
			}
		}()
	}

	written := int64(0)
	chunkSize := int64(1 << 24)
	for written < sizeTotal {
		data, err := ioutil.ReadAll(&io.LimitedReader{R: r, N: chunkSize})
		if err != nil {
			return err
		}
		expectedLen := chunkSize
		if written+chunkSize > sizeTotal {
			expectedLen = sizeTotal - written
		}
		if len(data) != int(expectedLen) {
			return fmt.Errorf("failed to read %d bytes from source", expectedLen)
		}

		chunk := uploadChunk{
			data:   data,
			offset: uint64(written),
			close:  written+chunkSize >= sizeTotal,
		}

		select {
		case workCh <- chunk:
		case err := <-errCh:
			return err
		}

		written += int64(len(data))
	}

	close(workCh)
	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
	}

	cursor := files.NewUploadSessionCursor(res.SessionId, uint64(written))
	args := files.NewUploadSessionFinishArg(cursor, commitInfo)
	_, err = ctx.client.UploadSessionFinish(args, nil)

	return err
}

func uploadOneChunk(dbx files.Client, args *files.UploadSessionAppendArg, data []byte) error {
	for {
		err := dbx.UploadSessionAppendV2(args, bytes.NewReader(data))
		if err != nil {
			return err
		}

		rl, ok := err.(auth.RateLimitAPIError)
		if !ok {
			return err
		}

		time.Sleep(time.Second * time.Duration(rl.RateLimitError.RetryAfter))
	}
}
