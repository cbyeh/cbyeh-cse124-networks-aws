package surfstore

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

/*
Implement the logic for a client syncing with the server here.
*/
func ClientSync(client RPCClient) {

	allFileMetaMap := make(map[string]bool)
	// Look at local index.txt (create one if it doesn't exist)
	indexFilePath := "./" + client.BaseDir + "/index.txt"
	_, err := os.Stat(indexFilePath)
	if os.IsNotExist(err) {
		indFile, _ := os.Create(indexFilePath)
		indFile.Close()
	}
	// Open index.txt
	indFile, err := os.OpenFile(indexFilePath, os.O_APPEND|os.O_RDWR, 0777)
	localIndexFileInfoMap := make(map[string]FileMetaData)
	// Read through index.txt file to parse local client file metadata
	scanner := bufio.NewScanner(indFile)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ",")

		filename := fields[0]
		version, _ := strconv.Atoi(fields[1])
		blockHashList := strings.Fields(fields[2])

		var newFileMetaData FileMetaData
		newFileMetaData.Filename = filename
		newFileMetaData.Version = version
		newFileMetaData.BlockHashList = blockHashList

		localIndexFileInfoMap[filename] = newFileMetaData
		_, inSet := allFileMetaMap[filename]
		if !inSet {
			allFileMetaMap[filename] = true
		}
	}

	indFile.Truncate(0)

	clientMetaMap := make(map[string]FileMetaData)
	clientBlockMap := make(map[string]Block)

	// Look at base directory, compute each file's hash list
	dir, err := os.Open("./" + client.BaseDir)
	if err != nil {
		log.Println("Error when opening base dir")
	}
	list, _ := dir.Readdirnames(0)  // 0 to read all files and folders
	for _, filename := range list { // Create block hash list for each file in base dir
		if filename == "index.txt" {
			continue
		}
		f, _ := os.Open("./" + client.BaseDir + "/" + filename)
		i := 0
		var newBlockHashList []string
		// Read block size bytes for each until until EOF
		for {
			buf := make([]byte, client.BlockSize)
			bytesRead, err := f.Read(buf)
			if err != nil {
				break
			}
			// Encode with sha256 and hex encoding
			var newBlock Block
			// Sum(buf...)
			blockHash := sha256.Sum256(buf[:bytesRead])
			encodedHash := hex.EncodeToString(blockHash[:])
			newBlock.BlockData = buf[:bytesRead]
			newBlock.BlockSize = bytesRead
			clientBlockMap[encodedHash] = newBlock
			newBlockHashList = append(newBlockHashList, encodedHash)
			i++
		}
		var newFileMetaData FileMetaData
		newFileMetaData.Filename = filename
		newFileMetaData.BlockHashList = newBlockHashList
		newFileMetaData.Version = 1

		fLoc, inLocalIndex := localIndexFileInfoMap[filename]
		if inLocalIndex {
			if IsBlockHashListModified(newBlockHashList, fLoc.BlockHashList) {
				newFileMetaData.Version = fLoc.Version + 1
			} else {
				newFileMetaData.Version = fLoc.Version
			}
		}

		clientMetaMap[filename] = newFileMetaData
		_, inSet := allFileMetaMap[filename]
		if !inSet {
			allFileMetaMap[filename] = true
		}
		f.Close()
	}
	dir.Close()

	// Compare local index with remote index
	// 1. If files on remote index not in local or base dir, download those files, add to local index
	// 2. If files not in local OR remote (add to remote first, then if successful add to local)
	remoteFileInfoMap := make(map[string]FileMetaData)
	succ := true
	client.GetFileInfoMap(&succ, &remoteFileInfoMap)
	for filename, _ := range remoteFileInfoMap {
		_, inSet := allFileMetaMap[filename]
		if !inSet {
			allFileMetaMap[filename] = true
		}
	}

	// Conflict cases
	// 1. No local changes, remote version higher -> download blocks from remote and update local index and file in local base dir
	// 2. Uncomitted local changes, local index and remote index version same, update mapping on remote, then local index (no file change necessary)
	// 3. Local modifications to file (uncommited local changes), file version on remote > local index -> update local with remote version / bring local version of file up to date with server
	for fname, _ := range allFileMetaMap {
		fBasedDir, inBasedDir := clientMetaMap[fname]
		fLocalIndex, inLocalIndex := localIndexFileInfoMap[fname]
		fServer, inServer := remoteFileInfoMap[fname]

		var newFileMetaData FileMetaData
		var latestVersion int

		// Case: in Based dir (in or not in Local) -> Update File (get from server if err)
		// Case: not in based dir, in Local (DELETED) -> UpdateFile (get from server if err)
		// Case: Otherwise, not in EITHER -> Get from Server

		if inBasedDir || inLocalIndex {
			// Modified or New File
			if inBasedDir {
				newFileMetaData = fBasedDir
				for _, blockHash := range newFileMetaData.BlockHashList {
					newBlock := clientBlockMap[blockHash]
					succ := true
					client.PutBlock(newBlock, &succ)
				}
				// Deleted File
			} else if inLocalIndex {
				newFileMetaData.Filename = fname
				newFileMetaData.Version = fLocalIndex.Version + 1
				newFileMetaData.BlockHashList = []string{"0"}
			}

			err = client.UpdateFile(&newFileMetaData, &latestVersion)
			// If err -> download from server
			if err != nil {
				// Check if updated version is to remove file
				if len(fServer.BlockHashList) == 1 && fServer.BlockHashList[0] == "0" {
					os.Remove("./" + client.BaseDir + "/" + fname)

				} else { // If updated version is new content in the file
					updatedFile, _ := os.OpenFile("./"+client.BaseDir+"/"+fname, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0755)
					updatedFile.Truncate(0)
					for _, blockHash := range fServer.BlockHashList {
						var blockToGet Block
						err = client.GetBlock(blockHash, &blockToGet)
						if err != nil {
							log.Println("Could not get block")
							log.Println(err)
						}
						updatedFile.Write(blockToGet.BlockData)
					}
					updatedFile.Close()
				}
				// Write to index.txt server version
				indFile.Write([]byte(fname + "," + strconv.Itoa(fServer.Version) + "," + strings.Join(fServer.BlockHashList, " ") + "\n"))
			} else {
				// Write local version to index.txt
				indFile.Write([]byte(fname + "," + strconv.Itoa(newFileMetaData.Version) + "," + strings.Join(newFileMetaData.BlockHashList, " ") + "\n"))
			}
		} else if inServer { // Download from Server

			if len(fServer.BlockHashList) == 1 && fServer.BlockHashList[0] == "0" {
				os.Remove("./" + client.BaseDir + "/" + fname)

			} else { // If updated version is new content in the file
				updatedFile, _ := os.OpenFile("./"+client.BaseDir+"/"+fname, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0755)
				updatedFile.Truncate(0)
				for _, blockHash := range fServer.BlockHashList {
					var blockToGet Block
					err = client.GetBlock(blockHash, &blockToGet)
					if err != nil {
						log.Println("Could not get block")
						log.Println(err)
					}
					updatedFile.Write(blockToGet.BlockData)
				}
				updatedFile.Close()
			}
			// Write to index.txt server version
			indFile.Write([]byte(fname + "," + strconv.Itoa(fServer.Version) + "," + strings.Join(fServer.BlockHashList, " ") + "\n"))
		}
	}
	indFile.Close()

	log.Println("\n Server Map After")
	client.GetFileInfoMap(&succ, &remoteFileInfoMap)
	PrintMetaMap(remoteFileInfoMap)

}

/*
Helper function to print the contents of the metadata map.
*/
func PrintMetaMap(metaMap map[string]FileMetaData) {
	fmt.Println("--------BEGIN PRINT MAP--------")
	for _, filemeta := range metaMap {
		fmt.Println("\t", filemeta.Filename, filemeta.Version, filemeta.BlockHashList)
	}
	fmt.Println("---------END PRINT MAP--------")
}

func IsBlockHashListModified(hashList1 []string, hashList2 []string) bool {
	res := false
	if ((hashList1 == nil) != (hashList2 == nil)) || (len(hashList1) != len(hashList2)) {
		res = true
	} else {
		for i, _ := range hashList1 {
			if hashList1[i] != hashList2[i] {
				res = true
				break
			}
		}
	}
	return res
}
