package surfstore

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type Server struct {
	BlockStore BlockStoreInterface
	MetaStore  MetaStoreInterface
}

func (s *Server) GetFileInfoMap(succ *bool, serverFileInfoMap *map[string]FileMetaData) error {
	return s.MetaStore.GetFileInfoMap(succ, serverFileInfoMap)
}

func (s *Server) UpdateFile(fileMetaData *FileMetaData, latestVersion *int) error {
	return s.MetaStore.UpdateFile(fileMetaData, latestVersion)
}

func (s *Server) GetBlock(blockHash string, blockData *Block) error {
	return s.BlockStore.GetBlock(blockHash, blockData)
}

func (s *Server) PutBlock(blockData Block, succ *bool) error {
	return s.BlockStore.PutBlock(blockData, succ)
}

func (s *Server) HasBlocks(blockHashesIn []string, blockHashesOut *[]string) error {
	return s.BlockStore.HasBlocks(blockHashesIn, blockHashesOut)
}

// This line guarantees all method for surfstore are implemented
var _ Surfstore = new(Server)

func NewSurfstoreServer() Server {
	blockStore := BlockStore{BlockMap: map[string]Block{}}
	metaStore := MetaStore{FileMetaMap: map[string]FileMetaData{}}
	return Server{
		BlockStore: &blockStore,
		MetaStore:  &metaStore,
	}
}

func ServeSurfstoreServer(hostAddr string, surfstoreServer Server) error {
	log.Println("Starting server on " + hostAddr)
	rpc.Register(&surfstoreServer)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", hostAddr)
	if e != nil {
		log.Println(e)
		return e
	}
	http.Serve(l, nil)
	return nil
}
