fixed: a library for fixed-point arithmetic
===========================================

[![Build Status](https://travis-ci.org/ReconfigureIO/fixed.svg?branch=master)](https://travis-ci.org/ReconfigureIO/fixed)
[![Documentation](https://godoc.org/github.com/ReconfigureIO/fixed?status.svg)](http://godoc.org/github.com/ReconfigureIO/fixed)

This is a fork of Go's [fixed point library][gofixed], optimized for FPGAs running on the Reconfigure.io platform.

It currently provides only Q26:6 and Q52:12 precision¹ types. If you need other precisions, open an issue or a pull request.

¹ See the Wikipedia page on the [Q number format][q] for information on this notation.

[q]: https://en.wikipedia.org/wiki/Q_(number_format)
[gofixed]: https://godoc.org/golang.org/x/image/math/fixed


Using in your kernels
---------------------

Reconfigure.io supports including vendor packages in your kernels. You can use your favorite Go dependency manager to add it to your kernel. We use [glide](https://github.com/Masterminds/glide) for our code.

```
$ glide create --non-interactive
[INFO]  Generating a YAML configuration file and guessing the dependencies
[INFO]  Attempting to import from other package managers (use --skip-import to skip)
[INFO]  Scanning code to look for dependencies
[INFO]  Writing configuration file (glide.yaml)
[INFO]  You can now edit the glide.yaml file. Consider:
[INFO]  --> Using versions and ranges. See https://glide.sh/docs/versions/
[INFO]  --> Adding additional metadata. See https://glide.sh/docs/glide.yaml/
[INFO]  --> Running the config-wizard command to improve the versions in your configuration
$ glide get github.com/ReconfigureIO/fixed
[INFO]  Preparing to install 1 package.
[INFO]  Attempting to get package github.com/ReconfigureIO/fixed
[INFO]  --> Gathering release information for github.com/ReconfigureIO/fixed
[INFO]  --> Adding github.com/ReconfigureIO/fixed to your configuration
[INFO]  Downloading dependencies. Please wait...
[INFO]  --> Fetching updates for github.com/ReconfigureIO/fixed
[INFO]  Resolving imports
[INFO]  Downloading dependencies. Please wait...
[INFO]  Exporting resolved dependencies...
[INFO]  --> Exporting github.com/ReconfigureIO/fixed
[INFO]  Replacing existing vendor dependencies
```

You should now see it in your `vendor` directory.

```
$ tree vendor
vendor
└── github.com
    └── ReconfigureIO
        └── fixed
            ├── examples
            │   └── mult
            │       ├── cmd
            │       │   └── test-mult
            │       │       └── main.go
            │       └── main.go
            ├── fixed.go
            ├── LICENSE
            ├── Makefile
            └── README.md

```

Contributing
------------

Pull requests & issues are enthusiastically accepted!

By participating in this project you agree to follow our [Code of Conduct](CODE_OF_CONDUCT.md).
