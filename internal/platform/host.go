package platform

// host execute a command on a remote host or on the local host.
// The output is parsed and displayed in the dashboard.
// A personalized command (or shell script) can be added to the config to render whatever output you want.

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"errors"

	"github.com/PiercingXX/devdash/internal/gokit"
)

// Host runs the commands of the widgets, on the local host or through SSH.
type Host struct {
	username, addr, knownHosts string
	localhost                  bool
}

// syntactic sugar
type runnerFunc func(cmd string) (string, error)

// NewHost connects to a remote host through SSH. The username and address "localhost" are the local host.
func NewHost(username, addr, knownHosts string) (*Host, error) {
	if username == "localhost" && addr == "localhost" {
		return &Host{localhost: true}, nil
	}

	h := &Host{username: username, addr: addr, knownHosts: knownHosts}
	if _, err := sshClient(username, addr, knownHosts); err != nil {
		return nil, err
	}

	return h, nil
}

// Runner runs a command on the remote server through SSH, or on the local host.
func (s *Host) Runner(command string) (string, error) {
	if s.localhost {
		return runLocalhost(command)
	}

	client, err := sshClient(s.username, s.addr, s.knownHosts)
	if err != nil {
		return "", err
	}

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("can't create session with SSH client for command %s: %w", command, err)
	}
	defer session.Close()

	var buf bytes.Buffer
	session.Stdout = &buf

	done := make(chan error, 1)
	go func() { done <- session.Run(command) }()

	select {
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("can't run command %s on remote server: %w", command, err)
		}
	case <-time.After(sshCommandTimeout):
		return "", fmt.Errorf("the command %s on the remote server didn't finish after %s", command, sshCommandTimeout)
	}

	return buf.String(), nil
}

// runLocalhost runs a command with the shell of the system.
func runLocalhost(command string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), sshCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	}

	var out, errs bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errs
	err := cmd.Run()
	if ctx.Err() != nil {
		return "", fmt.Errorf("the command %s didn't finish after %s", command, sshCommandTimeout)
	}
	if errs.Len() > 0 {
		return "", errors.New(strings.TrimSpace(errs.String()))
	}
	if err != nil {
		return "", fmt.Errorf("the command %s failed: %w", command, err)
	}

	return out.String(), nil
}

func HostUptime(runner runnerFunc) (int64, error) {
	command := "/bin/cat /proc/uptime"
	uptime, err := runner(command)
	if err != nil {
		return 0, err
	}

	d := strings.Fields(uptime)
	if len(d) < 1 {
		return 0, fmt.Errorf("command %s return nothing", command)
	}

	var secs float64
	secs, err = strconv.ParseFloat(d[0], strconv.IntSize)
	if err != nil {
		return 0, err
	}

	return int64(time.Duration(secs * 1e9)), nil
}

func HostLoad(runner runnerFunc) (string, error) {
	command := "/bin/cat /proc/loadavg"
	lines, err := runner(command)
	if err != nil {
		return "", err
	}
	res := strings.Fields(lines)
	if len(res) < 3 {
		return "", fmt.Errorf(
			"command %s return unexpected %v, needs to have 3 parts separated with whitespaces",
			command,
			res,
		)
	}

	return fmt.Sprintf("%s %s %s", res[0], res[1], res[2]), nil
}

func HostProcesses(runner runnerFunc) (string, error) {
	command := "/bin/cat /proc/loadavg"
	lines, err := runner(command)
	if err != nil {
		return "", err
	}

	res := strings.Fields(lines)
	if len(res) < 5 {
		return "", fmt.Errorf(
			"command %s return unexpected %v, needs to have 5 parts separated with whitespaces",
			command,
			res,
		)
	}

	runProc := "unknown"
	totalProc := "unknown"
	if i := strings.Index(res[3], "/"); i != -1 {
		runProc = res[3][0:i]
		if i+1 < len(res[3]) {
			totalProc = res[3][i+1:]
		}
	}

	return fmt.Sprintf("%s/%s", runProc, totalProc), nil
}

