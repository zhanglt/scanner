package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"github.com/neuvector/neuvector/share"
)

// global control data
type taskMain struct {
	ctx     context.Context
	outfile string
}

/////////////
func InitTaskMain(filename string) (*taskMain, bool) {
	tm := &taskMain{
		ctx:     context.Background(),
		outfile: filename,
	}
	return tm, true
}

// 扫描镜像库
func (tm *taskMain) ScanImage(req share.ScanImageRequest, imgPath string) (*share.ScanResult, error) {
	log.WithFields(log.Fields{
		"Registry": req.Registry, "image": fmt.Sprintf("%s:%s", req.Repository, req.Tag), "base": req.BaseImage,
	}).Debug()

	return cveTools.ScanImage(tm.ctx, &req, imgPath)
}

/////
func (tm *taskMain) ScanAppPackage(req share.ScanAppRequest) (*share.ScanResult, error) {
	log.WithFields(log.Fields{"packages": len(req.Packages)}).Debug()

	return cveTools.ScanAppPackage(&req, "")
}

/////
func (rs *taskMain) ScanImageData(data share.ScanData) (*share.ScanResult, error) {
	log.Debug()

	return cveTools.ScanImageData(&data)
}

/////
func (rs *taskMain) ScanAwsLambda(data share.ScanAwsLambdaRequest, imgPath string) (*share.ScanResult, error) {
	log.WithFields(log.Fields{"function": data.FuncName, "region": data.Region}).Debug()

	return cveTools.ScanAwsLambda(&data, imgPath)
}

///// worker
func (tm *taskMain) doScanTask(request interface{}, workingPath string) int {
	var err error
	var res *share.ScanResult

	switch request.(type) {
	case share.ScanImageRequest:
		log.WithFields(log.Fields{"扫描类型": "Registry"}).Info("开始扫描...")
		req := request.(share.ScanImageRequest)
		res, err = tm.ScanImage(req, workingPath)
	case share.ScanAppRequest:
		log.WithFields(log.Fields{"扫描类型": "APP"}).Info("开始扫描...")
		req := request.(share.ScanAppRequest)
		res, err = tm.ScanAppPackage(req)
	case share.ScanData:
		log.WithFields(log.Fields{"扫描类型": "Data"}).Info("开始扫描...")
		req := request.(share.ScanData)
		res, err = tm.ScanImageData(req)
	case share.ScanAwsLambdaRequest:
		log.WithFields(log.Fields{"扫描类型": "AWS"}).Info("开始扫描...")
		req := request.(share.ScanAwsLambdaRequest)
		res, err = tm.ScanAwsLambda(req, workingPath)
	default:
		err = errors.New("Invalid type")
	}

	if err != nil {
		return -1
	}

	// log.WithFields(log.Fields{"result": res}).Info("")
	// 反序列化结果数据
	data, _ := json.Marshal(res)
	// 将结果数据写入到结果文件中
	ioutil.WriteFile(tm.outfile, data, 0644)
	return 0
}
