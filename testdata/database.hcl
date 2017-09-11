database {
    provider = "postgres"
    secretDSN = "user=produ password=s3cret dbname=proddb host=prodhost"
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
