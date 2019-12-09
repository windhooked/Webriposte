package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	nodeFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wre_node_errors_total",
			Help: "Webriposte node errors total.",
		},
		[]string{"wrenode"},
	)
	nodeMarkerDelta = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wre_node_marker_delta",
			Help: "Webriposte Node Marker Delta",
		},
		[]string{"wrenode"},
	)
	nodeTimeDelta = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wre_node_time_delta",
			Help: "Webriposte Node Time Delta",
		},
		[]string{"wrenode"},
	)
)

type Node struct {
	NodeID          int
	GroupID         int
	Status          bool
	MarkerDelta     int64
	LastUpdateDelta int64
	Ip              string
}

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(nodeFailures)
	prometheus.MustRegister(nodeMarkerDelta)
	prometheus.MustRegister(nodeTimeDelta)
}

func readMapping(filename string, m map[int]string) {

	f, err := os.Open(filename)
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()

	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Panic(err)
	}

	for _, l := range lines {
		i, err := strconv.Atoi(l[0])
		if err == nil {
			//m[l[2]] = i
			m[i] = l[1]
		}

	}
}

func NodeList() []Node {
	cmdName := "ripostenode list"
	cmdArgs := strings.Fields(cmdName)

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:len(cmdArgs)]...)
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	var nodes []Node
	for {
		var n Node
		r := bufio.NewReader(stdout)
		line, _, err := r.ReadLine()
		s := string(line)
		//	fmt.Println(s)
		if strings.HasPrefix(s, "No Neighbors defined.") {
			break
		}
		// loop termination condition 1:  EOF.
		// this is the normal loop termination condition.
		if err == io.EOF {
			break
		}
		if len(strings.Fields(s)) == 3 {

			//(11014,1):      1,1,0   172.16.34.28
			fmt.Sscanf(strings.Fields(s)[0], "(%d,%d):", &n.GroupID, &n.NodeID)
			n.Ip = strings.Fields(s)[2]
			//	fmt.Printf("%d %d %s\n", n.GroupID, n.NodeID, n.Ip)

		}
		nodes = append(nodes, n)

	}
	return nodes

}

func NodeStatus(node *Node) {

	cmdName := fmt.Sprintf("ripostenode status %d %d", node.GroupID, node.NodeID)
	//fmt.Println(cmdName)
	cmdArgs := strings.Fields(cmdName)

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:len(cmdArgs)]...)
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()
	r := bufio.NewReader(stdout)

	for {
		//Node status for node (10001,1) for group 10001
		//    Online Connected
		//    <Mark:<1:1353179><2:1534720><3:1448814><4:1266864><5:1156710><6:1042390><7:1130385><8:2820335><9:1121372><10:1180878><11:1063010><12:501992><13:776523><14:390028><15:902646><16:2382593><17:1677686><18:2101295><19:552015><20:112031><21:138409><30:19><52:13098><62:577284><63:1073492>>
		//Marker Delta: 0 messages
		//Last Receive Time Delta: 2668 milliseconds
		//End Marker: <Mark:<1:1353179><2:1534720><3:1448814><4:1266864><5:1156710><6:1042390><7:1130385><8:2820335><9:1121372><10:1180878><11:1063010><12:501992><13:776523><14:390028><15:902646><16:2382593><17:1677686><18:2101295><19:552015><20:112031><21:138409><30:19><52:13098><62:577284><63:1073492>>

		line, _, err := r.ReadLine()
		// loop termination condition 1:  EOF.
		// this is the normal loop termination condition.
		if err == io.EOF {
			//fmt.Print(line)
			break
		}

		//fmt.Printf(">%c<\n", line)
		s := string(line)
		//  Online Disconnected
		if strings.HasPrefix(s, "    Online Connected") {
			node.Status = true
			continue
		}
		// Online Connected
		if strings.HasPrefix(s, "    Online Disconnected") {
			node.Status = false
			continue
		}

		// RiposteGetNodeStatus failed.The node status for the specified neighbor is not available.  (0xC10200C0)
		if strings.HasPrefix(s, "The node status for the specified neighbor is not available") {
			//fmt.Printf("Error : %s\n", s)
			node.Status = false
			break
		}

		if strings.HasPrefix(s, "Node status for node (") {
			//	fmt.Printf("Node : %s\n", strings.Fields(s)[4])
			continue
		}
		//var e error
		//Marker Delta: 0 messages
		if strings.HasPrefix(s, "Marker Delta") {
			node.MarkerDelta, _ = strconv.ParseInt(strings.Fields(s)[2], 10, 64)
			//fmt.Printf("Delta marker: %d, %v\n", node.MarkerDelta, e)
			continue
		}
		//Last Receive Time Delta: 1139 milliseconds
		if strings.HasPrefix(s, "Last Receive") {
			node.LastUpdateDelta, _ = strconv.ParseInt(strings.Fields(s)[4], 10, 64)
			//fmt.Printf("Delta time: %d, %v\n", node.LastUpdateDelta, e)
			continue
		}

	}

	cmd.Wait()
	//fmt.Printf("Node > %v\n", node)
}
func main() {
	m := make(map[int]string) //WRE GID lookup Map
	readMapping("groupid.csv", m)

	go func() {

		for {
			select {
			case <-time.After(time.Minute * 5):

				for _, v := range NodeList() {
					NodeStatus(&v)
					if v.Status == false {
						nodeFailures.With(prometheus.Labels{"wrenode": fmt.Sprintf("%s:%d", m[v.GroupID], v.NodeID)}).Inc()
					}
					nodeMarkerDelta.With(prometheus.Labels{"wrenode": fmt.Sprintf("%s:%d", m[v.GroupID], v.NodeID)}).Set(float64(v.MarkerDelta))
					nodeTimeDelta.With(prometheus.Labels{"wrenode": fmt.Sprintf("%s:%d", m[v.GroupID], v.NodeID)}).Set(float64(v.LastUpdateDelta))
				}
			}
		}
	}()



	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))

}
