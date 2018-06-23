package dmetest

import dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"

type VerifyLocRR struct {
	Req   dme.Match_Engine_Request
	Reply dme.Match_Engine_Loc_Verify
}

// VerifyLocation API test data.
// Replies are based on AppInst data generated by GenerateAppInsts()
// in this package.
var VerifyLocData = []VerifyLocRR{
	VerifyLocRR{
		Req: dme.Match_Engine_Request{
			CarrierID:   1,
			CarrierName: "TDG",
			GpsLocation: &dme.Loc{Lat: 50.73, Long: 7.1},
			DevName:     "1000realities",
			AppName:     "1000realities",
			AppVers:     "1.1",
		},
		Reply: dme.Match_Engine_Loc_Verify{
			GpsLocationStatus: 1,
		},
	},
	VerifyLocRR{
		Req: dme.Match_Engine_Request{
			CarrierID:   1,
			CarrierName: "TDG",
			GpsLocation: &dme.Loc{Lat: 52.65, Long: 12.341},
			DevName:     "1000realities",
			AppName:     "1000realities",
			AppVers:     "1.1",
		},
		Reply: dme.Match_Engine_Loc_Verify{
			GpsLocationStatus: 3,
		},
	},
	VerifyLocRR{
		Req: dme.Match_Engine_Request{
			CarrierID:   5,
			CarrierName: "ATT",
			GpsLocation: &dme.Loc{Lat: 52.65, Long: 10.341},
			DevName:     "1000realities",
			AppName:     "1000realities",
			AppVers:     "1.1",
		},
		Reply: dme.Match_Engine_Loc_Verify{
			GpsLocationStatus: 0,
		},
	},
	VerifyLocRR{
		Req: dme.Match_Engine_Request{
			CarrierID:   1,
			CarrierName: "TDG",
			GpsLocation: &dme.Loc{Lat: 50.75, Long: 7.9050},
			AppId:       5000,
			DevName:     "1000realities",
			AppName:     "1000realities",
			AppVers:     "1.1",
		},
		Reply: dme.Match_Engine_Loc_Verify{
			GpsLocationStatus: 3,
		},
	},
	VerifyLocRR{
		Req: dme.Match_Engine_Request{
			CarrierID:   1,
			CarrierName: "TDG",
			GpsLocation: &dme.Loc{Lat: 52.75, Long: 12.9050},
			AppId:       5005,
			DevName:     "Niantic Labs",
			AppName:     "Pokemon-go",
			AppVers:     "2.1",
		},
		Reply: dme.Match_Engine_Loc_Verify{
			GpsLocationStatus: 3,
		},
	},
	VerifyLocRR{
		Req: dme.Match_Engine_Request{
			CarrierID:   1,
			CarrierName: "TDG",
			GpsLocation: &dme.Loc{Lat: 50.75, Long: 11.9050},
			AppId:       5006,
			DevName:     "Niantic Labs",
			AppName:     "HarryPotter-go",
			AppVers:     "1.0",
		},
		Reply: dme.Match_Engine_Loc_Verify{
			GpsLocationStatus: 4,
		},
	},
	VerifyLocRR{
		Req: dme.Match_Engine_Request{
			CarrierID:   3,
			CarrierName: "TMUS",
			GpsLocation: &dme.Loc{Lat: 47.75, Long: 122.9050},
			AppId:       5010,
			DevName:     "Ever.AI",
			AppName:     "Ever",
			AppVers:     "1.7",
		},
		Reply: dme.Match_Engine_Loc_Verify{
			GpsLocationStatus: 3,
		},
	},
}
