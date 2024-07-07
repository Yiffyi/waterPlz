package waterplz

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/yiffyi/gorad/notification"
	"github.com/yiffyi/waterplz/upstream"
)

type Watchdog struct {
	Username     string
	Password     string
	ProjectId    int
	LoopInterval float64
	DeviceMac    string

	sess       *upstream.Session
	wecom      *notification.WeComBot
	tz         *time.Location
	down       bool
	lastOnline time.Time
}

func (w *Watchdog) notifyError(err error) {
	md := "# 热水系统 Down\n"
	md += "当前时间：" + time.Now().In(w.tz).Format("2006-01-02 15:04:05 -0700 MST") + "\n"
	md += `故障原因：<font color="warning">` + err.Error() + `</font>`
	w.wecom.SendMarkdown(md)
}

func (w *Watchdog) notifyRecover() {
	now := time.Now().In(w.tz)
	md := "# 热水系统 Up\n"
	md += "当前时间：" + now.Format("2006-01-02 15:04:05 -0700 MST") + "\n"
	md += fmt.Sprintf(`故障持续时间：<font color="warning">%.1f 分钟</font>`, now.Sub(w.lastOnline).Minutes())
	w.wecom.SendMarkdown(md)
}

func (w *Watchdog) checkDeviceStatus() error {
	deviceList, err := w.sess.DeviceInfoList(w.ProjectId, w.DeviceMac)
	if err != nil {
		slog.Error("checkDeviceStatus: upstream failure", "err", err)
		return errors.New("上游 DeviceInfoList 返回错误")
	}

	if len(deviceList) != 1 {
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
		if !ok || !strings.Contains(m, w.DeviceMac) {
			slog.Error("checkDeviceStatus: incorrect devMac returned from upstream", "w.mac", w.DeviceMac, "m", m)
			return errors.New("上游 DeviceInfoList devMac 错误")
		}

		ol, ok := dev["isOnline"].(float64)
		if !ok {
			slog.Error("checkDeviceStatus: device not found", "mac", m, "ol", ol)
			return errors.New("上游 DeviceInfoList 设备不存在")
		} else {
			if int(ol) == 1 {
				slog.Info("checkDeviceStatus: device is online", "mac", m)
			} else {
				slog.Error("checkDeviceStatus: device offline", "mac", m, "ol", ol)
				return errors.New("上游 DeviceInfoList 设备离线")
			}
		}
	}

	return nil
}

func (w *Watchdog) checkPerMoney() error {
	per, err := w.sess.PerMoney(w.ProjectId, w.ProjectId)
	if err != nil {
		slog.Error("checkPerMoney: upstream failure", "err", err)
		return errors.New("上游 PerMoney 返回错误")
	} else {
		if per != "5.0" {
			slog.Error("checkPerMoney: unknown perMoney returned", "value", per)
			return errors.New("上游 PerMoney 返回数据错误")
		}
	}
	slog.Info("checkPerMoney: check passed", "value", per, "desire", "5.0")
	return nil
}

func (w *Watchdog) Start(bot *notification.WeComBot) {

	// to avoid conflit, currently don't test functions require login
	w.sess = upstream.CreateAnonymousSession()
	w.down = false
	w.wecom = bot
	var err error
	w.tz, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(time.Duration(w.LoopInterval) * time.Second)
	for {
		now := <-ticker.C

		// 1. test connectivity by query money per order
		// if !w.checkPerMoney() {
		// 	w.down = true
		// 	continue
		// }

		// 2. check if my device is online
		if err := w.checkDeviceStatus(); err != nil {
			if !w.down {
				w.notifyError(err)
				w.down = true
			}
			continue
		}

		// recovered ?
		if w.down {
			w.down = false
			w.notifyRecover()
		}
		w.lastOnline = now
	}
}
