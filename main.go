// secure is a super simple TLS termination proxy
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var (
	addr       string
	certFile   string
	configFile string
	keyFile    string
	upstream   string
	version    bool
)

func init() {
	flag.StringVar(&addr, "addr", ":443", "host and port to listen on")
	flag.StringVar(&certFile, "cert", "", "path to cert file")
	flag.StringVar(&configFile, "config", "", `path to config file (default: "$HOME/secure.conf")`)
	flag.StringVar(&keyFile, "key", "", "path to key file")
	flag.BoolVar(&version, "version", false, "print version string and exit")

	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), shortUsage())
		flag.PrintDefaults()
	}
}

func shortUsage() string {
	return fmt.Sprintf("usage: %s -addr [host]:port -cert certfile -config conf -key keyfile [upstream]",
		filepath.Base(os.Args[0]))
}

func homeDir() (home string) {
	if h := os.Getenv("HOME"); h != "" {
		home = h
	} else if h := os.Getenv("USERPROFILE"); h != "" {
		home = h
	}
	return
}

func findConfigFile() (configFile string) {
	if home := homeDir(); home != "" {
		p := filepath.Join(home, DefaultConfigFileName)
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.IsDir() {
				configFile = p
			}
		}
	}
	return
}

func _main() error {
	flag.Parse()

	if version {
		fmt.Fprintln(flag.CommandLine.Output(), Version)
		os.Exit(0)
	}

	if flag.NArg() > 1 {
		flag.Usage()
		os.Exit(2)
	} else if flag.NArg() == 1 {
		upstream = flag.Arg(0)
	}

	if configFile == "" {
		configFile = findConfigFile()
	}

	if configFile != "" {
		config, err := ParseConfigFile(configFile)
		if err != nil {
			return fmt.Errorf("error parsing config file: %v", err)
		}
		fmt.Println("using config file: " + configFile)

		if addr == "" {
			if ad, ok := config["ListenAddr"]; ok {
				addr = ad
			}
		}

		if certFile == "" {
			if cf, ok := config["CertFile"]; ok {
				certFile = cf
			}
		}

		if keyFile == "" {
			if kf, ok := config["KeyFile"]; ok {
				keyFile = kf
			}
		}

		if upstream == "" {
			if u, ok := config["UpstreamURL"]; ok {
				upstream = u
			}
		}
	}

	if upstream == "" {
		fmt.Fprintln(flag.CommandLine.Output(), "no upstream url provided!")
		fmt.Fprintln(flag.CommandLine.Output(), shortUsage())
		os.Exit(2)
	}

	u, err := url.Parse(upstream)
	if err != nil {
		return fmt.Errorf("invalid upstream url: %v", err)
	}

	rp := httputil.NewSingleHostReverseProxy(u)
	srv := http.Server{
		Handler: rp,
		Addr:    addr,
	}

	done := make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		fmt.Println(<-sig)

		if err := srv.Shutdown(context.Background()); err != nil {
			fmt.Printf("Shutdown: %v", err)
		}
		close(done)
	}()

	fmt.Printf("addr=%s cert=%s key=%s upstream=%s\n", srv.Addr, certFile, keyFile, u.String())
	if err := srv.ListenAndServeTLS(certFile, keyFile); err != http.ErrServerClosed {
		return fmt.Errorf("ListenAndServeTLS: %v", err)
	}

	<-done
	return nil
}

func main() {
	err := _main()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
