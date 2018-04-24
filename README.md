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

                quiet = false
                output.name = "ping-example" # set output name
            }

            flow = ["toolkit.ssh.run"]
        }
    }
}
```

```bash
$ go-flow run --config flow.conf run

PING example.com (93.184.216.34): 56 data bytes
64 bytes from 93.184.216.34: icmp_seq=0 ttl=46 time=267.681 ms
--- example.com ping statistics ---
1 packets transmitted, 1 packets received, 0% packet loss
round-trip min/avg/max/stddev = 267.681/267.681/267.681/0.000 ms
/gopath
```

**output**

```json
[
    {
        "name": "ping-example",
        "value": {
            "host": "rijin-services-agent",
            "port": "20022",
            "user": "work",
            "command": {
                "environment": [
                    "GOPATH=/gopath"
                ],
                "command": [
                    "/bin/bash"
                ],
                "stdin": "ping -c 1 example.com\n                echo $GOPATH"
            },
            "output": "PING example.com (93.184.216.34): 56 data bytes\n64 bytes from 93.184.216.34: icmp_seq=0 ttl=46 time=267.681 ms\n--- example.com ping statistics ---\n1 packets transmitted, 1 packets received, 0% packet loss\nround-trip min/avg/max/stddev = 267.681/267.681/267.681/0.000 ms\n/gopath"
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
$ go-flow -v run --config flow.conf generate -o pwd-output.json

# Idempotent, if input the output's data before, 
# it will generate the same password, and will export to environment again
# go-flow -v run --config flow.conf generate -input pwd-output.json
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

## Readline

`flow.conf`

```hocon
packages = ["github.com/flow-contrib/toolkit/readline"]

app {
    name = "readline"
    usage = "This is a demo for readline"

    commands {
        read-text {
            usage = "read text"

            default-config = { 
                name = "gitlab-ci-url"
                prompt = "please input gitlab-ci url"
                confirm = true
                env = true ## env key = GITLAB_CI_URL
            }

            flow = ["toolkit.readline.text.read"]
        }

        read-password {
            usage = "read password"

            default-config = { 
                name = "gitlab-ci-token"
                prompt = "please input gitlab-ci token"
                confirm = false
                env = true ## env key = GITLAB_CI_TOKEN
            }

            flow = ["toolkit.readline.password.read"]
        }
    }
}
```

```
$ go-flow -v run --config flow.conf read-text

go-flow -v run --config flow.conf read-text
please input gitlab-ci url:https://gitlab.com
you are input 'https://gitlab.com', is it correct? (yes/no):yes
```

**output**

```
[
    {
        "name": "gitlab-ci-url",
        "value": {
            "name": "gitlab-ci-url",
            "input": "https://gitlab.com",
            "type": "text"
        },
        "tags": [
            "toolkit",
            "readline"
        ]
    }
]
```
