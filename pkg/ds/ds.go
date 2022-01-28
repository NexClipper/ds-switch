package ds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/NexClipper/ds-switch/pkg/config"
)

type DataSourceInfo struct {
	ID          int    `json:"id"`
	UID         string `json:"uid"`
	OrgID       int    `json:"orgId"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	TypeName    string `json:"typeName"`
	TypeLogoURL string `json:"typeLogoUrl"`
	Access      string `json:"access"`
	URL         string `json:"url"`
	IsDefault   bool   `json:"isDefault"`
}

type DataSource struct {
	BaseURL     string
	BearerToken string
	GetAPI      string
	UpdateAPI   string

	Primary string
	Backup  string

	info []DataSourceInfo
}

func New(cfg *config.Config) *DataSource {
	ds := &DataSource{
		BaseURL:     cfg.Grafana.End_Point,
		BearerToken: cfg.Grafana.Bearer,
		GetAPI:      cfg.Grafana.DS_Get_API.Method,
		UpdateAPI:   cfg.Grafana.DS_Update_API.Method,
		Primary:     cfg.Grafana.DS_Name.Primary,
		Backup:      cfg.Grafana.DS_Name.Backup,
	}

	if err := ds.refreshDataSource(); err != nil {
		return nil
	}

	return ds
}

func (d *DataSource) refreshDataSource() error {
	url := fmt.Sprintf("%s/%s", d.BaseURL, d.GetAPI)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	bearerToken := fmt.Sprintf("Bearer %s", d.BearerToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", bearerToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var infos []DataSourceInfo
	if err := json.NewDecoder(resp.Body).Decode(&infos); err != nil {
		return err
	}

	d.info = infos

	return nil
}

func (d *DataSource) getDatasourceID(name string) int {
	for _, i := range d.info {
		if i.Name == name {
			return i.ID
		}
	}

	return -1
}

func (d *DataSource) setDefaultDatasource(defaultName string) {
	for i, item := range d.info {
		if item.Name == defaultName {
			d.info[i].IsDefault = true
		} else {
			d.info[i].IsDefault = false
		}
	}
}

func (d *DataSource) getDatasource(name string) *DataSourceInfo {
	for _, i := range d.info {
		if i.Name == name {
			return &i
		}
	}

	return nil
}

func (d *DataSource) SetDefaultDatasource(name string) error {
	if err := d.refreshDataSource(); err != nil {
		return err
	}

	datasourceID := d.getDatasourceID(name)
	if datasourceID == -1 {
		return fmt.Errorf("there is no Datasource")
	}

	d.setDefaultDatasource(name)

	dataSourceBuf := d.getDatasource(name)

	buf, err := json.Marshal(&dataSourceBuf)
	//log.Println(string(buf))

	url := fmt.Sprintf("%s/%s/%d", d.BaseURL, d.UpdateAPI, datasourceID)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(buf))
	if err != nil {
		return err
	}

	bearerToken := fmt.Sprintf("Bearer %s", d.BearerToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", bearerToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("[StatusCode:%d] %s", resp.StatusCode, resp.Status)
	}

	return nil

}