func HostMemory(runner runnerFunc, metrics []string, unit string) (val []int, err error) {
	command := "/bin/cat /proc/meminfo"
	lines, err := runner(command)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(lines))
	var data strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)

		if len(parts) < 3 {
			continue
		}

		val, err := strconv.ParseUint(parts[1], 10, strconv.IntSize)
		if err != nil {
			return nil, err
		}

		for _, v := range metrics {
			if strings.Trim(parts[0], ":") == v {
				data.WriteString(strconv.FormatUint(val, 10) + ",")
			}
		}
	}

	values := formatToBar(data.String())
	result := []int{}
	for _, v := range values {
		result = append(result, int(gokit.ConvertBinUnit(float64(v), "kb", unit)))
	}

	return result, nil
}

func HostMemoryRate(runner runnerFunc) (float64, error) {
	lines, err := runner("/bin/cat /proc/meminfo")
	if err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(strings.NewReader(lines))
	var memTotal float64 = 0
	var memFree float64 = 0
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)

		if len(parts) < 3 {
			continue
		}

		val, err := strconv.ParseFloat(parts[1], strconv.IntSize)
		if err != nil {
			return 0, err
		}

		k := strings.Trim(parts[0], ":")
		switch k {
		case "MemTotal":
			memTotal = gokit.ConvertBinUnit(val, "kb", "mb")
		case "MemFree":
			memFree = gokit.ConvertBinUnit(val, "kb", "mb")
		}

	}

	memUsed := memTotal - memFree

	// prevent division by 0
	if memTotal == 0 {
		return 0, nil
	}

	return gokit.Round(float64(memUsed)*100/float64(memTotal), 2), nil
}

func HostSwapRate(runner runnerFunc) (float64, error) {
	lines, err := runner("/bin/cat /proc/meminfo")
	if err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(strings.NewReader(lines))
	var swapTotal float64 = 0
	var swapFree float64 = 0
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)

		if len(parts) < 3 {
			continue
		}

		val, err := strconv.ParseFloat(parts[1], strconv.IntSize)
		if err != nil {
			return 0, err
		}

		k := strings.Trim(parts[0], ":")
		switch k {
		case "SwapTotal":
			swapTotal = val
		case "SwapFree":
			swapFree = val
		}

	}

	swapUsed := swapTotal - swapFree

	// prevent division by 0
	if swapTotal == 0 {
		return 0, nil
	}

	return gokit.Round(float64(swapUsed)*100/float64(swapTotal), 2), nil
}

// See https://www.idnt.net/en-US/kb/941772
// cpuSampleInterval between the two measures of the CPU, network or disk usage.
const cpuSampleInterval = "0.5"

// sampleSeparator between the two outputs of a sampled command.
const sampleSeparator = "==== devdash ===="

// sampled command: the output of the command, then its output after the interval.
func sampled(command string) string {
	return command + " && sleep " + cpuSampleInterval + " && echo '" + sampleSeparator + "' && " + command
}

// samples of the output of a sampled command. An output without separator is one sample.
func samples(output string) []string {
	return strings.Split(output, sampleSeparator)
}

// cpuTimes of the first line of /proc/stat: the idle time (with the I/O wait), and the total time.
func cpuTimes(stat string) (idle uint64, total uint64, err error) {
	line, _, _ := strings.Cut(strings.TrimSpace(stat), "\n")
	cpu := strings.Fields(line)
	if len(cpu) < 5 || cpu[0] != "cpu" {
		return 0, 0, fmt.Errorf("needs 5 fields for cpu: header, user, nice, system, idle. Instead, having %s", cpu)
	}

	for k, v := range cpu[1:] {
		n, _ := strconv.ParseUint(v, 10, 64)
		total += n
		// idle and iowait
		if k == 3 || k == 4 {
			idle += n
		}
	}

	return idle, total, nil
}

// HostCPURate is the CPU usage, in percent, during half a second. It's the average since the boot if the command
// doesn't give two samples.
func HostCPURate(runner runnerFunc) (float64, error) {
	raw, err := runner(sampled("/bin/cat /proc/stat"))
	if err != nil {
		return 0, err
	}

	s := samples(raw)
	idle, total, err := cpuTimes(s[0])
	if err != nil {
		return 0, err
	}

	if len(s) > 1 {
		idle2, total2, err := cpuTimes(s[1])
		if err != nil {
			return 0, err
		}
		if total2 > total {
			idle, total = idle2-idle, total2-total
		}
	}

	if total == 0 {
		return 0, nil
	}

	// Percentage of not idle (busy)
	return gokit.Round(100-float64(idle)*100/float64(total), 2), nil
}

