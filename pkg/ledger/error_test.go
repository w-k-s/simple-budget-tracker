package ledger

import "github.com/w-k-s/simple-budget-tracker/pkg"

func errorFields(err error) map[string]string {
	if errWithFields, ok := err.(interface {
		InvalidFields() map[string]string
	}); ok {
		return errWithFields.InvalidFields()
	}
	return map[string]string{}
}

func errorCode(err error, defaultValue uint64) pkg.ErrorCode {
	if errWithCode, ok := err.(interface {
		Code() uint64
	}); ok {
		return pkg.ErrorCode(errWithCode.Code())
	}
	return pkg.ErrorCode(defaultValue)
}
