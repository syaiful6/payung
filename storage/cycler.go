package storage

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/syaiful6/payung/config"
	"github.com/syaiful6/payung/helper"
	"github.com/syaiful6/payung/logger"
	"github.com/syaiful6/payung/packager"
)

type PackageList []packager.Package

var (
	cyclerPath = path.Join(config.HomeDir, ".gobackup/cycler")
)

type Cycler struct {
	packages PackageList
	isLoaded bool
}

func (c *Cycler) add(backup packager.Package) {
	c.packages = append(c.packages, backup)
}

func (c *Cycler) shiftByKeep(keep int) (first *packager.Package) {
	total := len(c.packages)
	if total <= keep {
		return nil
	}

	first, c.packages = &c.packages[0], c.packages[1:]
	return
}

func (c *Cycler) run(model string, backup packager.Package, keep int, deletePackage func(backup *packager.Package) error) {
	cyclerFileName := path.Join(cyclerPath, model+".json")

	c.load(cyclerFileName)
	c.add(backup)
	defer c.save(cyclerFileName)

	if keep == 0 {
		return
	}

	for {
		pkg := c.shiftByKeep(keep)
		if pkg == nil {
			break
		}

		err := deletePackage(pkg)
		if err != nil {
			logger.Warn("remove failed: ", err)
		}
	}
}

func (c *Cycler) load(cyclerFileName string) {
	helper.MkdirP(cyclerPath)

	// write example JSON if not exist
	if !helper.IsExistsPath(cyclerFileName) {
		ioutil.WriteFile(cyclerFileName, []byte("[{}]"), os.ModePerm)
	}

	f, err := ioutil.ReadFile(cyclerFileName)
	if err != nil {
		logger.Error("Load cycler.json failed:", err)
		return
	}
	err = json.Unmarshal(f, &c.packages)
	if err != nil {
		logger.Error("Unmarshal cycler.json failed:", err)
	}
	c.isLoaded = true
}

func (c *Cycler) save(cyclerFileName string) {
	if !c.isLoaded {
		logger.Warn("Skip save cycler.json because it not loaded")
		return
	}

	data, err := json.Marshal(&c.packages)
	if err != nil {
		logger.Error("Marshal packages to cycler.json failed: ", err)
		return
	}

	err = ioutil.WriteFile(cyclerFileName, data, os.ModePerm)
	if err != nil {
		logger.Error("Save cycler.json failed: ", err)
		return
	}
}
