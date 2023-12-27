package main

import (
    "bytes"
    "fmt"
    "os"
    "os/exec"
    "strings"
    "time"
)

type NodeInfo struct {
    Name       string
    Status     string
    Kubelet    string
    Containerd string
    Scini      string
}

func insertNodeDefaultInfo() []NodeInfo {
    cmd := exec.Command("kubectl", "get", "nodes")

    var out bytes.Buffer
    var stderr bytes.Buffer

    cmd.Stdout = &out
    cmd.Stderr = &stderr

    err := cmd.Run()

    if err != nil {
        fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
        return nil
    }

    var nodes []NodeInfo

    lines := strings.Split(out.String(), "\n")

    for count, line := range lines {
        fields := strings.Fields(line)

        if len(fields) < 2 {
            continue
        }

        if count != 0 {
            node := NodeInfo{
                Name:   fields[0],
                Status: fields[1], 
            }
            nodes = append(nodes, node)
        }
    }

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