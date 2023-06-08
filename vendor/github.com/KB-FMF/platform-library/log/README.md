# For Developer `❤`

this document prepare for Developers, to make useful about program obfuscate


## Usage
### Set Go Module Off temporary and Install cmd program

```bash
$ go env -w  GO111MODULE=off
$ go get github.com/znly/strobfus 
```

### Dummy config 

Before use cmd `strobfus`, we named file program `config.go`. knowing about this file contact our Lead team 🫰

### Generator
```bash
$ strobfus -filename=log/config.go -seed="SECRETS"
```

it will be generated `config_gen.go`