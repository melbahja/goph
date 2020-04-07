<div align="center">
	<h1>Golang SSH Client.</h1>
    <a href="https://github.com/melbahja/goph">
        <img src="https://github.com/melbahja/goph/raw/master/.github/goph.png" width="200">
    </a>
    <h4 align="center">
	   Fast and easy golang ssh client module.
	</h4>
</div>

<p align="center">
    <a href="#installation">Installation</a> ❘
    <a href="#features">Features</a> ❘
    <a href="#usage">Usage</a> ❘
    <a href="#examples">Examples</a> ❘
    <a href="#license">License</a>
</p>


## Installation

```bash
go get github.com/melbahja/goph
```

## Features

- Easy to use.
- Supports **known hosts** by default.
- Supports connections with **passwords**.
- Supports connections with **private keys**.
- Supports connections with **protected private keys** with passphrase.
- Supports **upload** files from local to remote.
- Supports **download** files from remote to local.

## Usage

Run a command via ssh:
```go
package main

import (
	"log"
	"fmt"
	"github.com/melbahja/goph"
)

func main() {

	// Start new ssh connection with private key.
	client, err := goph.New("root", "192.1.1.3", goph.Key("/home/mohamed/.ssh/id_rsa", ""))

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

##### Start connection with protected private key:
```go
client, err := goph.New("root", "192.1.1.3", goph.Key("/home/mohamed/.ssh/id_rsa", "you_passphrase_here"))
```

##### Start connection with password:
```go
client, err := goph.New("root", "192.1.1.3", goph.Password("you_password_here"))
```

##### Upload local file to remote:
```go
err := client.Upload("/path/to/local/file", "/path/to/remote/file")
```

##### Download remote file to local:
```go
err := client.Download("/path/to/remote/file", "/path/to/local/file")
```

##### Execute bash commands:
```go
out, err := client.Run("bash -c 'printenv'")
```

##### Execute bash command with env variables:
```go
out, err := client.Run(`env MYVAR="MY VALUE" bash -c 'echo $MYVAR;'`)
```

For more read the [go docs](https://pkg.go.dev/github.com/melbahja/goph).

## Examples

See [Examples](https://github.com/melbahja/ssh/blob/master/examples).

## License

Goph is provided under the [MIT License](https://github.com/melbahja/goph/blob/master/LICENSE).
