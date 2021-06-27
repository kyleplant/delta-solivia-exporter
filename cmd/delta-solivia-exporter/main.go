package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-kit/kit/log/level"
	"github.com/kyleplant/delta-solivia-exporter/pkg/exporter"
	"github.com/lunixbochs/struc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/version"
)

var (
	metricsAddr = flag.String("metrics.addr", ":9134", "host:port for delta solivia exporter")
	metricsPath = flag.String("metrics.path", "/metrics", "URL path for surfacing collected metrics")

	table = []int{
		0x0000, 0xC0C1, 0xC181, 0x0140, 0xC301, 0x03C0, 0x0280, 0xC241,
		0xC601, 0x06C0, 0x0780, 0xC741, 0x0500, 0xC5C1, 0xC481, 0x0440,
		0xCC01, 0x0CC0, 0x0D80, 0xCD41, 0x0F00, 0xCFC1, 0xCE81, 0x0E40,
		0x0A00, 0xCAC1, 0xCB81, 0x0B40, 0xC901, 0x09C0, 0x0880, 0xC841,
		0xD801, 0x18C0, 0x1980, 0xD941, 0x1B00, 0xDBC1, 0xDA81, 0x1A40,
		0x1E00, 0xDEC1, 0xDF81, 0x1F40, 0xDD01, 0x1DC0, 0x1C80, 0xDC41,
		0x1400, 0xD4C1, 0xD581, 0x1540, 0xD701, 0x17C0, 0x1680, 0xD641,
		0xD201, 0x12C0, 0x1380, 0xD341, 0x1100, 0xD1C1, 0xD081, 0x1040,
		0xF001, 0x30C0, 0x3180, 0xF141, 0x3300, 0xF3C1, 0xF281, 0x3240,
		0x3600, 0xF6C1, 0xF781, 0x3740, 0xF501, 0x35C0, 0x3480, 0xF441,
		0x3C00, 0xFCC1, 0xFD81, 0x3D40, 0xFF01, 0x3FC0, 0x3E80, 0xFE41,
		0xFA01, 0x3AC0, 0x3B80, 0xFB41, 0x3900, 0xF9C1, 0xF881, 0x3840,
		0x2800, 0xE8C1, 0xE981, 0x2940, 0xEB01, 0x2BC0, 0x2A80, 0xEA41,
		0xEE01, 0x2EC0, 0x2F80, 0xEF41, 0x2D00, 0xEDC1, 0xEC81, 0x2C40,
		0xE401, 0x24C0, 0x2580, 0xE541, 0x2700, 0xE7C1, 0xE681, 0x2640,
		0x2200, 0xE2C1, 0xE381, 0x2340, 0xE101, 0x21C0, 0x2080, 0xE041,
		0xA001, 0x60C0, 0x6180, 0xA141, 0x6300, 0xA3C1, 0xA281, 0x6240,
		0x6600, 0xA6C1, 0xA781, 0x6740, 0xA501, 0x65C0, 0x6480, 0xA441,
		0x6C00, 0xACC1, 0xAD81, 0x6D40, 0xAF01, 0x6FC0, 0x6E80, 0xAE41,
		0xAA01, 0x6AC0, 0x6B80, 0xAB41, 0x6900, 0xA9C1, 0xA881, 0x6840,
		0x7800, 0xB8C1, 0xB981, 0x7940, 0xBB01, 0x7BC0, 0x7A80, 0xBA41,
		0xBE01, 0x7EC0, 0x7F80, 0xBF41, 0x7D00, 0xBDC1, 0xBC81, 0x7C40,
		0xB401, 0x74C0, 0x7580, 0xB541, 0x7700, 0xB7C1, 0xB681, 0x7640,
		0x7200, 0xB2C1, 0xB381, 0x7340, 0xB101, 0x71C0, 0x7080, 0xB041,
		0x5000, 0x90C1, 0x9181, 0x5140, 0x9301, 0x53C0, 0x5280, 0x9241,
		0x9601, 0x56C0, 0x5780, 0x9741, 0x5500, 0x95C1, 0x9481, 0x5440,
		0x9C01, 0x5CC0, 0x5D80, 0x9D41, 0x5F00, 0x9FC1, 0x9E81, 0x5E40,
		0x5A00, 0x9AC1, 0x9B81, 0x5B40, 0x9901, 0x59C0, 0x5880, 0x9841,
		0x8801, 0x48C0, 0x4980, 0x8941, 0x4B00, 0x8BC1, 0x8A81, 0x4A40,
		0x4E00, 0x8EC1, 0x8F81, 0x4F40, 0x8D01, 0x4DC0, 0x4C80, 0x8C41,
		0x4400, 0x84C1, 0x8581, 0x4540, 0x8701, 0x47C0, 0x4680, 0x8641,
		0x8201, 0x42C0, 0x4380, 0x8341, 0x4100, 0x81C1, 0x8081, 0x4040,
	}

	cmds = map[string][]interface{}{
		"\x10\x01": {"DC Cur1", 0, 10.0, "A"},
		"\x10\x02": {"DC Volts1", 0, 1, "V"},
		"\x10\x03": {"DC Pwr1", 0, 1, "W"},
		"\x10\x04": {"DC Cur2", 0, 10.0, "A"},
		"\x10\x05": {"DC Volts2", 0, 1, "V"},
		"\x10\x06": {"DC Pwr2", 0, 1, "W"},
		"\x10\x07": {"AC Current", 0, 10.0, "A"},
		"\x10\x08": {"AC Volts", 0, 1, "V"},
		"\x10\x09": {"AC Power", 0, 1, "W"},
		"\x11\x07": {"AC I Avg", 0, 10.0, "A"},
		"\x11\x08": {"AC V Avg", 0, 1, "V"},
		"\x11\x09": {"AC P Avg", 0, 1, "W"},
		"\x13\x03": {"Day Wh", 0, 1, "Wh"},
		"\x13\x04": {"Uptime", 0, 1, "min"},
		"\x00\x00": {"Inverter Type", 9, 0, ""},
		"\x00\x01": {"Serial", 1, 0, ""},
		"\x00\x08": {"Part", 1, 0, ""},
		"\x00\x40": {"FW Version", 10, 0, ""},
		"\x20\x05": {"AC Temp", 0, 1, "o"},
		"\x21\x08": {"DC Temp", 0, 1, "o"},
	}
)

