package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
)

type NodeInfo struct {
	Name       string
	Status     string
	Kubelet    string
	Containerd string
	Scini      string
}

var (
	formatString = "%-20s %-15s %-15s\n"
	username     = "dspaas"
)

func countNodeCondition(nodes []NodeInfo, statusCondition, serviceCondition string) []int {
	var countN = make([]int, 4)

	for _, node := range nodes {
		if node.Status == statusCondition {
			countN[0]++
		}

		if node.Kubelet == serviceCondition {
			countN[1]++
		}

		if node.Containerd == serviceCondition {
			countN[2]++
		}

		if node.Scini == serviceCondition {
			countN[3]++
		}

	}
	return countN
}

func getStatuses(IP string) (kubeletStatus, containerdStatus, sciniStatus string) {
	cmd := exec.Command(
		"ssh", "-o", "StrictHostKeyChecking=no", username+"@"+IP,
		"sudo systemctl status kubelet | awk -F'[()]' '/Active:/ {print $2}';",
		"sudo systemctl status containerd | awk -F'[()]' '/Active:/ {print $2}';",
		"sudo systemctl status sciniStatus | awk -F'[()]' '/Active:/ {print $2}'")

	out, err := cmd.CombinedOutput()
	output := string(out)
	statuses := strings.Split(output, "\n")

	if err != nil {
		fmt.Println("Error executing command:", err)
		return "", "", ""
	}

	kubeletStatus = strings.TrimSpace(statuses[0])
	containerdStatus = strings.TrimSpace(statuses[1])
	sciniStatus = strings.TrimSpace(statuses[2])

	return kubeletStatus, containerdStatus, sciniStatus
}

func insertNodeDefaultInfo() []NodeInfo {
	cmd := exec.Command("kubectl", "get", "nodes", "-o", "wide")
	out, err := cmd.Output()

	if err != nil {
		fmt.Println("Error executing command:", err)
		return nil
	}

	var nodes []NodeInfo
	lines := strings.Split(string(out), "\n")
	var wg sync.WaitGroup

	var mutex sync.Mutex

	for count, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 6 || count == 0 {
			continue
		}

		wg.Add(1)

		go func(f []string) {
			defer wg.Done()

			kubeletStatus, containerdStatus, sciniStatus := getStatuses(f[5])
			node := NodeInfo{
				Name:       f[0],
				Status:     f[1],
				Kubelet:    kubeletStatus,
				Scini:      sciniStatus,
				Containerd: containerdStatus,
			}

			mutex.Lock()
			nodes = append(nodes, node)
			mutex.Unlock()
		}(fields)
	}
	wg.Wait()

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})

	return nodes
}

func getMainInfo(command string) []string {
	parts := strings.Split(command, " ")

	cmd := exec.Command(parts[0], parts[1:]...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		return []string{}
	}

	output := string(out)
	info := strings.Fields(output)

	if len(info) >= 3 {
		return info[:3]
	}

	return []string{}
}

func createMainInfoList(nodes []NodeInfo) [][]string {
	var mainInfoList [][]string

	result := getMainInfo("kubectl get deployment -n kube-system coredns --no-headers")
	result2 := getMainInfo("kubectl get deployment -n kubesphere-system ks-apiserver --no-headers")

	readyList := countNodeCondition(nodes, "Ready", "running")
	errorList := countNodeCondition(nodes, "Error", "error")

	mainInfoList = append(mainInfoList, result)
	mainInfoList = append(mainInfoList, result2)

	mainInfoList = append(mainInfoList, []string{"", "", ""})

	mainInfoList = append(mainInfoList, []string{"Node", fmt.Sprintf("%d/%d", readyList[0], len(nodes)), fmt.Sprintf("%d/%d", errorList[0], len(nodes))})
	mainInfoList = append(mainInfoList, []string{"kubelet", fmt.Sprintf("%d/%d", readyList[1], len(nodes)), fmt.Sprintf("%d/%d", errorList[1], len(nodes))})
	mainInfoList = append(mainInfoList, []string{"containerD", fmt.Sprintf("%d/%d", readyList[2], len(nodes)), fmt.Sprintf("%d/%d", errorList[2], len(nodes))})
	mainInfoList = append(mainInfoList, []string{"SCINI", fmt.Sprintf("%d/%d", readyList[3], len(nodes)), fmt.Sprintf("%d/%d", errorList[3], len(nodes))})

	return mainInfoList
}

func printNodeInfo(mainInfoList [][]string) {
	fmt.Printf(formatString, "Service", "READY", "UP-TO-DATE")

	for _, info := range mainInfoList {
		if len(info) < 3 {
			fmt.Printf("Error\n")
			continue
		}
		fmt.Printf(formatString, info[0], info[1], info[2])
	}

	fmt.Printf("\n\n\n")

	fmt.Println("Node List: n")
	fmt.Println("Save Current PODs: s")
	fmt.Println("View PODs Differences: v")
	fmt.Println("Check Storage Class: c")
}

func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func main() {
	ticker := time.NewTicker(5 * time.Second)

	for range ticker.C {
		clearScreen()
		nodes := insertNodeDefaultInfo()
		mainInfoList := createMainInfoList(nodes)
		printNodeInfo(mainInfoList)

	}
}
