package modelfmo

import (
	"github.com/MrMiaoMIMI/goshared/db/dbhelper"
	"github.com/MrMiaoMIMI/goshared/db/dbspi"
)

type CommonFmo struct {
	ID      dbspi.Field[uint64]
	Creator dbspi.Field[string]
	Updater dbspi.Field[string]
	Ctime   dbspi.Field[uint64]
	Mtime   dbspi.Field[uint64]
	Deleted dbspi.Field[bool]
}

func NewCommonFmo() CommonFmo {
	return CommonFmo{
		ID:      dbhelper.NewField[uint64](dbspi.DefaultIdFieldName),
		Creator: dbhelper.NewField[string](dbspi.DefaultCreatorFieldName),
		Updater: dbhelper.NewField[string](dbspi.DefaultUpdaterFieldName),
		Ctime:   dbhelper.NewField[uint64](dbspi.DefaultCtimeFieldName),
		Mtime:   dbhelper.NewField[uint64](dbspi.DefaultMtimeFieldName),
		Deleted: dbhelper.NewField[bool](dbspi.DefaultDeletedFieldName),
	}
}
