package storage

type storageError string

const ErrNotFound = storageError("not found")

func (e storageError) Error() string {
	return string(e)
}
