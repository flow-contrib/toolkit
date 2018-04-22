Toolkit
=======


## SSH

`flow.conf`

```hocon
packages = ["github.com/flow-contrib/toolkit/ssh"]

app {
    name = "ssh"
    usage = "This is a demo for run script on remote server"

    commands {
        run {
            usage = "Run command on remote ssh server"

            default-config = { 
            
                user = "user" 
                host = "localhost"
                port = "22"
                identity-file="/Users/gogap/.ssh/id_rsa"

                environment = ["GOPATH=/gopath"]
                command     = ["/bin/bash"]
                timeout     = 10s

                stdin ="""
                ping -c 1 example.com
                echo $GOPATH
                """

                output.name = "ping-example" # set output name
            }

            flow = ["toolkit.ssh.run"]
        }
    }
}
```

```bash
$ go-flow run --config flow.conf run
```

**output**

```json
[
    {
        "name": "ping-example",
        "value": {
            "host": "localhost",
            "port": "22",
            "user": "user",
            "command": {
                "environment": [
                    "GOPATH=/gopath"
                ],
                "command": [
                    "/bin/bash"
                ],
                "stdin": "ping -c 1 example.com\n                echo $GOPATH"
            },
            "output": "PING example.com (93.184.216.34): 56 data bytes\n64 bytes from 93.184.216.34: icmp_seq=0 ttl=46 time=264.645 ms\n--- example.com ping statistics ---\n1 packets transmitted, 1 packets received, 0% packet loss\nround-trip min/avg/max/stddev = 264.645/264.645/264.645/0.000 ms\n"
        }
    }
]
```

## Pwgen

`flow.conf`

```hocon
packages = ["github.com/flow-contrib/toolkit/pwgen"]

app {
    name = "pwgen"
    usage = "This is a demo for generate password"

    commands {
        generate {
            usage = "generate password"

            default-config = { 

                 gitlab {

                    # name to append output, if is empty, will use config key 'gitlab' as name
                    name = "GITLAB_PASSWORD"

                    len = 16
                    symbols = true
                    env = true # it will set env to GITLAB_PASSWORD_PLAIN and GITLAB_PASSWORD_ENCODED
                 }

                 mysql-prod {
                    len = 16

                    # encoding could be: sha256, sha512, base64
                    encoding = md5

                    # it will set env to MYSQL_PROD_PLAIN and MYSQL_PROD_ENCODED
                    env = true
                 }
            }

            flow = ["toolkit.pwgen.generate"]
        }
    }
}
```

```bash
$ go-flow -v run --config flow.conf generate
```

**output**

```json
[
    {
        "name": "GITLAB_PASSWORD",
        "value": {
            "name": "GITLAB_PASSWORD",
            "length": 16,
            "encoding": "plain",
            "plain": "N!c,Hqub7KXB1S(R",
            "encoded": "N!c,Hqub7KXB1S(R",
            "symbols": true,
            "environment": [
                "GITLAB_PASSWORD_PLAIN",
                "GITLAB_PASSWORD_ENCODED"
            ]
        },
        "tags": [
            "toolkit",
            "pwgen"
        ]
    },
    {
        "name": "mysql-prod",
        "value": {
            "name": "mysql-prod",
            "length": 16,
            "encoding": "md5",
            "plain": "2MbupEycLjUkpOyt",
            "encoded": "2eafad5fd0808c78956901de39cfbe74",
            "symbols": false,
            "environment": [
                "MYSQL_PROD_PLAIN",
                "MYSQL_PROD_ENCODED"
            ]
        },
        "tags": [
            "toolkit",
            "pwgen"
        ]
    }
]
```