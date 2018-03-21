package utils

type ReportTextCharLimitError struct{
	Msg string
}

func (e *ReportTextCharLimitError) Error	() string{
	return e.Msg
}


