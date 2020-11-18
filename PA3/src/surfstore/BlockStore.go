package surfstore

import (
	"crypto/sha256"
	"encoding/hex"
)

type BlockStore struct {
	BlockMap map[string]Block
}

func (bs *BlockStore) GetBlock(blockHash string, blockData *Block) error {
	blockToGet, ok := bs.BlockMap[blockHash]
	if ok {
		*blockData = blockToGet
	}
	return nil
}

func (bs *BlockStore) PutBlock(block Block, succ *bool) error {
	blockHash := sha256.Sum256(block.BlockData)
	encodedHash := hex.EncodeToString(blockHash[:])
	_, ok := bs.BlockMap[encodedHash]
	*succ = false
	if !ok {
		bs.BlockMap[encodedHash] = block
		*succ = true
	}
	return nil
}

func (bs *BlockStore) HasBlocks(blockHashesIn []string, blockHashesOut *[]string) error {
	var blocksInServer []string
	for i, hash := range blockHashesIn {
		for key, _ := range bs.BlockMap {
			if key == hash {
				blocksInServer[i] = "Y"
				break
			}
		}
		if blocksInServer[i] == "" {
			blocksInServer[i] = "N"
		}
	}
	*blockHashesOut = blocksInServer
	return nil

}

// This line guarantees all method for BlockStore are implemented
var _ BlockStoreInterface = new(BlockStore)
