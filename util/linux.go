package util

import (
	"errors"
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	gopsutilnet "github.com/shirou/gopsutil/net"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
	"strings"
	"time"
	"trojan-panel-core/model/constant"
)

// IsPortAvailable determine whether the port is available
func IsPortAvailable(port uint, network string) bool {
	if network == "tcp" {
		listener, err := net.ListenTCP(network, &net.TCPAddr{
			IP:   net.IPv4(0, 0, 0, 0),
			Port: int(port),
		})
		defer func() {
			if listener != nil {
				listener.Close()
			}
		}()
		if err != nil {
			logrus.Warnf("port %d is taken err: %s", port, err)
			return false
		}
	}
	if network == "udp" {
		listener, err := net.ListenUDP("udp", &net.UDPAddr{
			IP:   net.IPv4(0, 0, 0, 0),
			Port: int(port),
		})
		defer func() {
			if listener != nil {
				listener.Close()
			}
		}()
		if err != nil {
			logrus.Warnf("port %d is taken err: %s", port, err)
			return false
		}
	}
	return true
}

// GetLocalIP get local IP address
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		// check the ip address to determine whether the loopback address
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New(constant.GetLocalIPError)
}

// GetCpuPercent get CPU usage
func GetCpuPercent() (float64, error) {
	var err error
	percent, err := cpu.Percent(time.Second, false)
	value, err := strconv.ParseFloat(fmt.Sprintf("%.1f", percent[0]), 64)
	return value, err
}

// GetMemPercent get memory usage
func GetMemPercent() (float64, error) {
	var err error
	memInfo, err := mem.VirtualMemory()
	value, err := strconv.ParseFloat(fmt.Sprintf("%.1f", memInfo.UsedPercent), 64)
	return value, err
}

// GetDiskPercent get disk usage
func GetDiskPercent() (float64, error) {
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(fmt.Sprintf("%.1f", diskInfo.UsedPercent), 64)
}

// GetNetworkSpeed measures aggregate non-loopback traffic over the supplied
// interval. Sent bytes are upload and received bytes are download from the
// node server's point of view.
func GetNetworkSpeed(interval time.Duration) (upload, download uint64, err error) {
	if interval <= 0 {
		return 0, 0, errors.New("network sampling interval must be positive")
	}
	beforeStats, err := gopsutilnet.IOCounters(true)
	if err != nil {
		return 0, 0, err
	}
	if len(beforeStats) == 0 {
		return 0, 0, errors.New("no network counters found")
	}
	time.Sleep(interval)
	afterStats, err := gopsutilnet.IOCounters(true)
	if err != nil {
		return 0, 0, err
	}
	if len(afterStats) == 0 {
		return 0, 0, errors.New("no network counters found")
	}
	before := aggregateNetworkCounters(beforeStats)
	after := aggregateNetworkCounters(afterStats)
	upload, download = calculateNetworkSpeed(before, after, interval)
	return upload, download, nil
}

func aggregateNetworkCounters(stats []gopsutilnet.IOCountersStat) gopsutilnet.IOCountersStat {
	var total gopsutilnet.IOCountersStat
	for _, stat := range stats {
		name := strings.ToLower(stat.Name)
		if name == "lo" || strings.HasPrefix(name, "loopback") {
			continue
		}
		total.BytesSent += stat.BytesSent
		total.BytesRecv += stat.BytesRecv
	}
	return total
}

func calculateNetworkSpeed(before, after gopsutilnet.IOCountersStat, interval time.Duration) (upload, download uint64) {
	seconds := interval.Seconds()
	if seconds <= 0 {
		return 0, 0
	}
	if after.BytesSent >= before.BytesSent {
		upload = uint64(float64(after.BytesSent-before.BytesSent) / seconds)
	}
	if after.BytesRecv >= before.BytesRecv {
		download = uint64(float64(after.BytesRecv-before.BytesRecv) / seconds)
	}
	return upload, download
}
