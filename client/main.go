package main

import (
	"flag"
	"log"
	"mime"
	"os"
	"os/signal"
	"path"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"golang.org/x/crypto/ssh"
)

const (
	keepaliveInterval = time.Second * 30
	keepaliveRequest  = "keepalive@file-dance"
)

func main() {
	user := flag.String("user", "", "user")
	password := flag.String("password", "", "password")
	httpDir := flag.String("dir", "", "dir")
	flag.Parse()

	if *user == "" || *password == "" {
		log.Fatal("credentials missing")
	}

	// var hostKey ssh.PublicKey
	config := &ssh.ClientConfig{
		User: *user,
		Auth: []ssh.AuthMethod{
			ssh.Password(*password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Dial your ssh server.
	conn, err := ssh.Dial("tcp", *user+".file.dance:2222", config)
	if err != nil {
		log.Fatal("unable to connect: ", err)
	}
	defer conn.Close()

	// log.Println("remote", conn.RemoteAddr(), "local", conn.LocalAddr())

	_, port, _ := strings.Cut(conn.LocalAddr().String(), ":")

	l, err := conn.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatal("unable to register tcp forward: ", err)
	}
	defer l.Close()

	doneCh := make(chan os.Signal, 1)
	signal.Notify(doneCh, os.Interrupt)

	go func() {
		// Keep-Alive Ticker
		ticker := time.Tick(keepaliveInterval)
		for {
			select {
			case <-ticker:
				_, _, err := conn.SendRequest(keepaliveRequest, false, nil)
				if err != nil {
					// Connection is gone
					log.Printf("[%s] Keepalive failed, closing conn: %s\n", conn.RemoteAddr(), err)
					conn.Close()
					return
				}

			case <-doneCh:
				conn.Close()
				return
			}
		}
	}()

	// Serve HTTP with your SSH server acting as a reverse proxy.
	app := fiber.New(fiber.Config{
		GETOnly:                 true,
		EnableTrustedProxyCheck: true,
		TrustedProxies:          []string{*user + ".file.dance"},
		StreamRequestBody:       true,
	})

	app.Use(etag.New())

	app.Use(func(c *fiber.Ctx) error {
		if !(c.Is("json")) {
			return c.Next()
		}

		rPath := strings.TrimSuffix(c.Path(), "/")
		name := path.Join(*httpDir, rPath)
		info, err := os.Stat(name)

		if err != nil {
			if os.IsNotExist(err) {
				return c.SendStatus(fiber.StatusNotFound)
			} else {
				log.Println("os.Stat", err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		if !info.IsDir() {
			return c.Next()
		}

		ls, err := os.ReadDir(name)
		if err != nil {
			log.Println("os.ReadDir", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		es := make([]fiber.Map, 0, len(ls))
		for _, e := range ls {
			ename := e.Name()
			xname := path.Ext(ename)
			if xname == ".gz" {
				continue
			}
			es = append(es, fiber.Map{
				"name": ename,
				"path": path.Join(rPath, ename),
				"dir":  e.IsDir(),
				"mime": mime.TypeByExtension(xname),
			})
		}

		return c.JSON(es)
	})

	app.Static("/", *httpDir, fiber.Static{
		Compress:  true,
		ByteRange: true,
		Browse:    true,
	})

	log.Fatal(app.Listener(l))
}
