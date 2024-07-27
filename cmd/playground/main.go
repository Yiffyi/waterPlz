package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/yiffyi/waterplz/upstream"
)

func fromMac(mac string) int64 {
	b, err := hex.DecodeString(strings.ReplaceAll(mac, ":", ""))
	if err != nil {
		panic(err)
	}
	var x int64
	x = int64(b[5]) | (int64(b[4]) << 8) | (int64(b[3]) << 16) | (int64(b[2]) << 24) | (int64(b[1]) << 32) | (int64(b[0]) << 40)
	return x
}

func toMac(x int64) string {
	b := make([]byte, 6)
	b[5] = byte(x & 0xff)
	b[4] = byte((x >> 8) & 0xff)
	b[3] = byte((x >> 16) & 0xff)
	b[2] = byte((x >> 24) & 0xff)
	b[1] = byte((x >> 32) & 0xff)
	b[0] = byte((x >> 40) & 0xff)

	dst := make([]byte, hex.EncodedLen(len(b)))
	hex.Encode(dst, b)
	s := ""
	for i := 0; i <= 8; i += 2 {
		s += string(dst[i : i+2])
		s += ":"
	}
	s += string(dst[10:12])
	return s
}

func compileMacList(begin int64, end int64) string {
	var builder strings.Builder
	for i := begin; i < end; i++ {
		builder.WriteString(toMac(i))
		builder.WriteRune(',')
	}

	builder.WriteString(toMac(end))
	return builder.String()
}

const BATCH_SIZE = 100

func work() error {
	me := fromMac("c4:7f:0e:d5:4e:cb")
	sess := upstream.CreateAnonymousSession()

	fp, err := os.OpenFile("result.txt", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(fp)
	for i := me - 10000; i <= me+10000; i += BATCH_SIZE {
		macList := compileMacList(i, i+BATCH_SIZE-1)
		deviceList, err := sess.DeviceInfoList(4, macList)
		if err != nil {
			slog.Error("checkDeviceStatus: upstream failure", "err", err)
			return fmt.Errorf("上游 DeviceInfoList 请求错误: %w", err)
		}

		if len(deviceList) != BATCH_SIZE {
			// however, if not found, it is still one anyway
			slog.Error("checkDeviceStatus: incorrect array len", "deviceList", deviceList, "err", err)
			return errors.New("上游 DeviceInfoList 数组长度错误")
		}

		for _, item := range deviceList {

			dev, ok := item.(map[string]interface{})
			if !ok {
				slog.Error("checkDeviceStatus: invalid array item", "item", item)
				return errors.New("上游 DeviceInfoList 数组内容错误")
			}

			slog.Debug("checkDeviceStatus: device found", "dev", dev)

			m, ok := dev["devMac"].(string)
			if !ok {
				slog.Error("checkDeviceStatus: incorrect devMac returned from upstream", "m", m)
				return errors.New("上游 DeviceInfoList devMac 错误")
			}

			ol, ok := dev["isOnline"].(float64)
			if !ok {
				// slog.Error("checkDeviceStatus: device not found", "mac", m, "ol", ol)
				// return errors.New("上游 DeviceInfoList 设备不存在")
				continue
			} else {
				if int(ol) == 1 {
					slog.Info("checkDeviceStatus: device is online", "mac", m)
				} else {
					slog.Error("checkDeviceStatus: device offline", "mac", m, "ol", ol)
					// continue
					// return errors.New("上游 DeviceInfoList 设备离线")
				}
			}
			// fmt.Println(dev)
			enc.Encode(dev)
		}
	}
	return nil
}

func main() {
	work()
}
