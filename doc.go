/*
Package hclconfig aims to make it easy for cloud-based server programs
to access configuration files.

Configuration files can be referenced as a URL or as a local file.
Supported URL schemes include http, https, s3 and file. The package
provides a mechanism to efficiently determine whether a config file
has changed.

It is a common requirement for a configuration file to include
sensitive information such as passwords, database connection strings,
API keys, and similar. It is poor practice to store this configuration
in clear text. This package provides a convenient mechanism for storing
sensitive information in a configuration file in encrypted form.

The package is somewhat opinionated. Configuration files are expected
to be in HCL format (https://github.com/hashicorp/hcl). Sensitive
data is encrypted using a data key, which is stored in the configuration
file in encrypted form. Currently AWS KMS is used to encrypt the data
key, but other mechanisms could be included in future versions of this
package.

The following example shows an HCL configuration file that stores sensitive
information.

 // example configuration file
 encryption {
	 // data key encrypted using AWS KMS
	 kms = <<EOF
		AQIDAHgLhsBflVB0KoR1VWanrwNzS+ylS6x/KfXjXLqRJA+I1AGdDZQVyAda6rR1A9A9qT7GAA
		AAfjB8BgkqhkiG9w0BBwagbzBtAgEAMGgGCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQMU28Y
		xot8ipSiVrmZAgEQgDvoJNL7unAdqIgQze98nfCBH0tF3+fbJOeZwjdvI4Od4Loentci39Zjrk
		otk6cofeipCC8UteWQ7lh2Pw==
		EOF
 }

 // storing sensitive information
 database {
	 username = "scott"
	 hostname = "db.example.com"
	 dbname = "production_db"
	 password = {
		 ciphertext = <<EOF
			CDjOKgnBIunYEfUru+jD7OgGmF9+nF3Y
			XsLaJWDe5nIjAYfGdrStPVVYJJdGao0N
			3VFf4bCUFJE=
			EOF
	 }
 }

In this file the `encryption` section has specified a data encryption key that is
encrypted using an AWS KMS encryption key. The sensitive data in the configuration
file (eg the database password) is encrypted using the data encryption key.

A command line utility `hclconfig` is provided to assist with encrypting and decrypting
data in a configuration file. See package "cmd/hclconfig" for details.
*/
package hclconfig
