Toolkit
=======


## SSH

`flow.conf`

```ssh
packages = ["github.com/flow-contrib/tookit/ssh"]

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

            flow = ["tookit.ssh.run"]
        }
    }
}
```

```bash
$ go-flow run --config flow.conf run
```

```
PING example.com (93.184.216.34): 56 data bytes
64 bytes from 93.184.216.34: icmp_seq=0 ttl=46 time=162.728 ms
......
```