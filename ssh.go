package main

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func makeSigner(keyname string) (signer ssh.Signer, err error) {
	fp, err := os.Open(keyname)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer fp.Close()

	buf, err := ioutil.ReadAll(fp)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	signer, _ = ssh.ParsePrivateKey(buf)

	return signer, nil
}

func makeKeyring() ssh.AuthMethod {
	signers := []ssh.Signer{}
	keys := []string{os.Getenv("HOME") + "/.ssh/id_rsa"}

	for _, keyname := range keys {
		signer, err := makeSigner(keyname)
		if err == nil {
			signers = append(signers, signer)
		}
	}

	return ssh.PublicKeys(signers...)
}

func main() {
	cmd := os.Args[1]
	hosts := os.Args[2:]

	results := make(chan string, 10)
	timeout := time.After(10 * time.Second)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "22"
	}

	config := &ssh.ClientConfig{
		User: os.Getenv("USER"),
		Auth: []ssh.AuthMethod{makeKeyring()},
	}
	config.SetDefaults()

	for _, hostname := range hosts {
		go func(hostname string) {
			conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", hostname, port), config)
			if err != nil {
				log.Fatalf("unable to connect: %s", err)
			}

			session, err := conn.NewSession()
			if err != nil {
				fmt.Println(err)
			}
			defer session.Close()

			var stdoutBuf bytes.Buffer
			session.Stdout = &stdoutBuf
			session.Run(cmd)

			results <- fmt.Sprintf("%s -> %s", hostname, stdoutBuf.String())
		}(hostname)
	}

	for i := 0; i < len(hosts); i++ {
		select {
		case res := <-results:
			fmt.Print(res)
		case <-timeout:
			fmt.Println("timeout")
			return
		}
	}
}