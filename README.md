# hvm

## What?

This is `hvm` the Hashi Version Manager. It is not an official HashiCorp project, but a personal one that was originally created to satisfy my need to work with multiple versions of the popular HashiCorp command line tools.

You can use `hvm` to install any known version of the following tools:

- consul
- nomad
- packer
- terraform
- vagrant
- vault

## How?

### Usage

Execute `hvm` alone and you'll encounter some handy usage notes. Use `hvm subcommand --help` at any time to get more details on a given subcommand:

```
Usage:
  hvm [command]

Available Commands:
  help        Help about any command
  info        Host information and current versions
  install     Install a supported binary at the latest available or specified version
  list        List available binary versions
  uninstall   Uninstall a binary
  use         Use a specific binary version
  version     Print hvm version

Flags:
      --config string   config file (default is $HOME/.hvm.yaml)
  -h, --help            help for hvm

Use "hvm [command] --help" for more information about a command.
```

#### info

#### list

#### install

Installation of binaries includes a live download phase which is internally handled by [go-getter](https://github.com/hashicorp/go-getter).

This provides the advantage that the SHA 256 summary is also compared between the Zip archive and what is posted on [releases.hashicorp.com](https://releases.hashicorp.com/) website for the binary in question, and download of the Zip archive occurs only if there is a match.

#### use

## Build

The simplest way to get going with an established Go environment is:

```
$ go get -u github.com:brianshumate/hvm
```

This will also pull down all of the dependent packages and build `hvm` into `$GOPATH/bin` so it'll be ready to use.

## Who?

hvm was created by [Brian Shumate](https://github.com/brianshumate) and made possible through the generous time of the good people named in [CONTRIBUTORS.md](https://github.com/brianshumate/hvm/blob/master/CONTRIBUTORS.md).
