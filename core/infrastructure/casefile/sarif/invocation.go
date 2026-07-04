package sarif

type invocation struct {
	ExecutionSuccessful            bool           `json:"executionSuccessful"`
	ToolExecutionNotifications     []notification `json:"toolExecutionNotifications"`
	ToolConfigurationNotifications []notification `json:"toolConfigurationNotifications"`
}

type notification struct {
	Message message `json:"message"`
}
