# gomo

[![Build Status](https://travis-ci.com/frasercobb/gomo.svg?branch=master)](https://travis-ci.com/frasercobb/gomo)
[![License](https://img.shields.io/github/license/frasercobb/gomo)](/license)
[![Release](https://img.shields.io/github/v/release/frasercobb/gomo.svg)](https://github.com/frasercobb/gomo/releases/latest)

Interactive CLI to upgrade go module dependencies.

![Screenshot](example_output.png)

Currently a work in progress.

## Installation

Binaries for OS X and Linux are available on the [releases page](https://github.com/frasercobb/gomo/releases).

Alternatively, you can install using go:

```
go install github.com/frasercobb/gomo
```

## Usage

```
gomo
```

Output will be coloured by update type:
* Green indicates a patch update
* Blue indicates a minor update