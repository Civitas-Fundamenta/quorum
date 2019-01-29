package raft

import (
	"io"
	"net"

	"fmt"
	"log"

	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rlp"
)

const nodeIDBits = 512

type EnodeID [nodeIDBits / 8]byte

// Serializable information about a Peer. Sufficient to build `etcdRaft.Peer`
// or `discover.Node`.
type Address struct {
	RaftId   uint16       `json:"raftId"`
	NodeId   [64]byte     `json:"nodeId"`
	Ip       net.IP       `json:"ip"`
	P2pPort  enr.TCP      `json:"p2pPort"`
	RaftPort enr.RAFTPORT `json:"raftPort"`
	PubKey   *ecdsa.PublicKey
}

func newAddress(raftId uint16, raftPort int, node *enode.Node) *Address {
	node.ID().Bytes()
	return &Address{
		RaftId:   raftId,
		NodeId:   []byte(node.EnodeID()),
		Ip:       node.IP(),
		P2pPort:  enr.TCP(node.TCP()),
		RaftPort: enr.RAFTPORT(raftPort),
		PubKey:   node.Pubkey(),
	}
}

// A peer that we're connected to via both raft's http transport, and ethereum p2p
type Peer struct {
	address *Address    // For raft transport
	p2pNode *enode.Node // For ethereum transport
}

func (addr *Address) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{addr.RaftId, addr.NodeId, addr.Ip, addr.P2pPort, addr.RaftPort})
}

func (addr *Address) DecodeRLP(s *rlp.Stream) error {
	// These fields need to be public:
	var temp struct {
		RaftId   uint16
		NodeId   enode.ID
		Ip       net.IP
		P2pPort  enr.TCP
		RaftPort enr.RAFTPORT
	}

	if err := s.Decode(&temp); err != nil {
		return err
	} else {
		addr.RaftId, addr.NodeId, addr.Ip, addr.P2pPort, addr.RaftPort = temp.RaftId, temp.NodeId, temp.Ip, temp.P2pPort, temp.RaftPort
		return nil
	}
}

// RLP Address encoding, for transport over raft and storage in LevelDB.

func (addr *Address) toBytes() []byte {
	size, r, err := rlp.EncodeToReader(addr)
	if err != nil {
		panic(fmt.Sprintf("error: failed to RLP-encode Address: %s", err.Error()))
	}
	var buffer = make([]byte, uint32(size))
	r.Read(buffer)

	return buffer
}

func bytesToAddress(bytes []byte) *Address {
	var addr Address
	if err := rlp.DecodeBytes(bytes, &addr); err != nil {
		log.Fatalf("failed to RLP-decode Address: %v", err)
	}
	return &addr
}
