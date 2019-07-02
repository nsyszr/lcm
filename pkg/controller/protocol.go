package controller

type CommandRequest struct {
	DeviceID  string
	Command   string
	Arguments interface{}
}

type CommandReqly struct {
	DeviceID string
	Results  interface{}
}
