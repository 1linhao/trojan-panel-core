package util

import (
	"testing"
	"time"

	gopsutilnet "github.com/shirou/gopsutil/net"
)

func TestCalculateNetworkSpeed(t *testing.T) {
	before := gopsutilnet.IOCountersStat{BytesSent: 1000, BytesRecv: 2000}
	after := gopsutilnet.IOCountersStat{BytesSent: 5000, BytesRecv: 8000}

	upload, download := calculateNetworkSpeed(before, after, 2*time.Second)
	if upload != 2000 {
		t.Fatalf("upload = %d, want 2000", upload)
	}
	if download != 3000 {
		t.Fatalf("download = %d, want 3000", download)
	}
}

func TestCalculateNetworkSpeedCounterReset(t *testing.T) {
	before := gopsutilnet.IOCountersStat{BytesSent: 5000, BytesRecv: 8000}
	after := gopsutilnet.IOCountersStat{BytesSent: 1000, BytesRecv: 2000}

	upload, download := calculateNetworkSpeed(before, after, time.Second)
	if upload != 0 || download != 0 {
		t.Fatalf("speeds = (%d, %d), want (0, 0)", upload, download)
	}
}

func TestAggregateNetworkCountersExcludesLoopback(t *testing.T) {
	stats := []gopsutilnet.IOCountersStat{
		{Name: "eth0", BytesSent: 1000, BytesRecv: 2000},
		{Name: "lo", BytesSent: 500, BytesRecv: 500},
		{Name: "wlan0", BytesSent: 3000, BytesRecv: 4000},
	}

	total := aggregateNetworkCounters(stats)
	if total.BytesSent != 4000 || total.BytesRecv != 6000 {
		t.Fatalf("totals = (%d, %d), want (4000, 6000)", total.BytesSent, total.BytesRecv)
	}
}
