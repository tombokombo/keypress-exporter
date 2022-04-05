package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Opts struct {
	InputDevice string `long:"input-device" short:"i" env:"INPUT_DEVICE" description:"Linux input device /dev/input/eventXYZ, hint look at /proc/bus/input/devices find keybord event num"`
	Port        string `long:"port" short:"p" env:"PORT" default:"9121" description:"port for binding metrics listener"`
}

type pressCollector struct {
	keyPressMetric *prometheus.Desc
}

func newPressCollector() *pressCollector {
	return &pressCollector{
		keyPressMetric: prometheus.NewDesc("key_press_counter",
			"key press counter",
			nil, nil,
		),
	}
}

func (collector *pressCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.keyPressMetric
}

func (collector *pressCollector) Collect(ch chan<- prometheus.Metric) {
	press := atomic.LoadUint64(&pressed)

	m1 := prometheus.MustNewConstMetric(collector.keyPressMetric, prometheus.CounterValue, float64(press))
	ch <- m1
}

var pressed uint64 = 0

func main() {

	opts := &Opts{}
	parser := flags.NewParser(opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if _, ok := err.(*flags.Error); ok {
			os.Exit(1)
		}
		fmt.Printf("Error parsing flags: %v", err)
	}

	input := opts.InputDevice
	if input == "" {
		fmt.Println("missing input device parameter see --help")
		os.Exit(1)
	}
	port := opts.Port

	f, err := os.Open(input)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	b := make([]byte, 24)

	go func() {
		var val int32
		for {
			f.Read(b)
			//sec := binary.LittleEndian.Uint64(b[0:8])
			//usec := binary.LittleEndian.Uint64(b[8:16])
			//t := time.Unix(int64(sec), int64(usec))
			var value int32
			binary.Read(bytes.NewReader(b[20:]), binary.LittleEndian, &value)
			if val != value && value > 2 {
				val = value
				//fmt.Println(t)
				//fmt.Printf("%d\n",val)
				atomic.AddUint64(&pressed, 1)
			}
		}
	}()

	pressCol := newPressCollector()
	prometheus.MustRegister(pressCol)

	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Unable to start exporter on port %s\n", port)
	}
}
