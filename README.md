# Deepclone

[![test](https://github.com/sarisia/deepclone/workflows/test/badge.svg)](https://github.com/sarisia/deepclone/actions)
[![build](https://github.com/sarisia/deepclone/workflows/build/badge.svg)](https://github.com/sarisia/deepclone/actions)
[![release](https://github.com/sarisia/deepclone/workflows/release/badge.svg)](https://github.com/sarisia/deepclone/releases)

[DevNote here](https://www.notion.so/sarisia/Deepclone-Devnote-773cf2f403914b9d83910b40a533ba0d)

## Run

* Download from [Releases](https://github.com/sarisia/deepclone/releases)
or build yourself

```
./deepclone --dir content --depth 2 https://www.apple.co.jp
```

## Build

```
go get github.com/sarisia/deepclone/cmd/deepclone
```

## Options

| Option | Required | Default | Usage |
|-|-|-|-|
| depth | | 1 | Set fetch depth |
| conn | | 16 | Set max concurrent connections |
| dir | | "content" | Set directory to save contents |
| debug | | false | Enable pprof debug endpoint |
