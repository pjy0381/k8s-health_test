package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
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

type PodInfo struct {
	Name   string
	Status string
}

var (
	formatString       = "%-20s %-15s\n"
	username           = "dspaas"
	inputCommand       = "m"
	previousMainInfoList [][]string
)

func countPodCondition(pods []PodInfo, statusCondition string) (int, error) {
	if len(statusCondition) == 0 {
		return 0, errors.New("podCondition is empty")
	}
	var count = 0
	for _, pod := range pods {
		if pod.Status == statusCondition {
			count++
		}
	}
	return count, nil
}

func parsePodInfo() ([]PodInfo, error) {
	cmd := exec.Command("kubectl", "get", "pods", "-o", "wide", "-A", "--no-headers")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(output), "\n")
	var podInfoList []PodInfo
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			podInfo := PodInfo{
				Name:   fields[0],
				Status: fields[3],
			}
			podInfoList = append(podInfoList, podInfo)
		}
	}
	return podInfoList, nil
}

func countNodeCondition(nodes []NodeInfo, statusCondition, serviceCondition string) ([]int, error) {
	if statusCondition == "" || serviceCondition == "" {
		return nil, errors.New("statusCondition or serviceCondition is empty")
	}
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
	return countN, nil
}

func getStatuses(IP string) (kubeletStatus, containerdStatus, sciniStatus string, err error) {
        cmd := exec.Command(
                "ssh", IP,
                "sudo systemctl status kubelet | grep Active | awk '{print \"kubelet:\"$2}' && "+
                        "sudo systemctl status containerd | grep Active | awk '{print \"containerd:\"$2}' && "+
                        "sudo systemctl status scini | grep Active | awk '{print \"scini:\"$2}'",
        )
        out, err := cmd.CombinedOutput()
        if err != nil {
                return "", "", "", fmt.Errorf("error executing command: %v", err)
        }
        output := string(out)
        statuses := strings.Split(output, "\n")
        for _, status := range statuses {
                if strings.Contains(status, "kubelet") {
                        kubeletStatus = strings.TrimPrefix(status, "kubelet:")
                        kubeletStatus = strings.TrimSpace(kubeletStatus)
                }
                if strings.Contains(status, "containerd") {
                        containerdStatus = strings.TrimPrefix(status, "containerd:")
                        containerdStatus = strings.TrimSpace(containerdStatus)
                }
                if strings.Contains(status, "scini") {
                        sciniStatus = strings.TrimPrefix(status, "scini:")
                        sciniStatus = strings.TrimSpace(sciniStatus)
                }
        }
        return kubeletStatus, containerdStatus, sciniStatus, nil
}

func insertNodeDefaultInfo() []NodeInfo {
        cmd := exec.Command("kubectl", "get", "nodes", "-o", "wide", "--no-headers")
        out, err := cmd.Output()
        if err != nil {
                fmt.Println("Error executing command:", err)
                return nil
        }
        lines := strings.Split(string(out), "\n")
        var nodes []NodeInfo
        var wg sync.WaitGroup
        var mutex sync.Mutex
        for _, line := range lines {
                fields := strings.Fields(line)
                if len(fields) < 3 {
                        continue
                }
                node := NodeInfo{Name: fields[0], Status: fields[1]}
                nodes = append(nodes, node)
                wg.Add(1)
                go updateNodeInfo(fields, &nodes, &wg, &mutex)
        }
        wg.Wait()
        sort.Slice(nodes, func(i, j int) bool {
                return nodes[i].Name < nodes[j].Name
        })
        return nodes
}

func updateNodeInfo(f []string, nodes *[]NodeInfo, wg *sync.WaitGroup, mutex *sync.Mutex) {
        defer wg.Done()
        kubeletStatus, containerdStatus, sciniStatus, err := getStatuses(f[0])
        if err != nil {
                fmt.Println("Error getting statuses:", err)
                return
        }
        mutex.Lock()
        for i := range *nodes {
                if (*nodes)[i].Name == f[0] {
                        (*nodes)[i].Kubelet = kubeletStatus
                        (*nodes)[i].Containerd = containerdStatus
                        (*nodes)[i].Scini = sciniStatus
                        break
                }
        }
        mutex.Unlock()
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

func createMainInfoList(pods []PodInfo, nodes []NodeInfo) [][]string {
	var mainInfoList [][]string
	prCount, prErr := countPodCondition(pods, "Running")
	readyList, readyListErr := countNodeCondition(nodes, "Ready", "active")
	if readyListErr != nil || prErr != nil {
		fmt.Println("Error(readyList):", readyListErr)
	}
	mainInfoList = append(mainInfoList, []string{"Pods", fmt.Sprintf("%d/%d", prCount, len(pods))})
	mainInfoList = append(mainInfoList, []string{"", "", ""})
	mainInfoList = append(mainInfoList, []string{"Node", fmt.Sprintf("%d/%d", readyList[0], len(nodes))})
	mainInfoList = append(mainInfoList, []string{"kubelet", fmt.Sprintf("%d/%d", readyList[1], len(nodes))})
	mainInfoList = append(mainInfoList, []string{"containerD", fmt.Sprintf("%d/%d", readyList[2], len(nodes))})
	mainInfoList = append(mainInfoList, []string{"SCINI", fmt.Sprintf("%d/%d", readyList[3], len(nodes))})
	return mainInfoList
}

func printNodeDetailInfo(nodes []NodeInfo) {
    if nodes == nil {
        return
    }
    fmt.Printf("%-5s %-40s %-10s %-10s %-20s %-20s\n", "No.", "Node Name", "Status", "kubelet", "containerd", "scini")
    for idx, node := range nodes {
        fmt.Printf("%-5d %-40s %-10s %-10s %-20s %-20s\n", idx+1, node.Name, node.Status, node.Kubelet, node.Containerd, node.Scini)
    }
        fmt.Printf("\n\n\n")
        fmt.Println("Main: m")
        fmt.Println("Node List: n")
        fmt.Println("Save Current PODs: s")
        fmt.Println("View PODs Differences: v")
        fmt.Println("Check Storage Class: c")
        fmt.Println("Exit: q")
}

func printNodeInfo(mainInfoList [][]string) {
	fmt.Printf(formatString, "Service", "READY")
	for _, info := range mainInfoList {
		if len(info) < 2 {
			fmt.Printf("Error\n")
			continue
		}
		fmt.Printf(formatString, info[0], info[1])
	}
	fmt.Printf("\n\n\n")
	fmt.Println("Node List: n")
	fmt.Println("Save Current PODs: s")
	fmt.Println("View PODs Differences: v")
	fmt.Println("Check Storage Class: c")
	fmt.Println("Exit: q")
}

func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func refreshLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if inputCommand == "m" {
				nodes := insertNodeDefaultInfo()
				pods, err := parsePodInfo()
				if err != nil {
					fmt.Println("error")
				}
				mainInfoList := createMainInfoList(pods, nodes)
				if !reflect.DeepEqual(mainInfoList, previousMainInfoList) {
					clearScreen()
					printNodeInfo(mainInfoList)
					previousMainInfoList = mainInfoList
				}
			}
		}
	}
}

func getUserInput() string {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	return input
}

func main() {
        var command string
        go refreshLoop()
        for {
                command = getUserInput()
                if command == "q" {
                        clearScreen()
                        break
                } else if command == "n"{
                        clearScreen()
                        nodes := insertNodeDefaultInfo()
                        printNodeDetailInfo(nodes)
                } else if command == "m" {
                        clearScreen()
                        printNodeInfo(previousMainInfoList)
                }
                inputCommand = command
        }
}
