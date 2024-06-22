package watchdog

import (
	"log/slog"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/yiffyi/waterplz"
	"github.com/yiffyi/waterplz/upstream"
)

type Watchdog struct {
	Username     string
	Password     string
	ProjectId    int
	LoopInterval float64
	DeviceMac    string

	sess  *upstream.Session
	wecom *waterplz.WeComBot
	tz    *time.Location
	down  bool
}

func (w *Watchdog) notifyError(hint string) {
	md := "# 热水系统 Down\n"
	md += "当前时间：" + time.Now().In(w.tz).String() + "\n"
	md += `故障原因：<font color="warning">` + hint + `</font>`
	w.wecom.SendMarkdown(md)
}

func (w *Watchdog) notifyRecover() {
	md := "# 热水系统 Up\n"
	md += "时间：" + time.Now().In(w.tz).String() + "\n"
	w.wecom.SendMarkdown(md)
}

func (w *Watchdog) checkDeviceStatus() bool {
	deviceList, err := w.sess.DeviceInfoList(w.ProjectId, w.DeviceMac)
	if err != nil {
		slog.Error("checkDeviceStatus: upstream failure", "err", err)
		w.notifyError("上游 DeviceInfoList 返回错误")
		return false
	}

	if len(deviceList) != 1 {
		// however, if not found, it is still one anyway
		slog.Error("checkDeviceStatus: incorrect array len", "deviceList", deviceList, "err", err)
		w.notifyError("上游 DeviceInfoList 数组长度错误")
		return false
	}

	allGood := true
	for _, item := range deviceList {

		dev, ok := item.(map[string]interface{})
		if !ok {
			slog.Error("checkDeviceStatus: invalid array item", "item", item)
			w.notifyError("上游 DeviceInfoList 数组内容错误")
			continue
		}

		slog.Debug("checkDeviceStatus: device found", "dev", dev)

		m, ok := dev["devMac"].(string)
		if !ok || !strings.Contains(m, w.DeviceMac) {
			slog.Error("checkDeviceStatus: incorrect devMac returned from upstream", "w.mac", w.DeviceMac, "m", m)
			w.notifyError("上游 DeviceInfoList devMac 错误")
			allGood = false
			continue
		}

		ol, ok := dev["isOnline"].(float64)
		if ok && int(ol) == 1 {
			slog.Info("checkDeviceStatus: device is online", "mac", m)
		} else {
			slog.Error("checkDeviceStatus: device not found or offline", "mac", m, "ol", ol)
			w.notifyError("上游 DeviceInfoList 设备不存在或离线")
			allGood = false
		}
	}

	return allGood
}

func (w *Watchdog) checkPerMoney() bool {
	per, err := w.sess.PerMoney(w.ProjectId, w.ProjectId)
	if err != nil {
		slog.Error("checkPerMoney: upstream failure", "err", err)
		w.notifyError("上游 PerMoney 返回错误")
		return false
	} else {
		if per != "5.0" {
			slog.Error("checkPerMoney: unknown perMoney returned", "value", per)
			w.notifyError("上游 PerMoney 返回数据错误")
			return false
		}
	}
	slog.Info("checkPerMoney: check passed", "value", per, "desire", "5.0")
	return true
}

func (w *Watchdog) Start(wecom *waterplz.WeComBot) {

	// to avoid conflit, currently don't test functions require login
	w.sess = upstream.CreateAnonymousSession()
	w.down = false
	w.wecom = wecom
	var err error
	w.tz, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(time.Duration(w.LoopInterval) * time.Second)
	for {
		<-ticker.C

		// 1. test connectivity by query money per order
		if !w.checkPerMoney() {
			w.down = true
			continue
		}

		// 2. check if my device is online
		if !w.checkDeviceStatus() {
			w.down = true
			continue
		}

		// recovered ?
		if w.down {
			w.down = false
			w.notifyRecover()
		}
	}
}