// virtualInterfaces don't count in the network I/O: their traffic goes through a physical interface too.
// They're the loopback, the bridges and interfaces of containers and virtual machines, and the VPN tunnels.
var virtualInterfaces = []string{
	"lo", "docker", "veth", "br-", "virbr", "vnet", "cni", "flannel", "cali", "podman",
	"tailscale", "tun", "tap", "wg", "zt",
}

// netTotals received and transmitted by the network interfaces, in bytes.
func netTotals(netDev string) (rx uint64, tx uint64) {
	scanner := bufio.NewScanner(strings.NewReader(netDev))
	for scanner.Scan() {
		// The interface, then the received bytes, packets... and the transmitted bytes, packets...
		device, values, ok := strings.Cut(scanner.Text(), ":")
		parts := strings.Fields(values)
		if !ok || len(parts) < 9 {
			continue
		}

		device = strings.TrimSpace(device)
		if slices.ContainsFunc(virtualInterfaces, func(v string) bool { return strings.HasPrefix(device, v) }) {
			continue
		}

		r, _ := strconv.ParseUint(parts[0], 10, 64)
		t, _ := strconv.ParseUint(parts[8], 10, 64)
		rx += r
		tx += t
	}

	return rx, tx
}

// HostNetIO received / transmitted since the boot, or per second with rate.
func HostNetIO(runner runnerFunc, unit string, rate bool) (string, error) {
	command := "/bin/cat /proc/net/dev"
	if rate {
		command = sampled(command)
	}

	lines, err := runner(command)
	if err != nil {
		return "", err
	}

	s := samples(lines)
	rx, tx := netTotals(s[0])
	if rate && len(s) > 1 {
		rx2, tx2 := netTotals(s[1])
		rx, tx = perSecond(rx, rx2), perSecond(tx, tx2)
	}

	return formatIO(rx, tx, unit), nil
}

// perSecond between two samples.
func perSecond(before uint64, after uint64) uint64 {
	interval, _ := strconv.ParseFloat(cpuSampleInterval, 64)
	if after < before {
		return 0
	}

	return uint64(float64(after-before) / interval)
}

// formatIO of two values in bytes.
func formatIO(in uint64, out uint64, unit string) string {
	fi := gokit.ConvertBinUnit(float64(in), "b", unit)
	fo := gokit.ConvertBinUnit(float64(out), "b", unit)

	return strconv.FormatFloat(fi, 'f', 2, 64) + " / " + strconv.FormatFloat(fo, 'f', 2, 64)
}

func HostDisk(runner runnerFunc, headers []string, unit string) ([][]string, error) {
	// GetIOStat returns io stat
	lines, err := runner("/bin/df -x devtmpfs -x tmpfs -x debugfs")
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(lines))
	data := ""
	count := 0

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)

		if len(parts) < 6 || count == 0 {
			count++
			continue
		}

		filesystem := parts[0]
		size, _ := strconv.ParseFloat(parts[1], strconv.IntSize)
		used, _ := strconv.ParseFloat(parts[2], strconv.IntSize)
		available, _ := strconv.ParseFloat(parts[3], strconv.IntSize)
		useRate := parts[4]
		mount := parts[5]

		d := []string{
			filesystem,
			strconv.FormatFloat(gokit.ConvertBinUnit(size, "kb", unit), 'f', 2, strconv.IntSize) + unit,
			strconv.FormatFloat(gokit.ConvertBinUnit(used, "kb", unit), 'f', 2, strconv.IntSize) + unit,
			strconv.FormatFloat(gokit.ConvertBinUnit(available, "kb", unit), 'f', 2, strconv.IntSize) + unit,
			useRate,
			mount,
		}

		data += strings.Join(d, ",") + ","
		count++
	}

	data = strings.Trim(data, ",")

	c := [][]string{headers}
	c = append(c, formatToTable(len(headers), data)...)

	return c, nil
}

