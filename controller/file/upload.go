package file

import (
	"context"
	"errors"

	"github.com/qml-123/AppService/controller/file"
	"github.com/qml-123/AppService/pkg/db"
	"github.com/qml-123/app_log/error_code"
	"gorm.io/gorm"
)

func Upload(ctx context.Context, user_id int64, file_content []byte, chunk_num, chunk_size int32, file_key string, has_more bool, file_type string) (err error) {
	file_share := &db.FileShare{}
	result := db.GetDB().Select("id").Where("file_key = ? and user_id = ? and permission = ? and `delete` = ?", file_key, user_id, file.FileOwner, false).First(file_share)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return error_code.NoPermission
		}
		return result.Error
	}

	file := &db.File{
		FileKey:   file_key,
		Chunk:     file_content,
		ChunkNum:  int(chunk_num),
		ChunkSize: int(chunk_size),
		OwnUserID: user_id,
		FileType:  file_type,
		HasMore:   has_more,
		Delete:    false,
	}
	result = db.GetDB().Create(file)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
