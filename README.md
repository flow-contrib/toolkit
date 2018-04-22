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
            usage = "This command will print hello"

            default-config = { 
            
                user = "user" 
                host = "localhost"
                port = "22"
                identity-file="/Users/gogap/.ssh/id_rsa"

                environment = ["GOPATH=/gopath"]
                command     = ["/bin/bash"]
                timeout     = 100s

                stdin ="""
                ping -c 10 example.com
                echo $GOPATH
                """
            }

            flow = ["toolkit.ssh.run"]
        }
    }
}
```

```bash
$ go-flow run --config flow.conf run

PING example.com (93.184.216.34): 56 data bytes
64 bytes from 93.184.216.34: icmp_seq=0 ttl=46 time=162.728 ms
......
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
{
    "output": [
        {
            "name": "GITLAB-PASSWORD",
            "value": {
                "name": "GITLAB-PASSWORD",
                "length": 16,
                "encoding": "plain",
                "plain": "uZIP1}*T^vbhY_Nz",
                "value": "uZIP1}*T^vbhY_Nz",
                "symbols": true,
                "env": ""
            }
        },
        {
            "name": "mysql-prod",
            "value": {
                "name": "mysql-prod",
                "length": 16,
                "encoding": "md5",
                "plain": "uAxK70hqMvR6vvXb",
                "value": "1965c122cc388357ab76822cad593a32",
                "symbols": false,
                "env": "MYSQL_PROD_PASSWORD"
            }
        }
    ]
}
```