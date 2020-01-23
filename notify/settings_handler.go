package notify

import (
	"context"

	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
)

// Define a settings receive handler for the common case of copying
// the settings to a global var on notify receive, and optionally performing
// some action as a result. This avoids duplicating the settings
// in both a SettingsCache and a global Settings object (the global
// object is desired for easy access).

func GlobalSettingsRecv(settings *edgeproto.Settings, updatedCb func(ctx context.Context, old *edgeproto.Settings, new *edgeproto.Settings)) *SettingsRecv {
	handler := settingsHandler{}
	handler.settings = settings
	handler.updatedCb = updatedCb
	return NewSettingsRecv(&handler)
}

type settingsHandler struct {
	settings  *edgeproto.Settings
	updatedCb func(ctx context.Context, old *edgeproto.Settings, new *edgeproto.Settings)
}

func (s *settingsHandler) Update(ctx context.Context, in *edgeproto.Settings, rev int64) {
	old := *s.settings
	*s.settings = *in
	if s.updatedCb != nil {
		s.updatedCb(ctx, &old, s.settings)
	}
}

func (s *settingsHandler) Delete(ctx context.Context, in *edgeproto.Settings, rev int64) {}

func (s *settingsHandler) Prune(ctx context.Context, keys map[edgeproto.SettingsKey]struct{}) {}

func (s *settingsHandler) Flush(ctx context.Context, notifyId int64) {}
