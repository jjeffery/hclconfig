package hclconfig

import "log"

func ExampleGet() {
	// get a config file from a HTTP URL
	file, err := Get("https://example.com/config/file.hcl")
	if err != nil {
		log.Fatal(err)
	}
	doSomethingWith(file)

	// get a config file from an S3 URL
	file, err = Get("s3://bucket-name/config/file.hcl")
	if err != nil {
		log.Fatal(err)
	}
	doSomethingWith(file)

	// get a config file from the local filesystem
	file, err = Get("./config/file.hcl")
	if err != nil {
		log.Fatal(err)
	}
	doSomethingWith(file)
}

func ExampleFile_Decode() {
	file, err := Get("./config.hcl")
	if err != nil {
		log.Fatal(err)
	}

	var config struct {
		Database struct {
			Provider  string
			SecretDSN string
		}
	}

	if err := file.Decode(&config); err != nil {
		log.Fatal(err)
	}

	doSomethingWith(config)

	/* Config file would look something like:

	encryption {
		kms = "<cipher-text-blob>"
	}

	database {
		provider = "postgresql"
		secretDSN {
			ciphertext = "<cipher-text-blob>"
		}
	}

	*/
}

func doSomethingWith(v interface{}) {}
