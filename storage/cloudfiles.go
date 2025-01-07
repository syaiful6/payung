package storage

import (
	"context"
	"os"
	"path"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/objectstorage/v1/objects"
	"github.com/syaiful6/payung/helper"
	"github.com/syaiful6/payung/logger"
	"github.com/syaiful6/payung/packager"
	"github.com/syaiful6/payung/pkg/rackspace"
	"github.com/thatique/awan/verr"
)

// Rackspace - CloudFiles
//
// type: rackspace
// container: payung-test
// region: IAD
// path: backups
// username: your_username
// password: your_apikey
type CloudFiles struct {
	Base
	container string
	path      string
	service   *gophercloud.ServiceClient
}

func (ctx *CloudFiles) open() (err error) {
	ctx.viper.SetDefault("region", "IAD")
	client, err := rackspace.AuthenticatedClient(context.Background(), gophercloud.AuthOptions{
		Username: ctx.viper.GetString("username"),
		Password: ctx.viper.GetString("password"),
	})
	if err != nil {
		return err
	}

	service, err := rackspace.NewObjectStorageV1(client, gophercloud.EndpointOpts{Region: ctx.viper.GetString("region")})
	if err != nil {
		return err
	}

	ctx.container = ctx.viper.GetString("container")
	ctx.path = ctx.viper.GetString("path")

	ctx.service = service
	return
}

func (ctx *CloudFiles) close() {}

func (ctx *CloudFiles) upload(backupPackage *packager.Package) (err error) {
	remotePath := ctx.RemotePath(ctx.path, backupPackage)
	uploadLogger := logger.Tag("Storage CloudFiles")
	uploadLogger.Info("-> CloudFiles uploading...")

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

		opts := objects.CreateOpts{
			Content: progress.Reader,
		}
		result := objects.Create(context.Background(), ctx.service, ctx.container, dest, opts)
		_, err = result.Extract()
		if err != nil {
			return progress.Errorf("%v", err)
		}
		progress.Done(dest)
	}

	return nil
}

func (ctx *CloudFiles) delete(backupPackage *packager.Package) (err error) {
	remotePath := ctx.RemotePath(ctx.path, backupPackage)
	fileNames := backupPackage.FileNames()

	var (
		errlist []error
		dest    string
	)

	for i := range fileNames {
		dest = path.Join(remotePath, fileNames[i])
		result := objects.Delete(context.Background(), ctx.service, ctx.container, dest, objects.DeleteOpts{})
		_, err := result.Extract()
		if err != nil {
			errlist = append(errlist, err)
		}
	}

	if len(errlist) > 0 {
		return verr.NewAggregate(errlist)
	}

	return nil
}
