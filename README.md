# temple

Template rendering command line tool. The command line is your temple.

```bash
$ echo '{"hello": "{{.who}}"}' | temple -var who=world
{"hello", "world"}
```

## installation

### linux

```bash
wget -qO- https://github.com/aybabtme/temple/releases/download/0.2/temple_linux.tar.gz | tar xvz
```

### darwin

```bash
wget -qO- https://github.com/aybabtme/temple/releases/download/0.2/temple_darwin.tar.gz | tar xvz
```

## contribution ideas

* more template engines
* support to read variables from consul/etcd/zk/wtv
* support to verify that all template variables have a value

## license

MIT
