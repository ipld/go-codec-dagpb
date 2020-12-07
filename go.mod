module github.com/rvagg/go-dagpb/dagpb

go 1.15

require (
	github.com/ipfs/go-cid v0.0.7
	github.com/ipld/go-ipld-prime v0.5.1-0.20201114141345-1110155de1fb
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-multihash v0.0.14 // indirect
	github.com/multiformats/go-varint v0.0.6 // indirect
	github.com/rvagg/go-dagpb/pb v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.0.0-20201124201722-c8d3bf9c5392 // indirect
	golang.org/x/sys v0.0.0-20201202213521-69691e467435 // indirect
)

replace github.com/rvagg/go-dagpb/pb => ./pb
