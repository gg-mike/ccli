package ssh

import "golang.org/x/crypto/ssh"

func NewConnection(user, addr string, privateKey string) (*ssh.Client, error) {
	cfg, err := NewConfig(user, privateKey)
	if err != nil {
		return nil, err
	}
	return ssh.Dial("tcp", addr, &cfg)
}

func CheckConnection(user, addr string, privateKey string) error {
	conn, err := NewConnection(user, addr, privateKey)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}
