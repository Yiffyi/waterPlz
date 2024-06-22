package upstream

import "fmt"

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
