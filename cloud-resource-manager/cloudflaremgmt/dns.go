package cloudflaremgmt

import (
	"context"
	"fmt"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/mobiledgex/edge-cloud/log"
)

var LocalTestZone = "localtest.net"

// GetDNSRecords returns a list of DNS records for the given domain name. Error returned otherwise.
// if name is provided, that is used as a filter
func GetDNSRecords(ctx context.Context, api *cloudflare.API, zone string, name string) ([]cloudflare.DNSRecord, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetDNSRecords", "name", name)

	if zone == "" {
		return nil, fmt.Errorf("missing domain zone")
	}

	zoneID, err := api.ZoneIDByName(zone)
	if err != nil {
		return nil, err
	}

	queryRecord := cloudflare.DNSRecord{}
	if name != "" {
		queryRecord.Name = name
	}

	records, err := api.DNSRecords(zoneID, queryRecord)
	if err != nil {
		return nil, err
	}
	return records, nil
}

// CreateOrUpdateDNSRecord changes the existing record if found, or adds a new one
func CreateOrUpdateDNSRecord(ctx context.Context, api *cloudflare.API, zone, name, rtype, content string, ttl int, proxy bool) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateOrUpdateDNSRecord", "zone", zone, "name", name, "content", content)

	if !strings.Contains(name, zone) {
		// mismatch between what the controller thinks is the appdnsroot is and what the value
		// the CRM is running with.
		return fmt.Errorf("Mismatch between requested DNS record zone: %s and CRM zone: %s", name, zone)
	}

	if zone == LocalTestZone {
		log.SpanLog(ctx, log.DebugLevelInfra, "Skip record creation for test zone", "zone", zone)
		return nil
	}

	zoneID, err := api.ZoneIDByName(zone)
	if err != nil {
		return err
	}

	queryRecord := cloudflare.DNSRecord{
		Name: strings.ToLower(name),
		Type: strings.ToUpper(rtype),
	}
	records, err := api.DNSRecords(zoneID, queryRecord)
	if err != nil {
		return err
	}
	found := false
	for _, r := range records {
		found = true
		if r.Content == content {
			log.SpanLog(ctx, log.DebugLevelInfra, "CreateOrUpdateDNSRecord existing record matches", "name", name, "content", content)
		} else {
			log.SpanLog(ctx, log.DebugLevelInfra, "CreateOrUpdateDNSRecord updating", "name", name, "content", content)

			updateRecord := cloudflare.DNSRecord{
				Name:    strings.ToLower(name),
				Type:    strings.ToUpper(rtype),
				Content: content,
				TTL:     ttl,
				Proxied: proxy,
			}
			err := api.UpdateDNSRecord(zoneID, r.ID, updateRecord)
			if err != nil {
				return fmt.Errorf("cannot update DNS record for zone %s name %s, %v", zone, name, err)
			}
		}
	}
	if !found {
		addRecord := cloudflare.DNSRecord{
			Name:    strings.ToLower(name),
			Type:    strings.ToUpper(rtype),
			Content: content,
			TTL:     ttl,
			Proxied: false,
		}
		_, err := api.CreateDNSRecord(zoneID, addRecord)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "CreateOrUpdateDNSRecord failed", "zone", zone, "name", name, "err", err)
			return fmt.Errorf("cannot create DNS record for zone %s, %v", zone, err)
		}
	}
	return nil
}

// CreateDNSRecord creates a new DNS record for the zone
func CreateDNSRecord(ctx context.Context, api *cloudflare.API, zone, name, rtype, content string, ttl int, proxy bool) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateDNSRecord", "name", name, "content", content)

	if zone == "" {
		return fmt.Errorf("missing zone")
	}

	if name == "" {
		return fmt.Errorf("missing name")
	}

	if rtype == "" {
		return fmt.Errorf("missing rtype")
	}

	if content == "" {
		return fmt.Errorf("missing content")
	}

	if ttl <= 0 {
		return fmt.Errorf("invalid TTL")
	}
	//ttl = 1 // automatic

	zoneID, err := api.ZoneIDByName(zone)
	if err != nil {
		return err
	}

	record := cloudflare.DNSRecord{
		Name:    name,
		Type:    strings.ToUpper(rtype),
		Content: content,
		TTL:     ttl,
		Proxied: proxy,
	}

	_, err = api.CreateDNSRecord(zoneID, record)
	if err != nil {
		return fmt.Errorf("cannot create DNS record for zone %s, %v", zone, err)
	}

	return nil
}

// DeleteDNSRecord deletes DNS record specified by recordID in zone.
func DeleteDNSRecord(ctx context.Context, api *cloudflare.API, zone, recordID string) error {
	if zone == LocalTestZone {
		return nil
	}
	if zone == "" {
		return fmt.Errorf("missing zone")
	}

	if recordID == "" {
		return fmt.Errorf("missing recordID")
	}

	zoneID, err := api.ZoneIDByName(zone)
	if err != nil {
		return err
	}

	return api.DeleteDNSRecord(zoneID, recordID)
}
