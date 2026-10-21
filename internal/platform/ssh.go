package platform

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

const (
	sshAgentEnv = "SSH_AUTH_SOCK"

	sshDialTimeout    = 10 * time.Second
	sshCommandTimeout = 30 * time.Second
)

// sshClients are kept between the refreshes of the dashboard, to reuse the connections.
var sshClients = struct {
	sync.Mutex
	m map[string]*ssh.Client
}{m: map[string]*ssh.Client{}}

// sshClient connected to the address, reused if it was already connected.
func sshClient(username, addr, knownHostsFile string) (*ssh.Client, error) {
	key := username + "@" + addr + " " + knownHostsFile

	sshClients.Lock()
	defer sshClients.Unlock()

	if c, ok := sshClients.m[key]; ok {
		// The connection is alive if it answers.
		if _, _, err := c.SendRequest("keepalive@devdash", true, nil); err == nil {
			return c, nil
		}
		c.Close()
		delete(sshClients.m, key)
	}

	c, err := sshAgentAuth(username, addr, knownHostsFile)
	if err != nil {
		return nil, err
	}
	sshClients.m[key] = c

	return c, nil
}

// sshAgentAuth connects with the keys of the SSH agent, and verifies the key of the host with the known hosts file.
func sshAgentAuth(username, addr, knownHostsFile string) (*ssh.Client, error) {
	s := os.Getenv(sshAgentEnv)
	if s == "" {
		return nil, fmt.Errorf("the environment variable %s is empty: an SSH agent is needed to connect to %s", sshAgentEnv, addr)
	}

	if knownHostsFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		knownHostsFile = filepath.Join(home, ".ssh", "known_hosts")
	}

	hostKeyCallback, err := knownhosts.New(knownHostsFile)
	if err != nil {
		return nil, fmt.Errorf("can't read the known hosts %s: %w", knownHostsFile, err)
	}

	agentConn, err := net.Dial("unix", s)
	if err != nil {
		return nil, fmt.Errorf("can't connect to the SSH agent: %w", err)
	}
	defer agentConn.Close()

	config := &ssh.ClientConfig{
		User:              username,
		Auth:              []ssh.AuthMethod{ssh.PublicKeysCallback(agent.NewClient(agentConn).Signers)},
		HostKeyCallback:   hostKeyCallback,
		HostKeyAlgorithms: knownHostKeyAlgorithms(hostKeyCallback, addr),
		Timeout:           sshDialTimeout,
	}

	c, err := ssh.Dial("tcp", addr, config)
	var keyErr *knownhosts.KeyError
	if errors.As(err, &keyErr) {
		if len(keyErr.Want) == 0 {
			return nil, fmt.Errorf("the host %s isn't in %s: connect to it once with ssh to add it: %w", addr, knownHostsFile, err)
		}
		return nil, fmt.Errorf("the key of the host %s doesn't match %s: %w", addr, knownHostsFile, err)
	}

	return c, err
}

// knownHostKeyAlgorithms are the algorithms of the keys known for the host. The server could otherwise send a key
// of another type than the known one, which would fail the verification.
func knownHostKeyAlgorithms(callback ssh.HostKeyCallback, addr string) []string {
	// A key which can't be known gives the known keys in the error.
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil
	}

	var keyErr *knownhosts.KeyError
	if !errors.As(callback(addr, tcpAddr, unknownKey{}), &keyErr) {
		return nil
	}

	algos := []string{}
	for _, k := range keyErr.Want {
		switch k.Key.Type() {
		case ssh.KeyAlgoRSA:
			algos = append(algos, ssh.KeyAlgoRSASHA512, ssh.KeyAlgoRSASHA256, ssh.KeyAlgoRSA)
		default:
			algos = append(algos, k.Key.Type())
		}
	}

	return algos
}

// unknownKey is a public key which is in no known hosts file.
type unknownKey struct{}

func (unknownKey) Type() string                        { return "unknown" }
func (unknownKey) Marshal() []byte                     { return []byte("unknown") }
func (unknownKey) Verify([]byte, *ssh.Signature) error { return errors.New("unknown key") }
