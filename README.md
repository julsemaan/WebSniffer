# WebSniffer

Out of band HTTP and HTTPS destination decoding

## This is a work in progress

## Throughput

Environment: 
* Tests made on Intel(R) Core(TM) i5-4670K CPU @ 3.40GHz
* Tests are parsing parsing HTTP packets on port 80 and HTTPS packets on port 443
* PCAP file contained 791615 packets total - accounting for 355417784 bytes
 * 411194 packets were on port 80 and 443 - accounting for 264590701 bytes
* A pcap filter was used as an argument to the sniffer to reduce reconstructing useless packets.
* 4 concurrent parsing threads

Parsing PPS (logged to file - not persisted to a database): 
* Command : `WebSniffer -connection_max_buffer 5 -parsing-concurrency 4 -o samples/bigFlows.pcap -dont-record-destinations -f 'tcp port 80 or 443' > out.txt 2>&1`
* Timing: 1.631s
* Pure HTTP - HTTPS parsing: 252111 PPS - 1297 Mbits/s 
* Network parsing: 485355 PPS - 1743 Mbits/s

Parsing + Persisting PPS (SQLite3)
* Command : `WebSniffer -connection_max_buffer 5 -parsing-concurrency 4 -o samples/bigFlows.pcap -f 'tcp port 80 or 443' > out.txt 2>&1`
* Timing: 42.169
* Pure HTTP - HTTPS parsing: 9751 PPS - 50 Mbits/s 
* Network parsing: 18772 PPS - 67 Mbits/s
* As seen above, persisting in the SQLite3 backend is far from being performant and becomes the central bottleneck of the application. This _should_ not be used other than for testing.

## Licence

GPL

