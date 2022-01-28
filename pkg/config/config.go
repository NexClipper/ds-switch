package config

import "github.com/jinzhu/configor"

type Config struct {
	APPName string

	DS_Switch struct {
		Monitor_Interval  int
		Evaluate_Interval int
	}

	Prometheus struct {
		End_Point   string
		Monitor_API struct {
			Type   string
			Method string
		}
	}
	Grafana struct {
		Bearer        string
		End_Point     string
		DS_Update_API struct {
			Type   string
			Method string
		}
		DS_Get_API struct {
			Type   string
			Method string
		}
		DS_Name struct {
			Primary string
			Backup  string
		}
	}
}

func New(configPath string) (*Config, error) {
	c := &Config{}
	if err := configor.Load(c, configPath); err != nil {
		return nil, err
	}
	return c, nil
}
