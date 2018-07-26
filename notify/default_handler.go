package notify

// DefaultHandler implements both ClientHandler and ServerHandler.
// It starts out with all handlers nil, so no actions
// will be taken by the notify code. The handlers can be assigned
// depending upon what the calling code wants to implement.
type DefaultHandler struct {
	// server handlers
	SendAppInst         SendAppInstHandler
	SendCloudlet        SendCloudletHandler
	SendFlavor          SendFlavorHandler
	SendClusterInst     SendClusterInstHandler
	RecvAppInstInfo     RecvAppInstInfoHandler
	RecvClusterInstInfo RecvClusterInstInfoHandler
	RecvCloudletInfo    RecvCloudletInfoHandler
	// client handlers
	SendAppInstInfo     SendAppInstInfoHandler
	SendClusterInstInfo SendClusterInstInfoHandler
	SendCloudletInfo    SendCloudletInfoHandler
	RecvAppInst         RecvAppInstHandler
	RecvCloudlet        RecvCloudletHandler
	RecvFlavor          RecvFlavorHandler
	RecvClusterInst     RecvClusterInstHandler
}

func (s *DefaultHandler) SendAppInstHandler() SendAppInstHandler {
	return s.SendAppInst
}

func (s *DefaultHandler) SendCloudletHandler() SendCloudletHandler {
	return s.SendCloudlet
}

func (s *DefaultHandler) SendFlavorHandler() SendFlavorHandler {
	return s.SendFlavor
}

func (s *DefaultHandler) SendClusterInstHandler() SendClusterInstHandler {
	return s.SendClusterInst
}

func (s *DefaultHandler) RecvAppInstInfoHandler() RecvAppInstInfoHandler {
	return s.RecvAppInstInfo
}

func (s *DefaultHandler) RecvClusterInstInfoHandler() RecvClusterInstInfoHandler {
	return s.RecvClusterInstInfo
}

func (s *DefaultHandler) RecvCloudletInfoHandler() RecvCloudletInfoHandler {
	return s.RecvCloudletInfo
}

func (s *DefaultHandler) SendAppInstInfoHandler() SendAppInstInfoHandler {
	return s.SendAppInstInfo
}

func (s *DefaultHandler) SendClusterInstInfoHandler() SendClusterInstInfoHandler {
	return s.SendClusterInstInfo
}

func (s *DefaultHandler) SendCloudletInfoHandler() SendCloudletInfoHandler {
	return s.SendCloudletInfo
}

func (s *DefaultHandler) RecvAppInstHandler() RecvAppInstHandler {
	return s.RecvAppInst
}

func (s *DefaultHandler) RecvCloudletHandler() RecvCloudletHandler {
	return s.RecvCloudlet
}

func (s *DefaultHandler) RecvFlavorHandler() RecvFlavorHandler {
	return s.RecvFlavor
}

func (s *DefaultHandler) RecvClusterInstHandler() RecvClusterInstHandler {
	return s.RecvClusterInst
}
