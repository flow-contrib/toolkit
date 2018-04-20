package ssh

type Config struct {
	User           string
	Password       string
	Host           string
	Port           string
	IdentityFile   string
	ConnectRetries int
}
