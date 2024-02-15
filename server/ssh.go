package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/gliderlabs/ssh"
	"gitlab.com/proctorexam/go/env"
	"golang.org/x/crypto/bcrypt"
)

const (
	maxTimeout  = 12 * time.Hour
	idleTimeout = 1 * time.Minute
)

var vmAddr = env.Fetch("FLY_PRIVATE_IP", "localhost")

func serveSSH() error {
	ss := sshServer()
	exitCallbacks["ssh"] = func() { log.Fatal(ss.Shutdown(context.Background())) }
	return ss.ListenAndServe()
}

func sshServer() *ssh.Server {
	forwardHandler := &ssh.ForwardedTCPHandler{}

	return &ssh.Server{
		Addr: ":2222",
		PasswordHandler: func(ctx ssh.Context, pass string) bool {
			user := ctx.User()

			log.Println("PasswordHandler", ctx, user)

			if user == "" || user == "root" || pass == "" {
				return false
			}

			if userExists(user) {
				return false
			}

			hash := getPassword(user)
			if hash == "" {
				return false
			}

			return checkPasswordHash(pass, hash)
		},
		PtyCallback: func(ctx ssh.Context, pty ssh.Pty) bool {
			log.Println("PtyCallback", ctx, pty)
			return false
		},
		SessionRequestCallback: func(sess ssh.Session, requestType string) bool {
			log.Println("SessionRequestCallback", sess, requestType)
			return false
		},
		LocalPortForwardingCallback: func(ctx ssh.Context, destinationHost string, destinationPort uint32) bool {
			log.Println("LocalPortForwardingCallback", ctx, destinationHost, destinationPort)
			return false
		},
		ReversePortForwardingCallback: func(ctx ssh.Context, bindHost string, bindPort uint32) bool {
			log.Println("ReversePortForwardingCallback", ctx, bindHost, bindPort)
			user := ctx.User()

			if err := setProxyAddress(user, fmt.Sprintf("[%s]:%v", vmAddr, bindPort)); err != nil {
				log.Println(err)
				return false
			}

			removeCallback := func() {
				if err := removeProxyAddress(user); err != nil {
					log.Println(err)
				}
				delete(exitCallbacks, user)
			}

			exitCallbacks[user] = removeCallback

			go func() {
				<-ctx.Done()
				removeCallback()
			}()

			return true
		},

		ConnectionFailedCallback: func(conn net.Conn, err error) {
			log.Println("ConnectionFailedCallback", conn, err)
		},

		RequestHandlers: map[string]ssh.RequestHandler{
			"tcpip-forward":        forwardHandler.HandleSSHRequest,
			"cancel-tcpip-forward": forwardHandler.HandleSSHRequest,
		},

		MaxTimeout: maxTimeout,

		IdleTimeout: idleTimeout,
	}
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