// diskTotals read and written on the physical disks, in bytes.
func diskTotals(diskstats string) (read uint64, write uint64) {
	type stat struct {
		name        string
		read, write uint64
	}
	stats := []stat{}
	names := []string{}
	scanner := bufio.NewScanner(strings.NewReader(diskstats))
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 10 || parts[3] == "0" {
			continue
		}

		// The sectors are always 512 bytes.
		r, _ := strconv.ParseUint(parts[5], 10, 64)
		w, _ := strconv.ParseUint(parts[9], 10, 64)
		stats = append(stats, stat{name: parts[2], read: r * 512, write: w * 512})
		names = append(names, parts[2])
	}

	for _, s := range stats {
		if isPhysicalDisk(s.name, names) {
			read += s.read
			write += s.write
		}
	}

	return read, write
}

// HostDiskIO read / written on the physical disks since the boot, or per second with rate.
func HostDiskIO(runner runnerFunc, unit string, rate bool) (string, error) {
	command := "/bin/cat /proc/diskstats"
	if rate {
		command = sampled(command)
	}

	lines, err := runner(command)
	if err != nil {
		return "", err
	}

	s := samples(lines)
	read, write := diskTotals(s[0])
	if rate && len(s) > 1 {
		read2, write2 := diskTotals(s[1])
		read, write = perSecond(read, read2), perSecond(write, write2)
	}

	return formatIO(read, write, unit), nil
}

// partitionSuffix of a partition's name after its disk's name, like "1" for sda1 or "p2" for nvme0n1p2.
var partitionSuffix = regexp.MustCompile(`^p?[0-9]+$`)

// isPhysicalDisk excludes the partitions, which would count the I/O of their disk twice,
// and the virtual devices (device mapper, RAID, loop, RAM disks...) which count the I/O of other disks, or none.
func isPhysicalDisk(name string, devices []string) bool {
	for _, prefix := range []string{"loop", "ram", "zram", "dm-", "md", "sr", "fd", "nbd"} {
		if strings.HasPrefix(name, prefix) {
			return false
		}
	}

	for _, d := range devices {
		if d != name && strings.HasPrefix(name, d) && partitionSuffix.MatchString(strings.TrimPrefix(name, d)) {
			return false
		}
	}

	return true
}

func HostBar(runner runnerFunc, command string) (data []int, err error) {
	lines, err := runner(command)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(lines))
	lineNumber := 0
	var str strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.FieldsSeq(line)

		for v := range parts {
			str.WriteString(v + ",")
		}
		lineNumber++
	}

	f := formatToBar(str.String())
	result := []int{}
	for _, f := range f {
		result = append(result, int(f))
	}
	return result, nil
}

func HostBox(runner runnerFunc, command string) (string, error) {
	lines, err := runner(command)
	if err != nil {
		return "", err
	}

	// The last new line of the output would be an empty line in the box.
	return strings.TrimRight(lines, "\n"), nil
}

func HostGauge(runner runnerFunc, command string) (float64, error) {
	lines, err := runner(command)
	if err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(strings.NewReader(lines))

	for scanner.Scan() {
		f, _ := strconv.ParseFloat(scanner.Text(), strconv.IntSize)
		return f, nil
	}

	return 0, nil
}

func HostTable(runner runnerFunc, command string, headers []string) (cells [][]string, err error) {
	lines, err := runner(command)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(lines))
	data := ""

	lineNumber := 0
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)

		if len(headers) == 0 {
			headers = parts
			lineNumber++
			continue
		}

		data += strings.Join(parts, ",") + ","
		lineNumber++
	}
	data = strings.Trim(data, ",")

	cells = [][]string{headers}
	cells = append(cells, formatToTable(len(headers), data)...)

	return
}

func formatToBar(data string) (val []uint64) {
	data = strings.Trim(data, ",")
	s := strings.Split(data, ",")
	val = []uint64{}
	for _, v := range s {
		k, _ := strconv.ParseUint(v, 10, strconv.IntSize)
		val = append(val, k)
	}

	return
}

// formatToTable splits comma separated values into rows of col cells. An incomplete last row is dropped.
func formatToTable(col int, data string) (cells [][]string) {
	c := strings.Split(data, ",")
	lenCells := len(c)
	for i := 0; i < lenCells; i += col {
		next := c[i:gokit.Min(i+col, lenCells)]
		if len(next) == col {
			cells = append(cells, next)
		}
	}

	return cells
}
