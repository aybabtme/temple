# temple

Template rendering command line tool. The command line is your temple.

```bash
$ echo '{"hello": "{{.who}}"}' | temple -var who=world
{"hello", "world"}
```

## installation

### Debian/Ubuntu

```bash
wget https://github.com/aybabtme/temple/releases/download/vundefined/temple_undefined_linux_amd64.deb
dpkg -i temple_undefined_linux_amd64.deb
```

### darwin

```bash
brew install aybabtme/homebrew-tap/bitflip
```

## contribution ideas

* more template engines
* support to read variables from consul/etcd/zk/wtv
* support to verify that all template variables have a value

## license

MIT
