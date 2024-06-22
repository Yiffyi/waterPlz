package upstream

import (
	"fmt"
	"log/slog"
	"strconv"
)

func (s *Session) UserInfo() (map[string]interface{}, error) {
	result, err := s.httpGet("/user/info", nil)
	if err != nil {
		return nil, err
	}

	return result["data"].(map[string]interface{}), nil
}

func (s *Session) CreateOrder(SN string) error {
	result, err := s.httpPost("/order/downRate/snWater", map[string]string{
		"deviceSncode": SN,
	})

	fmt.Println(result)
	if err != nil {
		if result != nil {
			if int(result["errorCode"].(float64)) == 307 {
				s.timeId = result["data"].(map[string]interface{})["timeIds"].(string)
				return err
			} else {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (s *Session) CloseOrder(SN string) error {
	_, err := s.httpPost("/order/send/closeOrder", map[string]string{
		"deviceSncode": SN,
		"timeIds":      s.timeId,
	})

	return err
}

func (s *Session) PerMoney(mainTypeId int, projectId int) (string, error) {
	// do you really understand English?
	ret, err := s.httpGet("/order/perMoney", map[string]string{
		"mainTypeId": strconv.Itoa(mainTypeId),
		"projectId":  strconv.Itoa(projectId),
	})

	if err != nil {
		slog.Error("call GET /order/perMoney failed", "ret", ret, "err", err)
		return "", err
	}

	return ret["data"].(string), nil
}

// macList: "c4:7f:0e:d5:4e:cb,c4:7f:0e:d5:4e:cc"
func (s *Session) DeviceInfoList(projectId int, macList string) ([]interface{}, error) {
	ret, err := s.httpGet("/device/info/list", map[string]string{
		"macList":   macList,
		"projectId": strconv.Itoa(projectId),
	})
	if err != nil {
		slog.Error("call GET /device/info/list failed", "ret", ret, "err", err)
		return nil, err
	}

	return ret["data"].([]interface{}), nil
}
