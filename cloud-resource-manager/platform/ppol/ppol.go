package ppol

/*
func GetCloudletPrivacyPolicy(ctx context.Context, platformConfig *platform.PlatformConfig, caches *platform.Caches) (*edgeproto.PrivacyPolicy, error) {
	log.WarnLog("GetPrivacyPolicy")
	if platformConfig.PrivacyPolicy != "" {
		pp := edgeproto.PrivacyPolicy{}
		pk := edgeproto.PolicyKey{
			Name:         platformConfig.PrivacyPolicy,
			Organization: platformConfig.CloudletKey.Organization,
		}
		if !caches.PrivacyPolicyCache.Get(&pk, &pp) {
			log.SpanLog(ctx, log.DebugLevelInfra, "Cannot find Privacy Policy from cache", "pk", pk, "pp", pp)
			return nil, fmt.Errorf("fail to find Privacy Policy from cache: %s", pk)
		} else {
			log.SpanLog(ctx, log.DebugLevelInfra, "Found Privacy Policy from cache", "pk", pk, "pp", pp)
			return &pp, nil
		}
	}
	return nil, fmt.Errorf("No privacy policy specified")
}
*/
