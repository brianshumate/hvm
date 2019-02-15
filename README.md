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

## Build

The simplest way to get going with an established Go environment is:

```
$ go get -u github.com:brianshumate/hvm
```

This will also pull down all of the dependent packages and build `hvm` into `$GOPATH/bin` so it'll be ready to use.

## Who?

hvm was created by [Brian Shumate](https://github.com/brianshumate) and made possible through the generous time of the good people named in [CONTRIBUTORS.md](https://github.com/brianshumate/hvm/blob/master/CONTRIBUTORS.md).
