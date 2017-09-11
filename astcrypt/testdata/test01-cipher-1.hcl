// test 01

database "oltp" {
    host = "oltp.corporation.net"
    db = "oltp_prod"
    user = "oltp_user"

    password {
        ciphertext = "ciphertext('oltp_password')"
    }
}

database "mis" {
    host = "mis.corporation.net"
    db = "mis_prod"
    user = "mis_user"
    password {
        ciphertext = <<END
            ciphertext('mis_password')
        END
    }
}

cors {
    secret {
        ciphertext = <<END
            ciphertext('fried eggs and ham')
        END
    }
}
