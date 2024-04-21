# OpenSMTPD-framework

THIS IS A WIP, DO NOT USE UNLESS YOU KNOW WHAT YOU'RE DOING.


## cmd/table

`cmd/table` is a utility to help test table backends during development.

```
$ table -table foobar -service userinfo -lookup gilles \
    /usr/local/libexec/smtpd/table-passwd /etc/passwd
lookup-result|deadbeefabadf00d|found|441:441:/var/empty

$ table -table foobar -service userinfo -lookup gillou \
    /usr/local/libexec/smtpd/table-passwd /etc/passwd
lookup-result|deadbeefabadf00d|not-found
```
