package main

import (
	"Tuny/core"
	"flag"
	"log"

	"github.com/rs/zerolog"
)

func main() {
	isIran := flag.Bool("i", false, "iran")
	isKharej := flag.Bool("f", false, "kharej")
	lAddr := flag.String("la", "", "local address")
	tAddr := flag.String("ta", "", "tunnel address")
	dAddr := flag.String("da", "", "destination address")
	proto := flag.Int("p", 17, "ip protocol number")
	flag.Parse()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	if *isIran {
		err := core.StartListener(core.Config{
			LocalAddr:  *lAddr,
			TunnelAddr: *tAddr,
			Proto:      *proto,
		})
		if err != nil {
			log.Fatalln(err)
		}
	} else if *isKharej {
		err := core.StartForwarder(core.Config{
			LocalAddr:  *lAddr,
			TunnelAddr: *tAddr,
			DestAddr:   *dAddr,
			Proto:      *proto,
		})
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		log.Fatalln("select iran or kharej")
	}
	<-make(chan int)
}
