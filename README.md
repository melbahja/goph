<div align="center">
	<h1>Golang SSH Client.</h1>
	<a href="https://github.com/melbahja/goph">
		<img src="https://github.com/melbahja/goph/raw/master/.github/goph.png" width="200">
	</a>
	<h4 align="center">
		Fast and easy golang ssh client module.
	</h4>
	<p>Goph is a lightweight Go SSH client focusing on simplicity!</p>
</div>

<p align="center">
	<a href="#-features">Features</a> ❘
	<a href="#-installation">Installation</a> ❘
	<a href="#-get-started">Get Started</a> ❘
	<a href="#-usage-examples">Usage Examples</a> ❘
	<a href="#-license">License</a>
</p>

## 🤘&nbsp; Features

- Easy to use and **simple API**
- Supports **known hosts** by default.
- Supports connections with **passwords**.
- Supports connections with **private keys**.
- Supports connections with **protected private keys** with passphrase.
- Supports **multiple auth methods** in a single connection.
- Supports **upload** files from local to remote.
- Supports **download** files from remote to local.
- Supports connections with **ssh agent**.
- Supports connections with **custom signers**.
- Supports adding new hosts to **known_hosts file**.
- Supports host key callback check from **default known_hosts file**.
- Supports **file operations** like: `Open, Create, Chmod...` via SFTP.
- Supports **context.Context** for command cancellation.
- Supports **SOCKS5 proxy** for connecting through intermediaries.
- Supports **proxy jump** for connecting through jump hosts.

## 🚀&nbsp; Installation

```bash
go get github.com/melbahja/goph/v2
```

## 📄&nbsp; Get Started

Run a command via ssh can be simple as this example:
```go
package main

import (
	"log"
	"fmt"
	"github.com/melbahja/goph"
)

func main() {

	// Start new ssh connection with private key.
	client, err := goph.New("root", "192.1.1.3", goph.WithPassword("try_password_first"))
	if err != nil {
		log.Fatal(err)
	}

	// Defer closing the network connection.
	defer client.Close()

	// Execute your command.
	out, err := client.Run("ls /tmp/")

	if err != nil {
		log.Fatal(err)
	}

	// Get your output as []byte.
	fmt.Println(string(out))
}
```

