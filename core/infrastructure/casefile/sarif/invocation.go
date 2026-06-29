package sarif

type invocation struct {
	ExecutionSuccessful        bool           `json:"executionSuccessful"`
	ToolExecutionNotifications []notification `json:"toolExecutionNotifications"`
}

type notification struct {
	Message message `json:"message"`
}
