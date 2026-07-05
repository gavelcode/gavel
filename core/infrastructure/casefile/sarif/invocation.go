package sarif

type invocation struct {
	ExecutionSuccessful            bool           `json:"executionSuccessful"`
	ToolExecutionNotifications     []notification `json:"toolExecutionNotifications"`
	ToolConfigurationNotifications []notification `json:"toolConfigurationNotifications"`
}

type notification struct {
	Level   string  `json:"level"`
	Message message `json:"message"`
}
