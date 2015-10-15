# `./temple`

Template rendering command line tool. The command line is your temple.

## Usage

Render one file at a time, or a whole tree at once!

### file rendering

With a single file:

```bash
$ echo '{"hello": "{{.who}}"}' | temple file -var who=world
{"hello", "world"}
```
### tree rendering

A whole file tree at once! All the file in `/tmp/etc` are rendered
and the result is written onto `/etc/`.

```bash
$ temple tree -src /tmp/etc \
              -dst /etc/ \
              -var proxy_addr=192.168.0.1
```

## Installation

### linux

```bash
wget -qO- https://github.com/aybabtme/temple/releases/download/0.2/temple_linux.tar.gz | tar xvz
mv temple /opt/bin
```

### darwin

```bash
wget -qO- https://github.com/aybabtme/temple/releases/download/0.2/temple_darwin.tar.gz | tar xvz
mv temple /opt/bin
```

## Contribution Ideas

* more template engines
* support to read variables from consul/etcd/zk/wtv
* support to verify that all template variables have a value

## License

MIT
