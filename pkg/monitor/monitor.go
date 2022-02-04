package monitor

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/NexClipper/ds-switch/pkg/ds"
	"github.com/NexClipper/ds-switch/pkg/queue"
	"github.com/go-co-op/gocron"
)

type Monitor struct {
	PrometheusURL   string
	monitorInterval int
	lastStatus      bool

	scheduler   *gocron.Scheduler
	statusQueue *queue.StatusQueue
	datasource  *ds.DataSource
}

func New(url string, monitorInterval, evaluateInterval int, datasource *ds.DataSource) *Monitor {
	m := &Monitor{
		PrometheusURL:   url,
		monitorInterval: monitorInterval,

		scheduler:   gocron.NewScheduler(time.UTC),
		statusQueue: queue.New(),
		datasource:  datasource,
	}

	cnt := evaluateInterval / monitorInterval
	mod := evaluateInterval % monitorInterval
	if mod > 0 {
		cnt += 1
	}

	m.statusQueue.RegistFire(uint(cnt), func() {
		// primary Datasource로 다시 연결한다.
		log.Println("primary datasource connect")
		m.datasource.SetDefaultDatasource(m.datasource.Primary)
		m.lastStatus = true
	})

	return m
}

func (m *Monitor) Run() {
	var init sync.Once
	init.Do(func() {
		req, err := http.NewRequest("GET", m.PrometheusURL, nil)
		if err != nil {
			log.Printf("init request build fail: %s\n", err.Error)
			m.datasource.SetDefaultDatasource(m.datasource.Backup)
			m.lastStatus = false
			return
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			log.Println("init backup datasource connect")
			m.datasource.SetDefaultDatasource(m.datasource.Backup)
			m.lastStatus = false
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			log.Printf("init %d success(primary datasource connect)\n", resp.StatusCode)
			m.datasource.SetDefaultDatasource(m.datasource.Primary)
			m.lastStatus = true
		} else {
			log.Printf("init %d fail(backup datasource connect)\n", resp.StatusCode)
			m.datasource.SetDefaultDatasource(m.datasource.Backup)
			m.lastStatus = false
		}

		m.lastStatus = false
	})

	m.scheduler.Every(m.monitorInterval).Second().Do(m.monitoring)
	m.scheduler.StartAsync()
}

func (m *Monitor) monitoring() {
	req, err := http.NewRequest("GET", m.PrometheusURL, nil)
	if err != nil {
		log.Printf("request build fail: %s\n", err.Error)
		if m.lastStatus == false {
			m.statusQueue.RemoveAll()
		} else {
			// backup datasource로 연결
			log.Println("backup datasource connect")
			m.datasource.SetDefaultDatasource(m.datasource.Backup)
			m.lastStatus = false
		}
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		if m.lastStatus == false {
			m.statusQueue.RemoveAll()
		} else {
			// backup datasource로 연결
			log.Println("backup datasource connect")
			m.datasource.SetDefaultDatasource(m.datasource.Backup)
			m.lastStatus = false
		}
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Printf("%d success\n", resp.StatusCode)
		if m.lastStatus == false {
			m.statusQueue.Enqueue(true)
		}
	} else {
		log.Printf("%d fail\n", resp.StatusCode)
		if m.lastStatus == false {
			m.statusQueue.RemoveAll()
		} else {
			// backup datasource로 연결
			log.Println("backup datasource connect")
			m.datasource.SetDefaultDatasource(m.datasource.Backup)
			m.lastStatus = false

		}
	}
}
