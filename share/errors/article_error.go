package errors_blog

import "fmt"

// NotFoundError untuk resource yang tidak ditemukan
type NotFoundError struct {
	Resource string
	Field    string
	Value    interface{}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s dengan %s '%v' tidak ditemukan", e.Resource, e.Field, e.Value)
}

// DuplicateError untuk data yang sudah ada (unique constraint)
type DuplicateError struct {
	Resource string
	Field    string
	Value    interface{}
}

func (e *DuplicateError) Error() string {
	return fmt.Sprintf("%s dengan %s '%v' sudah ada", e.Resource, e.Field, e.Value)
}

// ValidationError untuk error validasi
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validasi error pada field '%s': %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	return "terdapat error validasi"
}
