package service

import "github.com/MrMiaoMIMI/goshared/util/servererr"

func validationErrorf(format string, args ...any) error {
	return servererr.NewBizErrorf(servererr.ErrBadRequest, format, args...)
}

func conflictErrorf(format string, args ...any) error {
	return servererr.NewBizErrorf(servererr.ErrConflict, format, args...)
}

func notFoundErrorf(format string, args ...any) error {
	return servererr.NewBizErrorf(servererr.ErrNotFound, format, args...)
}
