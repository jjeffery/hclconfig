# hclconfig:  Configuration file support for cloud-based servers
[![GoDoc](https://godoc.org/github.com/jjeffery/hclconfig?status.svg)](https://godoc.org/github.com/jjeffery/hclconfig)
[![License](http://img.shields.io/badge/license-MIT-green.svg?style=flat)](https://raw.githubusercontent.com/jjeffery/hclconfig/master/LICENSE.md)
[![GoReportCard](https://goreportcard.com/badge/github.com/jjeffery/hclconfig)](https://goreportcard.com/report/github.com/jjeffery/hclconfig)

Package hclconfig is designed to reduce the effort required acquire a 
configuration file from a cloud-based server program.

The main features this package provides are:

* Download configuration from a HTTP/HTTPS URL, or an S3 URL or a local file.
* Provides encryption at rest for sensitive information in the configuration file.
* Detect changes in the configuration file.

For more information, refer to the [Godoc](https://godoc.org/github.com/jjeffery/hclconfig) documentation.
