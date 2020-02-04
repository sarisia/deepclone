# Deepclone

[![ci](https://github.com/sarisia/deepclone/workflows/ci/badge.svg)](https://github.com/sarisia/deepclone/actions)
[![release](https://github.com/sarisia/deepclone/workflows/release/badge.svg)](https://github.com/sarisia/deepclone/releases)

## Build

```
go get github.com/sarisia/deepclone/cmd/deepclone
```

## Run

* Download from [Releases](https://github.com/sarisia/deepclone/releases)
or build yourself

```
./deepclone --depth 2 https://www.apple.co.jp
```

## Options

| Option | Required | Default | Usage |
|-|-|-|-|
| depth | false | 1 | Set fetch depth |
| conn | false | 10 | Set max concurrent connections |