Docs are available in [Go Docs](https://pkg.go.dev/github.com/melbahja/goph).

## 📚&nbsp; Usage Examples

Expand each group then expand individual examples.

### 🔐&nbsp; Authentication Examples

<details>
<summary>Password Authentication</summary>

```go
client, err := goph.New("root", "192.1.1.3",
	goph.WithPassword("your_password"),
)
```
</details>

<details>
<summary>Key File Authentication</summary>

```go
client, err := goph.New("root", "192.1.1.3",
	goph.WithKeyFile("/home/user/.ssh/id_rsa", ""),
)
```
</details>

<details>
<summary>Protected Private Key (with Passphrase)</summary>

```go
client, err := goph.New("root", "192.1.1.3",
	goph.WithKeyFile("/home/user/.ssh/id_rsa", "passphrase"),
)
```
</details>

<details>
<summary>Raw Private Key (from bytes)</summary>

```go
privateKey, _ := os.ReadFile("/home/user/.ssh/id_rsa")
client, err := goph.New("root", "192.1.1.3",
	goph.WithKey(privateKey, "passphrase"),
)
```
</details>

<details>
<summary>SSH Agent (auto-detect)</summary>

```go
if goph.HasAgent() {
	client, err := goph.New("root", "192.1.1.3",
		goph.WithDefaultAgent(),
	)
}
```
</details>

<details>
<summary>Custom Agent Socket Path</summary>

```go
client, err := goph.New("root", "192.1.1.3",
	goph.WithAgentSocket("/run/user/1000/keyring/ssh"),
)
```
</details>

<details>
<summary>Custom Agent Connection (net.Conn)</summary>

```go
conn, err := net.Dial("unix", "/custom/agent.sock")
if err != nil {
	log.Fatal(err)
}

client, err := goph.New("root", "192.1.1.3",
	goph.WithAgent(conn),
)
```
</details>

<details>
<summary>Keyboard-Interactive (password prompt)</summary>

```go
client, err := goph.New("root", "192.1.1.3",
	goph.WithKeyboardInteractive(func(user, instruction, question string, echo bool) (string, error) {
		if strings.Contains(strings.ToLower(question), "password") {
			return "your_password", nil
		}
		return "", fmt.Errorf("unexpected prompt: %s", question)
	}),
)
```
</details>

<details>
<summary>Keyboard-Interactive OTP / 2FA</summary>

```go
client, err := goph.New("root", "192.1.1.3",
	goph.WithKeyboardInteractive(func(user, instruction, question string, echo bool) (string, error) {
		if strings.Contains(strings.ToLower(question), "otp") ||
			strings.Contains(strings.ToLower(question), "verification code") {
			fmt.Print(question, " ")
			var code string
			fmt.Scanln(&code)
			return code, nil
		}
		return "", fmt.Errorf("unexpected prompt: %s", question)
	}),
)
```
</details>

<details>
<summary>Multiple Auth Methods</summary>

```go
client, err := goph.New("root", "192.1.1.3",
	goph.WithPassword("try_password_first"),
	goph.WithKeyFile("/home/user/.ssh/id_rsa", "passphrase"),
	goph.WithDefaultAgent(),
)
```
</details>

<details>
<summary>Custom Signer (any ssh.Signer)</summary>

```go
signer, err := ssh.NewSignerFromKey(myPrivateKey) // or from any source
if err != nil {
	log.Fatal(err)
}

client, err := goph.New("root", "192.1.1.3",
	goph.WithSigner(signer),
)
```
</details>

<details>
<summary>Custom Auth Method (ssh.AuthMethod)</summary>

```go
client, err := goph.New("root", "192.1.1.3",
	goph.WithAuth(ssh.Password("pass")),
	goph.WithAuth(ssh.PublicKeysCallback(myCustomCallback)),
)
```
</details>

### 📡&nbsp; Features Examples

<details>
<summary>Custom Port and Timeout</summary>

```go
client, err := goph.New("root", "192.1.1.3",
	goph.WithPassword("pass"),
	goph.WithPort(2222),
	goph.WithTimeout(10*time.Second),
)
```
</details>

<details>
<summary>Known Hosts Verification (Default)</summary>

```go
// Known hosts verification is ON by default via ~/.ssh/known_hosts
client, err := goph.New("root", "192.1.1.3",
	goph.WithPassword("pass"),
)

// Or specify a custom known_hosts file:
client, err := goph.New("root", "192.1.1.3",
	goph.WithPassword("pass"),
	goph.WithKnownHosts("/path/to/known_hosts"),
)
```
</details>

<details>
<summary>Disable Host Key Verification (Insecure)</summary>

```go
client, err := goph.New("root", "192.1.1.3",
	goph.WithPassword("pass"),
	goph.WithInsecureIgnoreHostKey(),
)
```
</details>

<details>
<summary>Custom Host Key Callback</summary>

```go
hostKeyCallback := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
	// Check the key against a database or prompt the user
	return nil
}

client, err := goph.New("root", "192.1.1.3",
	goph.WithPassword("pass"),
	goph.WithHostKeyCallback(hostKeyCallback),
)
```
</details>

<details>
<summary>Banner Callback</summary>

```go
client, err := goph.New("root", "192.1.1.3",
	goph.WithPassword("pass"),
	goph.WithBannerCallback(func(message string) error {
		log.Printf("SSH Banner: %s", message)
		return nil
	}),
)
```
</details>

<details>
<summary>Connect Through SOCKS5 Proxy</summary>

```go
client, err := goph.New("root", "target-host",
	goph.WithPassword("pass"),
	goph.WithProxy("socks5://127.0.0.1:1080"),
)

// With proxy authentication:
client, err = goph.New("root", "target-host",
	goph.WithPassword("pass"),
	goph.WithProxy("socks5://proxyuser:proxypass@127.0.0.1:1080"),
)
```
</details>

<details>
<summary>Connect Through Jump Host (Bastion)</summary>

```go
// First connect to the jump host
jump, err := goph.New("jumpuser", "bastion.example.com",
	goph.WithKeyFile("/home/user/.ssh/id_rsa", ""),
	goph.WithPort(2222),
)
if err != nil {
	log.Fatal(err)
}
defer jump.Close()

// Then tunnel through it to the target
client, err := goph.New("root", "internal-host",
	goph.WithKeyFile("/home/user/.ssh/id_rsa", ""),
	goph.WithJump(jump),
)
```
</details>

<details>
<summary>Proxy + Jump Host Together</summary>

```go
// Put the proxy on the JUMP client, then pass the jump to the target
jump, err := goph.New("jumpuser", "bastion.example.com",
	goph.WithKeyFile("/home/user/.ssh/id_rsa", ""),
	goph.WithProxy("socks5://127.0.0.1:1080"),
)
if err != nil {
	log.Fatal(err)
}
defer jump.Close()

client, err := goph.New("root", "internal-host",
	goph.WithKeyFile("/home/user/.ssh/id_rsa", ""),
	goph.WithJump(jump),
)
```
</details>

<details>
<summary>Run a Command</summary>

```go
out, err := client.Run("ls /tmp/")
if err != nil {
	log.Fatal(err)
}
fmt.Println(string(out))
```
</details>

<details>
<summary>Run with Timeout (Context)</summary>

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()

// Sends SIGINT and returns error after 1 second
out, err := client.RunContext(ctx, "sleep 5")
```
</details>

<details>
<summary>Cancel from Another Goroutine</summary>

```go
ctx, cancel := context.WithCancel(context.Background())
cmd, err := client.CommandContext(ctx, "sleep", "60")
if err != nil {
	log.Fatal(err)
}

// Cancel from another goroutine
go func() {
	time.Sleep(5 * time.Second)
	cancel()
}()

if err := cmd.Run(); err != nil {
	// context canceled
	fmt.Println("Command stopped:", err)
}
```
</details>

<details>
<summary>Custom Cancel Function (override SIGINT)</summary>

```go
cmd, err := client.CommandContext(ctx, "sleep", "30")
if err != nil {
	log.Fatal(err)
}

// Send SIGKILL instead of the default SIGINT on cancellation
cmd.Cancel = func() error {
	return cmd.Signal(ssh.SIGKILL)
}

if err := cmd.Run(); err != nil {
	log.Fatal(err)
}
```
</details>

<details>
<summary>Using Cmd for Fine Control</summary>

```go
cmd, err := client.Command("ls", "-alh", "/tmp")
if err != nil {
	log.Fatal(err)
}

// Set env vars (server must accept them)
cmd.Env = []string{"MY_VAR=MYVALUE"}

// Run (CombinedOutput, Output, Start, Wait also available)
err = cmd.Run()

// With context:
cmd, err = client.CommandContext(ctx, "ls", "-alh", "/tmp")
```
</details>

<details>
<summary>Upload File</summary>

```go
err := client.Upload("/path/to/local/file", "/path/to/remote/file")
```
</details>

<details>
<summary>Download File</summary>

```go
err := client.Download("/path/to/remote/file", "/path/to/local/file")
```
</details>

<details>
<summary>SFTP File Operations</summary>

```go
sftp, err := client.NewSftp()
if err != nil {
	log.Fatal(err)
}
defer sftp.Close()

// Create a file
file, err := sftp.Create("/tmp/remote_file")
if err != nil {
	log.Fatal(err)
}
defer file.Close()

file.Write([]byte("Hello world"))

// Open existing file
f, _ := sftp.Open("/etc/hostname")
// ... read from f

// Other methods: MkdirAll, Remove, Rename, ReadDir, Stat, etc.
```
</details>

<details>
<summary>Execute a Script (streaming from io.Reader)</summary>

```go
// Execute a script from any io.Reader (e.g. a string, a file handle, an HTTP body)
cmd, err := client.Script(ctx, strings.NewReader("echo hello"))
if err != nil {
	log.Fatal(err)
}
out, err := cmd.CombinedOutput()
```
</details>

<details>
<summary>Execute a Script File</summary>

```go
// Read a local script file into memory and execute it on the remote host
cmd, err := client.ScriptFile(ctx, "/path/to/local/script.sh")
if err != nil {
	log.Fatal(err)
}

// Override the interpreter for a PHP script:
cmd, err = client.ScriptFile(ctx, "/path/to/script.php", goph.WithPath("/usr/bin/php"))

// Override the shell path (default: /bin/sh)
cmd, err = client.ScriptFile(ctx, "/path/to/local/script.sh", goph.WithPath("/bin/zsh"))
```
</details>

<details>
<summary>Override Config Before Dial</summary>

```go
client, err := goph.New("root", "192.1.1.3",
	goph.WithPassword("pass"),
	goph.WithConfig(func(config *ssh.ClientConfig) error {
		config.Timeout = 30 * time.Second
		config.ClientVersion = "SSH-2.0-MyApp"
		return nil
	}),
)
```
</details>

<details>
<summary>Low-Level: Custom ssh.ClientConfig (Dial)</summary>

```go
c := &goph.Client{
	User: "root",
	Addr: "192.1.1.3",
	Port: 22,
}

config := &ssh.ClientConfig{
	Auth:            []ssh.AuthMethod{ssh.Password("pass")},
	HostKeyCallback: ssh.InsecureIgnoreHostKey(),
}

if err := goph.Dial(c, config); err != nil {
	log.Fatal(err)
}
defer c.Close()
```
</details>

<details>
<summary>Reusable Dialer (Multiple Connections)</summary>

```go
d := goph.NewDialer(
	goph.WithKeyFile("/home/user/.ssh/id_rsa", ""),
	goph.WithPort(2222),
)

client1, _ := d.New("root", "host1")
client2, _ := d.New("root", "host2", goph.WithPassword("fallback"))
```
</details>

### 🛠️&nbsp; Utils and Helpers

<details>
<summary>Check if Host is Known</summary>

```go
found, err := goph.CheckKnownHost("myhost", remoteAddr, publicKey, "")
if err != nil {
	// key mismatch! possible MITM attack!
	log.Fatal(err)
}
if !found {
	// host is new, ask user to trust it
}
```
</details>

<details>
<summary>Add Host to Known Hosts</summary>

```go
err := goph.AddKnownHost("myhost", remoteAddr, publicKey, "")
```
</details>

<details>
<summary>Parse Private Key File</summary>

```go
signer, err := goph.ParseKeyFile("/home/user/.ssh/id_rsa", "passphrase")
if err != nil {
	log.Fatal(err)
}

client, err := goph.New("root", "192.1.1.3",
	goph.WithSigner(signer),
)
```
</details>

<details>
<summary>Parse Raw Key (from bytes)</summary>

```go
keyBytes, _ := os.ReadFile("/home/user/.ssh/id_rsa")
signer, err := goph.ParseKey(keyBytes, "passphrase")
if err != nil {
	log.Fatal(err)
}

client, err := goph.New("root", "192.1.1.3",
	goph.WithSigner(signer),
)
```
</details>

<details>
<summary>Check SSH Agent Availability</summary>

```go
if goph.HasAgent() {
	fmt.Println("SSH agent is available at", os.Getenv("SSH_AUTH_SOCK"))
} else {
	fmt.Println("SSH agent not running or SSH_AUTH_SOCK not set")
}
```
</details>

<details>
<summary>Default Known Hosts Path and Ensure</summary>

```go
path, err := goph.DefaultKnownHostsPath()
if err != nil {
	log.Fatal(err)
}
fmt.Println("Known hosts file:", path)

// EnsureKnownHosts returns a callback and creates the file (and ~/.ssh/)
// if it does not exist yet — useful when writing first-time tools.
cb, err := goph.EnsureKnownHosts(filepath.Join(os.Getenv("HOME"), ".ssh", "my_hosts"))
if err != nil {
	log.Fatal(err)
}

client, err := goph.New("root", "192.1.1.3",
	goph.WithPassword("pass"),
	goph.WithHostKeyCallback(cb),
)
```
</details>

## 🛡️&nbsp; Security Notes

- **Known-hosts verification is enabled by default.** Do not use `goph.WithInsecureIgnoreHostKey()` in production; it makes you vulnerable to MITM attacks.
- **A missing `~/.ssh/known_hosts` file is created automatically** as an empty file, but unknown host keys are still rejected until you explicitly trust them.
- **Command arguments are not shell-escaped.** `Cmd.String()` returns raw `Path` and `Args`. Never pass untrusted input directly into commands; sanitize or use fixed argument lists.
- **Prompt the user before calling `goph.AddKnownHost`.** A mismatched key from `goph.CheckKnownHost` should be treated as a potential MITM attack.


## ❓&nbsp; FAQ

Expand each question to see the answer.

<details>
<summary>Why is there no <code>Dir</code> field like <code>os/exec.Cmd</code>?</summary>

SSH <code>exec</code> requests do not support setting a working directory in the protocol. The server runs the command in the user's default shell context (typically their home directory). To run a command in a specific directory, prefix it:

```go
out, err := client.Run("cd /var/log && ls -la")
```
</details>

<details>
<summary>How do I run <code>sudo</code> commands?</summary>

Either connect as root, or use <code>sudo -S</code> and feed the password through stdin. Be aware that <code>sudo</code> may require a TTY, which <code>goph</code> does not provide by default:

```go
cmd, err := client.Command("sudo", "-S", "systemctl", "restart", "nginx")
if err != nil {
	log.Fatal(err)
}

cmd.Stdin = strings.NewReader("your_sudo_password\n")
out, err := cmd.CombinedOutput()
```
</details>

<details>
<summary>Does goph work on Windows?</summary>

Yes. <code>goph</code> is pure Go and works on Windows, Linux, and macOS. For Windows-specific key providers (CNG/CAPI, Pageant, named-pipe OpenSSH agent), build or obtain an <code>ssh.Signer</code> and use <code>WithSigner</code>:

```go
signer, err := ssh.NewSignerFromKey(myPrivateKey)
client, err := goph.New("user", "host", goph.WithSigner(signer))
```
</details>


## 🤝&nbsp; Missing a Feature?

Feel free to open a new issue, or contact me.

## 📘&nbsp; License

Goph is provided under the [MIT License](https://github.com/melbahja/goph/blob/master/LICENSE).
