# temple

Template rendering command line tool. The command line is your temple.

# installation

## linux

```bash
wget -qO- https://github.com/aybabtme/temple/releases/download/0.1/temple_linux.tar.gz | tar xvz
```

## darwin

```bash
wget -qO- https://github.com/aybabtme/temple/releases/download/0.1/temple_darwin.tar.gz | tar xvz
```

# usage

```bash
temple < cfg.tpl.json > /etc/service/cfg.json -var "thread_count=9000"
```

# contribution idea

* more template engines
* support to read variables from consul/etcd/zk/wtv
* support to render a whole filesystem tree at once

# license

MIT


