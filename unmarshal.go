package dagpb

import (
	"io"

	ipld "github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	pb "github.com/rvagg/go-dagpb/pb"
)

func Unmarshal(na ipld.NodeAssembler, reader io.Reader) error {
	var curLink ipld.MapAssembler

	ma, err := na.BeginMap(2)
	if err != nil {
		return err
	}
	if err = ma.AssembleKey().AssignString("Links"); err != nil {
		return err
	}
	links, err := ma.AssembleValue().BeginList(0)
	if err != nil {
		return err
	}

	tokenReceiver := func(tok pb.Token) error {
		switch tok.Type {
		case pb.TypeData:
			if err := links.Finish(); err != nil {
				return err
			}
			links = nil
			if err := ma.AssembleKey().AssignString("Data"); err != nil {
				return err
			}
			if err := ma.AssembleValue().AssignBytes(tok.Bytes); err != nil {
				return err
			}
		case pb.TypeLink:
		case pb.TypeLinkEnd:
			if err := curLink.Finish(); err != nil {
				return err
			}
		case pb.TypeHash:
			curLink, err = links.AssembleValue().BeginMap(3)
			if err != nil {
				return err
			}
			if err := curLink.AssembleKey().AssignString("Hash"); err != nil {
				return err
			}
			if err := curLink.AssembleValue().AssignLink(cidlink.Link{*tok.Cid}); err != nil {
				return err
			}
		case pb.TypeName:
			if err := curLink.AssembleKey().AssignString("Name"); err != nil {
				return err
			}
			if err := curLink.AssembleValue().AssignString(string(tok.Bytes)); err != nil {
				return err
			}
		case pb.TypeTSize:
			if err := curLink.AssembleKey().AssignString("Tsize"); err != nil {
				return err
			}
			if err := curLink.AssembleValue().AssignInt(int(tok.Int)); err != nil {
				return err
			}
		}
		return nil
	}

	if err := pb.Unmarshal(reader, tokenReceiver); err != nil {
		return err
	}
	if links != nil {
		if err := links.Finish(); err != nil {
			return err
		}
	}
	return ma.Finish()
}
