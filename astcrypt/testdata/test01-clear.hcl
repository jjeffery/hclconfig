// test 01

database "oltp" {
    host = "oltp.corporation.net"
    db = "oltp_prod"
    user = "oltp_user"
    password = "oltp_password"
}

database "mis" {
    host = "mis.corporation.net"
    db = "mis_prod"
    user = "mis_user"
    password = "mis_password"
}

cors {
    secret = "fried eggs and ham"
}
