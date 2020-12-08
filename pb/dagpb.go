package pb

import (
	"sort"

	cid "github.com/ipfs/go-cid"
)

// PBLink ...
type PBLink struct {
	Hash  *cid.Cid
	Name  *string
	Tsize *uint64
}

// PBNode ...
type PBNode struct {
	Links []*PBLink
	Data  []byte
}

func NewPBNode() *PBNode {
	n := &PBNode{Links: make([]*PBLink, 0)}
	return n
}

func NewPBNodeFromData(data []byte) *PBNode {
	n := &PBNode{Data: data, Links: make([]*PBLink, 0)}
	return n
}

type TokenType byte

const (
	TypeData    TokenType = 'd'
	TypeLinks   TokenType = '['
	TypeLink    TokenType = 'l'
	TypeLinkEnd TokenType = 'e'
	TypeHash    TokenType = 'h'
	TypeName    TokenType = 'n'
	TypeTSize   TokenType = 's'
	TypeEnd     TokenType = 'x'
)

type Token struct {
	Type  TokenType
	Bytes []byte
	Cid   *cid.Cid
	Int   uint64
}

func NewPBLinkFromCid(c cid.Cid) *PBLink {
	l := &PBLink{Hash: &c}
	return l
}

func NewPBLink(name string, c cid.Cid, tsize uint64) *PBLink {
	l := &PBLink{Name: &name, Hash: &c, Tsize: &tsize}
	return l
}

func (node *PBNode) SortLinks() {
	SortLinks(node.Links)
}

func SortLinks(links []*PBLink) {
	sort.Stable(pbLinkSlice(links))
}

type pbLinkSlice []*PBLink

func (ls pbLinkSlice) Len() int           { return len(ls) }
func (ls pbLinkSlice) Swap(a, b int)      { ls[a], ls[b] = ls[b], ls[a] }
func (ls pbLinkSlice) Less(a, b int) bool { return pbLinkLess(ls[a], ls[b]) }

func pbLinkLess(a *PBLink, b *PBLink) bool {
	return *a.Name < *b.Name
}
