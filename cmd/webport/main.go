package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/OutOfBedlam/webterm/webport"
)

func main() {
	serveSet := flag.NewFlagSet("serve", flag.ExitOnError)
	//	agentSet := flag.NewFlagSet("agent", flag.ExitOnError)
	flag.Usage = usage
	flag.Parse()
	switch flag.Arg(0) {
	case "serve":
		serveCommand(serveSet, flag.Args()[1:])
	// case "agent":
	// 	agentCommand(agentSet, flag.Args()[1:])
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage: webport <command> [options]")
	fmt.Println("Commands:")
	fmt.Println("  serve    Start the webport server")
	//	fmt.Println("  agent    Start the webport agent")
	fmt.Println("Use 'webport <command> -h' for more information about a command.")
}

type PortVar []string

var _ flag.Value = (*PortVar)(nil)

func (p *PortVar) String() string {
	return fmt.Sprint(*p)
}

func (p *PortVar) Set(value string) error {
	*p = append(*p, value)
	return nil
}

func serveCommand(fs *flag.FlagSet, args []string) {
	// flag -L [local_ip:]local_port:remote_ip:remote_port if local_ip is omitted, bind to all interfaces
	var portFlags PortVar
	fs.Var(&portFlags, "L", "Local port forwarding, [local_ip:]local_port:remote_ip:remote_port")
	if err := fs.Parse(args); err != nil {
		panic(err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	var ports []*webport.WebPort
	for _, p := range portFlags {
		addrBind, err := webport.ParseAddrBind(p)
		if err != nil {
			panic(err)
		}
		var sc webport.Config
		sc.LocalAddr = addrBind.String()
		srv, err := webport.New(sc)
		if err != nil {
			panic(fmt.Sprintf("failed to create server: %v", err))
		}
		if err := srv.Start(); err != nil {
			panic(fmt.Sprintf("failed to start server: %v", err))
		}
		fmt.Println("webport start on", addrBind.String())
		ports = append(ports, srv)
	}

	// wait signal ^C
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh

	for _, srv := range ports {
		srv.Stop()
	}
	slog.Info("Shutting down webport")
}
