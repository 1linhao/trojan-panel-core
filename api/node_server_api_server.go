package api

import (
	"context"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"sync"
	"time"
	"trojan-panel-core/util"
)

type NodeServerApiServer struct {
}

func (s *NodeServerApiServer) GetNodeServerInfo(ctx context.Context, nodeServerInfoDto *NodeServerInfoDto) (*Response, error) {
	if err := authRequest(ctx); err != nil {
		return &Response{Success: false, Msg: err.Error()}, nil
	}
	var cpuUsed, memUsed, diskUsed float64
	var uploadSpeed, downloadSpeed uint64
	var cpuErr, memErr, diskErr, networkErr error
	var waitGroup sync.WaitGroup
	waitGroup.Add(4)
	go func() { defer waitGroup.Done(); cpuUsed, cpuErr = util.GetCpuPercent() }()
	go func() { defer waitGroup.Done(); memUsed, memErr = util.GetMemPercent() }()
	go func() { defer waitGroup.Done(); diskUsed, diskErr = util.GetDiskPercent() }()
	go func() {
		defer waitGroup.Done()
		uploadSpeed, downloadSpeed, networkErr = util.GetNetworkSpeed(time.Second)
	}()
	waitGroup.Wait()
	for _, metricErr := range []error{cpuErr, memErr, diskErr, networkErr} {
		if metricErr != nil {
			return &Response{Success: false, Msg: metricErr.Error()}, nil
		}
	}
	nodeServerInfoVo := &NodeServerInfoVo{
		CpuUsed:                       float32(cpuUsed),
		MemUsed:                       float32(memUsed),
		DiskUsed:                      float32(diskUsed),
		NetworkUploadBytesPerSecond:   uploadSpeed,
		NetworkDownloadBytesPerSecond: downloadSpeed,
		SampledAt:                     time.Now().UnixMilli(),
	}
	data, err := anypb.New(proto.Message(nodeServerInfoVo))
	if err != nil {
		return &Response{Success: false, Msg: err.Error()}, nil
	}
	return &Response{Success: true, Msg: "", Data: data}, nil
}
