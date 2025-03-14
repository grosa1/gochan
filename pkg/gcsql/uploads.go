package gcsql

import (
	"errors"

	"github.com/gochan-org/gochan/pkg/events"
	"github.com/gochan-org/gochan/pkg/gcutil"
)

const (
	selectFilesBaseSQL = `SELECT
	id, post_id, file_order, original_filename, filename, checksum,
	file_size, is_spoilered, thumbnail_width, thumbnail_height, width, height
	FROM DBPREFIXfiles `
)

var (
	ErrAlreadyAttached = errors.New("upload already processed")
)

// GetThreadFiles gets a list of the files owned by posts in the thread, including thumbnails for convenience.
func GetThreadFiles(post *Post) ([]Upload, error) {
	query := selectFilesBaseSQL + `WHERE post_id IN (
		SELECT id FROM DBPREFIXposts WHERE thread_id = (
			SELECT thread_id FROM DBPREFIXposts WHERE id = ?)) AND filename != 'deleted'`
	rows, err := QuerySQL(query, post.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var uploads []Upload
	for rows.Next() {
		var upload Upload
		if err = rows.Scan(
			&upload.ID, &upload.PostID, &upload.FileOrder, &upload.OriginalFilename, &upload.Filename, &upload.Checksum,
			&upload.FileSize, &upload.IsSpoilered, &upload.ThumbnailWidth, &upload.ThumbnailHeight, &upload.Width, &upload.Height,
		); err != nil {
			return uploads, err
		}
		uploads = append(uploads, upload)
	}
	return uploads, nil
}

func (p *Post) nextFileOrder() (int, error) {
	const query = `SELECT COALESCE(MAX(file_order) + 1, 0) FROM DBPREFIXfiles WHERE post_id = ?`
	var next int
	err := QueryRowSQL(query, interfaceSlice(p.ID), interfaceSlice(&next))
	return next, err
}

func (p *Post) AttachFile(upload *Upload) error {
	if upload == nil {
		return nil // no upload to attach, so no error
	}
	_, recovered := events.TriggerEvent("incoming-upload", upload)
	if recovered {
		gcutil.LogWarning().Caller().
			Str("triggeredEvent", "incoming-upload").
			Str("originalFilename", upload.OriginalFilename).
			Str("filename", upload.Filename).
			Msg("Recovered from a panic in event handler")
	}

	const query = `INSERT INTO DBPREFIXfiles (
		post_id, file_order, original_filename, filename, checksum, file_size,
		is_spoilered, thumbnail_width, thumbnail_height, width, height)
	VALUES(?,?,?,?,?,?,?,?,?,?,?)`
	if upload.ID > 0 {
		return ErrAlreadyAttached
	}
	tx, err := BeginTx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := PrepareSQL(query, tx)
	if err != nil {
		return err
	}

	if upload.FileOrder < 1 {
		upload.FileOrder, err = p.nextFileOrder()
		if err != nil {
			return err
		}
	}
	upload.PostID = p.ID
	if _, err = stmt.Exec(
		&upload.PostID, &upload.FileOrder, &upload.OriginalFilename, &upload.Filename, &upload.Checksum, &upload.FileSize,
		&upload.IsSpoilered, &upload.ThumbnailWidth, &upload.ThumbnailHeight, &upload.Width, &upload.Height,
	); err != nil {
		return err
	}
	if upload.ID, err = getLatestID("DBPREFIXfiles", tx); err != nil {
		return err
	}
	return tx.Commit()
}

// ThumbnailPath returns the thumbnail path of the upload, given an thumbnail type ("thumbnail" or "catalog")
func (u *Upload) ThumbnailPath(thumbType string) string {
	return gcutil.GetThumbnailPath(thumbType, u.Filename)
}
