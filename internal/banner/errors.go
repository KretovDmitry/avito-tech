package banner

import "fmt"

type InvalidTypeError struct {
	ParamName string
}

func (e *InvalidTypeError) Error() string {
	return fmt.Sprintf("invalid type for parameter: %s", e.ParamName)
}
