package ledger

func errorFields(err error) map[string]string {
	if errWithFields, ok := err.(interface {
		InvalidFields() map[string]string
	}); ok {
		return errWithFields.InvalidFields()
	}
	return map[string]string{}
}

func errorCode(err error, defaultValue uint64) uint64 {
	if errWithCode, ok := err.(interface {
		Code() uint64
	}); ok {
		return errWithCode.Code()
	}
	return defaultValue
}
