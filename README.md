# hclconfig:  Configuration file support for cloud-based servers
[![GoDoc](https://godoc.org/github.com/jjeffery/hclconfig?status.svg)](https://godoc.org/github.com/jjeffery/hclconfig)
[![License](http://img.shields.io/badge/license-MIT-green.svg?style=flat)](https://raw.githubusercontent.com/jjeffery/hclconfig/master/LICENSE.md)
[![GoReportCard](https://goreportcard.com/badge/github.com/jjeffery/hclconfig)](https://goreportcard.com/report/github.com/jjeffery/hclconfig)

Package hclconfig is designed to reduce the effort required to access
a configuration file.

The main features this package provides are:

* Download configuration via HTTP/HTTPS, from an S3 bucket or from a local file
* Detect if the configuration file has changed since it was downloaded
* Provide encryption at rest for confidential information in the configuration file

This package is designed to work with configuration files that are in 
[HCL](https://github.com/hashicorp/hcl) format. The reason for this choice
is that it is straightforward to parse an HCL file into an 
[AST](https://en.wikipedia.org/wiki/Abstract_syntax_tree), which makes it
possible to implement a convenient mechanism for encrypting and decrypting
confidential information.

## Simple Example
```go
// eg "https://config.my-app.net/my-app-config.hcl"
// eg "s3://config-bucket/my-app-config.hcl"
// eg "/etc/my-ap-config.hcl"
location := os.Getenv("CONFIG")

// download the config file, and decrypt any confidential information
file, err := hclconfig.Get(location)
exitIfError(err)

var db struct {
    Database struct {
        Provider  string
        SecretDSN string
    }
}

// decode the information we are after into db
err = file.Decode(&db)
exitIfError(err)

db, err := sql.Open(db.Database.Provider, db.Database.SecretDSN)
exitIfError(err)

// simple example of a goroutine that will initiate graceful shutdown
// if it detects a change in the configuration file
go func() {
    for {
        time.Sleep(time.Minute)
        changed, err := file.HasChanged()
        handleErr(err)
        if changed {
            initiateGracefulShutdown()
        }
    }
}()
```

## Encryption

Confidential information is encrypted using AES-256 CBC + HMAC-SHA256.

The 256-bit data encryption key is stored as a ciphertext blob in the
configuration file. The data encryption key is encrypted using 
[AWS KMS](https://aws.amazon.com/kms/). Other encryption providers could 
be implemented in a future version of this package.

Example of an unencrypted configuration file
```hcl
database {
    provider = "postgres"
    secretDSN = "user=produ password=s3cret dbname=proddb host=prodhost"
}
```

Example of an encrypted configuration file
```hcl
database {
    provider = "postgres"

    secretDSN {
        ciphertext = <<END
            RPbAjbNg/2iRsbifmJ3cp4vP8DSM2k6jp7JIFvji3oWjWe50rO5bHFOhMTNfVpTA4T4CBdxJ08
            1AkQOtOFLnj5F1YUzYFbqDW3j3wAvvDgT1lynt5F+DPT/CLQPC0llNKlMbAUAmliChGESdOL4f
            Dw==
            END
    }
}

// data encryption key
encryption {
    // alias/master-kms-key
    kms = <<END
        AQIDAHgLhsBflVB0KoR1VWanrwNzS+ylS6x/KfXjXLqRJA+I1AHRE6ev8Jq+7FsFvelMxsGLAAAAfj
        B8BgkqhkiG9w0BBwagbzBtAgEAMGgGCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQMhf8Dkptf+b8i
        VKEpAgEQgDusdz5gglVC/aF+15h8majTR8UrdFt3kniu4XHem6NJn4FZCrqVGock5Zd7H96njJgPrJ
        7jhtM7X/st3g==
        END
}
```

A command line utility is included in `./cmd/hclconfig`, which makes it easy to
encrypt confidential information.
```
$ hclconfig --help
manage secrets in HCL config files

Usage:
  hclconfig [flags]
  hclconfig [command]

Available Commands:
  encrypt     encrypt secrets in HCL file
  decrypt     decrypt secrets in HCL file
  generate    generate data key for use in HCL config file

Use "hclconfig [command] --help" for more information about a command.
```

For more information, refer to the [Godoc](https://godoc.org/github.com/jjeffery/hclconfig) documentation.
