package monitor

import (
	"fmt"
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
			fmt.Println(err)
			return
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			m.datasource.SetDefaultDatasource(m.datasource.Backup)
			m.lastStatus = false
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			log.Println("init 200 success")
			m.lastStatus = true
		} else {
			log.Println("init xxx fail")
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
		log.Println(err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		if m.lastStatus == true {
			m.datasource.SetDefaultDatasource(m.datasource.Backup)
			m.lastStatus = false
		}
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Println("200 success")
		if m.lastStatus == false {
			m.statusQueue.Enqueue(true)
		}
	} else {
		log.Println("xxx fail")
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
