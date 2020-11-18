package surfstore

import (
	"errors"
)

type MetaStore struct {
	FileMetaMap map[string]FileMetaData
}

func (m *MetaStore) GetFileInfoMap(_ignore *bool, serverFileInfoMap *map[string]FileMetaData) error {
	*serverFileInfoMap = m.FileMetaMap
	return nil
}

func (m *MetaStore) UpdateFile(fileMetaData *FileMetaData, latestVersion *int) (err error) {
	metafile, exist := m.FileMetaMap[fileMetaData.Filename]
	if exist {
		if metafile.Version == fileMetaData.Version-1 {
			*latestVersion = fileMetaData.Version
			m.FileMetaMap[fileMetaData.Filename] = *fileMetaData
		} else {
			*latestVersion = metafile.Version
			return errors.New("Version mismatch")
		}
	} else {
		// fileMetaData.Version = 1
		*latestVersion = fileMetaData.Version
		m.FileMetaMap[fileMetaData.Filename] = *fileMetaData
	}
	return nil
}

var _ MetaStoreInterface = new(MetaStore)
