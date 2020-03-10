# Simple Golang SSH Client

Golang module to execute commands over SSH connection


## Installation

```bash
go get github.com/melbahja/ssh
```

## Usage

```go
package main

import (
	"log"
	"fmt"
	"github.com/melbahja/ssh"
)

func main() {

	// Start ssh connection
	client, err := ssh.New(ssh.Config{
		User: "root",
		Addr: "192.168.122.163",
		Auth: ssh.Key("/home/mohamed/.ssh/id_rsa", ""),
	})

	// Execute a command
	out, err := client.Exec("ls /tmp/")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(out)
}

```

### Run Custom Command

```go
cmd := client.Command("ls", []string{"-alh"})

result, err := cmd.Run()

if err != nil {
	log.Fatal(err)
}

fmt.Println(result.Stdout.String())

```

### SSH Connection With Passphrase

```go
client, err := ssh.New(ssh.Config{
	User: "root",
	Addr: "192.168.122.163",
	Auth: ssh.Key("/home/mohamed/.ssh/id_rsa", "123456_your_passphrase_here"),
})
```

### SSH Connection With Password (Unsafe!)

```go
client, err := ssh.New(ssh.Config{
	User: "root",
	Addr: "192.168.122.163",
	Auth: ssh.Password("123456_your_password_here"),
})
```

### Add ClientConfig

```go
// import sh "golang.org/x/crypto/ssh"

client, err := ssh.New(ssh.Config{
	User: "root",
	Addr: "192.168.122.163",
	Port: 2000, // You can change the default 22 port
	Auth: ssh.Key("/path/to/privateKey", ""),
	Config: &sh.ClientConfig{
		Timeout: 10 * time.Second,
		// options here: https://pkg.go.dev/golang.org/x/crypto/ssh?tab=doc#ClientConfig 
	},
})
```

### Examples

See [Examples](https://github.com/melbahja/ssh/blob/master/examples)


### Contributing
Welcome!

## License

[MIT](https://github.com/melbahja/ssh/blob/master/LICENSE) © [Mohammed El Bahja](https://git.io/mohamed)