func init() {
	prometheus.MustRegister(version.NewCollector("delta_solivia_exporter"))
}

func calcString(st []byte) int {
	// Given a bunary string and starting CRC, Calc a final CRC-16
	var crc = 0x0000
	for i := 0; i < len(st); i++ {
		fmt.Printf("%x ", st[i])
		crc = (crc >> 8) ^ table[(crc^int(st[i]))&0xFF]
	}

	return crc

	//for ch in st:
	//	crc = (crc >> 8) ^ self.table[(crc ^ ch) & 0xFF]
	//return crc
}

func findCmd(cmd string) {
	// for k, v in self.cmds.iteritems():
	// 	if(v[0] == strValue):
	// 		return k
	for key, element := range cmds {
		if element[0] == cmd {
			fmt.Println("Key:", key, "=>", "Element:", element)
		}
	}
}

type SoliviaCmd struct {
	A int `struc:"int8"`

	// B will be encoded/decoded as a 16-bit int (a "short")
	// but is stored as a native int in the struct
	B int `struc:"int8"`

	C int `struc:"int8"`

	// // the sizeof key links a buffer"s size to any int field
	// Size int `struc:"int8,little,sizeof=Str"`
	// Str  string

	// you can get freaky if you want
	Str2 string `struc:"[2]int8"`
}

func main() {
	var (
		opts = exporter.SerialOpts{}
	)

	promlogConfig := &promlog.Config{}
	logger := promlog.New(promlogConfig)

	// table := crc16.MakeTable(crc16.CRC16_MODBUS)

	// fmt.Printf("%v\n", table)

	// var myStruct CmdStruct
	// myStruct.A = 5
	// myStruct.B = 1
	// myStruct.C = 2
	// //myStruct.D = [2]string{"\x13", "\x03"}

	// binary.littleEndian.PutUint64(binary.LittleEndian, "5", 0x05)

	// buf := &bytes.Buffer{}
	// err := binary.Write(buf, binary.LittleEndian, myStruct)
	// fmt.Printf("Error: %v\n", err)
	// fmt.Printf("Sizeof myStruct: %d, Sizeof buf: %d, Len of buf: %d\n", unsafe.Sizeof(myStruct), unsafe.Sizeof(buf), buf.Len())

	// reply := &Reply{5, 1, 2, "\x13\x03"}
	// err := binary.Read(r, binary.BigEndian, &reply.Length)
	// if err != nil {
	// 	return nil, err
	// }
	// makeit = make([]byte, reply.Length+32)
	// _, err = r.Read(reply.Data)
	// if err != nil {
	// 	return nil, err
	// }
	// mt.Printf("%#v\n", makeit)

	// struct.pack(
	// 	"BBB{0}s".format(l), 5, self.inverterNum, l, cmd.encode("utf-8")

	findCmd("Day Wh")

	var buf bytes.Buffer
	t := &SoliviaCmd{5, 1, 2, "\x13\x03"}
	err := struc.Pack(&buf, t)
	//calcString(buf.Bytes())
	fmt.Printf("%d ", calcString(buf.Bytes()))
	o := &SoliviaCmd{}
	err = struc.Unpack(&buf, o)

	// bs := make([]byte, 5)
	// binary.LittleEndian.PutUint64(bs, uint64(5))
	// fmt.Printf("%#v\n", bs)

	// i := binary.LittleEndian.Uint64(bs)
	// fmt.Println(i)

	// crc := crc16.Checksum([]byte("\x13\x03"), table)
	// fmt.Printf("CRC-16 MAXIM: %X\n", crc)

	// // using the standard library hash.Hash interface
	// h := crc16.New(table)
	// h.Write([]byte("\x13\x03"))
	// fmt.Printf("CRC-16 MAXIM: %X\n", h.Sum16())

	level.Info(logger).Log("msg", "Starting delta_solivia_exporter", "version", version.Info())
	level.Info(logger).Log("build_context", version.BuildContext())

	exporter, err := exporter.New(opts, logger)
	if err != nil {
		level.Error(logger).Log("msg", "Error creating the exporter", "err", err)
		os.Exit(1)
	}

	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Delta Solivia Exporter</title></head>
             <body>
             <h1>Delta Solivia Exporter</h1>
             <p><a href="` + *metricsPath + `">Metrics</a></p>
             </body>
             </html>`))
	})

	http.HandleFunc("/-/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	http.HandleFunc("/-/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	level.Info(logger).Log("msg", "Listening on address", "address", *metricsAddr)
	if err := http.ListenAndServe(*metricsAddr, nil); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
