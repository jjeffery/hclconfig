// This is an example of what an encrypted config file could look like.

encryption {
  // arn:aws:kms:ap-southeast-2:464878861558:key/bd35e69f-66b7-47c3-84d1-438ae73f59dc
  // alias/sp-master
  kms = <<END
        AQIDAHgLhsBflVB0KoR1VWanrwNzS+ylS6x/KfXjXLqRJA+I1AGmYHDAvjmEKx8+
        FdJ9NefjAAAAfjB8BgkqhkiG9w0BBwagbzBtAgEAMGgGCSqGSIb3DQEHATAeBglg
        hkgBZQMEAS4wEQQMdYSDxuS3ZL0WbCNcAgEQgDs5Yr0IkdOih89eCWti1qczplu6
        E6kaJFKvo64uTLQDdqieTkI9DMInkzGSKbym0Ii+W1vzztO69vx7pg==
        END
}

oauth2 "google" {
  clientId = "686122639682-3hn5fic7vqc9dt4h5nn6bsp29c7jc9be.apps.googleusercontent.com"
  clientSecret {
    ciphertext = <<EOF
      QHo/jx0L8H2+yonJxVXvDh4jtGP2qyh+VEMA+IFGMGybFKlQ9la1lwKEPDW9oZQ1Nu
      CwuentPVRaHq4NP2koq1seY4iIdnmOPsYKaBqtb+U=
      EOF
  }
}