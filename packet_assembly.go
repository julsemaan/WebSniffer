package main

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"github.com/julsemaan/WebSniffer/log"
	WebSnifferUtil "github.com/julsemaan/WebSniffer/util"
	"time"
)

// simpleStreamFactory implements tcpassembly.StreamFactory
type sniffStreamFactory struct{}

// sniffStream will handle the actual decoding of sniff requests.
type sniffStream struct {
	net, transport                         gopacket.Flow
	bytesLen, packets, outOfOrder, skipped int64
	start, end                             time.Time
	sawStart, sawEnd                       bool
	bytes                                  []byte
}

// New creates a new stream.  It's called whenever the assembler sees a stream
// it isn't currently following.
func (factory *sniffStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	//	log.Printf("new stream %v:%v started", net, transport)
	s := &sniffStream{
		net:       net,
		transport: transport,
		start:     time.Now(),
	}
	s.end = s.start
	// ReaderStream implements tcpassembly.Stream, so we can return a pointer to it.
	return s
}

// Reassembled is called whenever new packet data is available for reading.
// Reassembly objects contain stream data IN ORDER.
func (s *sniffStream) Reassembled(reassemblies []tcpassembly.Reassembly) {
	for _, reassembly := range reassemblies {
		if reassembly.Seen.Before(s.end) {
			s.outOfOrder++
		} else {
			s.end = reassembly.Seen
		}
		s.bytesLen += int64(len(reassembly.Bytes))
		s.packets += 1
		if reassembly.Skip > 0 {
			s.skipped += int64(reassembly.Skip)
		}
		s.bytes = append(s.bytes, reassembly.Bytes...)
		s.sawStart = s.sawStart || reassembly.Start
		s.sawEnd = s.sawEnd || reassembly.End
	}
}

// ReassemblyComplete is called when the TCP assembler believes a stream has
// finished.
func (s *sniffStream) ReassemblyComplete() {
	//diffSecs := float64(s.end.Sub(s.start)) / float64(time.Second)
	//	log.Printf("Reassembly of stream %v:%v complete - start:%v end:%v bytes:%v packets:%v ooo:%v bps:%v pps:%v skipped:%v",
	//s.net, s.transport, s.start, s.end, s.bytesLen, s.packets, s.outOfOrder,
	//float64(s.bytesLen)/diffSecs, float64(s.packets)/diffSecs, s.skipped)

	go func() {
		wg.Add(1)
		parsingConcurrencyChan <- 1

		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if ok && err.Error() == "runtime error: index out of range" {
					log.Logger().Debug("Error decoding packet due to its unknown format. This is likely normal.", err.Error())
				} else {
					log.Logger().Error("Error decoding packet.", r)
				}
			}
			<-parsingConcurrencyChan
			wg.Done()
		}()

		var destination *Destination
		if unencryptedPorts[s.transport.Src().String()] || unencryptedPorts[s.transport.Dst().String()] {
			http_packet := &WebSnifferUtil.Packet{Hosts: s.net, Ports: s.transport, Payload: s.bytes}
			destination = ParseHTTP(http_packet)
			if destination != nil {
				log.Logger().Info("Found the following server name (HTTP) : ", destination.ServerName)
				recordingQueue.push(destination)
			}
		}

		if encryptedPorts[s.transport.Src().String()] || encryptedPorts[s.transport.Dst().String()] {
			https_packet := &WebSnifferUtil.Packet{Hosts: s.net, Ports: s.transport, Payload: s.bytes}
			destination = ParseHTTPS(https_packet)
			if destination != nil {
				log.Logger().Info("Found the following server name (HTTPS) : ", destination.ServerName)
				recordingQueue.push(destination)
			}
		}

		<-parsingConcurrencyChan
		wg.Done()
	}()
}