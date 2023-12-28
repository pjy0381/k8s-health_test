package main

import (
    "fmt"
    "os"
    "os/exec"
    "strings"
    "time"
    "sync"
    "sort"
)

type NodeInfo struct {
    Name       string
    Status     string
    Kubelet    string
    Containerd string
    Scini      string
}

func checkKubeletStatus(IP string) string {
    cmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "root@"+IP, "systemctl status kubelet | awk -F'[()]' '/Active:/ {print $2}'")
    out, err := cmd.Output()
    if err != nil {
        fmt.Println("Error executing command:", err)
        return ""
    }
    return string(out)
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

            node := NodeInfo{
                Name:    f[0],
                Status:  f[1],
                Kubelet: checkKubeletStatus(f[5]),
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

func printNodeInfo(nodes []NodeInfo) {
    if nodes == nil {
        return
    }

    fmt.Printf("%-5s %-40s %-10s %-10s %-20s %-20s\n", "No.", "Node Name", "Status", "kubelet", "containerd", "scini")

    for idx, node := range nodes {
        fmt.Printf("%-5d %-40s %-10s %-10s %-20s %-20s\n", idx+1, node.Name, node.Status, node.Kubelet, node.Containerd, node.Scini)
    }
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
        nodeList := insertNodeDefaultInfo()
				// 데이터 추가 실행 부분
        printNodeInfo(nodeList)
    }
}
