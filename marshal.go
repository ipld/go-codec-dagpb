package dagpb

import (
	"fmt"
	"io"

	ipld "github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	pb "github.com/rvagg/go-dagpb/pb"
)

func Marshal(inNode ipld.Node, out io.Writer) error {
	builder := Type.PBNode.NewBuilder()
	if err := builder.AssignNode(inNode); err != nil {
		return err
	}
	node := builder.Build()

	curField := pb.TypeLinks
	var linksIter ipld.ListIterator
	var link ipld.Node

	tokenSource := func() (pb.Token, error) {
		if curField == pb.TypeLinks {
			links, err := node.LookupByString("Links")
			if err != nil {
				return pb.Token{}, err
			}
			if links.Length() == 0 {
				curField = pb.TypeData
			} else {
				curField = pb.TypeHash
				linksIter = links.ListIterator()
				_, link, err = linksIter.Next()
				if err != nil {
					return pb.Token{}, err
				}
			}
		}

		if curField == pb.TypeData {
			curField = pb.TypeEnd

			d, err := node.LookupByString("Data")
			if err != nil {
				return pb.Token{}, err
			}
			if !d.IsAbsent() {
				b, err := d.AsBytes()
				if err != nil {
					return pb.Token{}, err
				}
				return pb.Token{Type: pb.TypeData, Bytes: b}, nil
			}
		}

		if curField == pb.TypeEnd {
			return pb.Token{Type: pb.TypeEnd}, nil
		}

		for {
			if curField == pb.TypeHash {
				curField = pb.TypeName

				d, err := link.LookupByString("Hash")
				if err != nil {
					return pb.Token{}, err
				}
				l, err := d.AsLink()
				if err != nil {
					return pb.Token{}, err
				}
				if cl, ok := l.(cidlink.Link); ok {
					return pb.Token{Type: pb.TypeHash, Cid: &cl.Cid}, nil
				}
				return pb.Token{}, fmt.Errorf("unexpected Link type [%v]", l)
			}

			if curField == pb.TypeName {
				curField = pb.TypeTSize

				nameNode, err := link.LookupByString("Name")
				if err != nil {
					return pb.Token{}, err
				}
				if !nameNode.IsAbsent() {
					name, err := nameNode.AsString()
					if err != nil {
						return pb.Token{}, err
					}
					return pb.Token{Type: pb.TypeName, Bytes: []byte(name)}, nil
				}
			}

			if curField == pb.TypeTSize {
				curField = pb.TypeLinkEnd

				tsizeNode, err := link.LookupByString("Tsize")
				if err != nil {
					return pb.Token{}, err
				}
				if !tsizeNode.IsAbsent() {
					tsize, err := tsizeNode.AsInt()
					if err != nil {
						return pb.Token{}, err
					}
					if tsize < 0 {
						return pb.Token{}, fmt.Errorf("Link has negative Tsize value [%v]", tsize)
					}
					return pb.Token{Type: pb.TypeTSize, Int: uint64(tsize)}, nil
				}
			}

			if curField == pb.TypeLinkEnd {
				if linksIter.Done() {
					curField = pb.TypeData
				} else {
					curField = pb.TypeHash
					var err error
					_, link, err = linksIter.Next()
					if err != nil {
						return pb.Token{}, err
					}
				}
				return pb.Token{Type: pb.TypeLinkEnd}, nil
			}

			if curField != pb.TypeHash {
				return pb.Token{}, fmt.Errorf("unexpected and invalid token state")
			}
		}
	}

	return pb.Marshal(out, tokenSource)
}
