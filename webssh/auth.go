package webssh

import "golang.org/x/crypto/ssh"

func AuthPassword(password string) ssh.AuthMethod {
	return ssh.Password(password)
}

func AuthKeyboardInteractive(fn ssh.KeyboardInteractiveChallenge) ssh.AuthMethod {
	return ssh.KeyboardInteractive(fn)
}

func AuthPrivateKey(key []byte) ssh.AuthMethod {
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(signer)
}

func AuthPrivateKeyWithPassphrase(key []byte, passphrase string) ssh.AuthMethod {
	signer, err := ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(signer)
}
