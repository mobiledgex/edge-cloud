// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apis

import (
	"log"
	"net/http"
	"regexp"
)

func noRedirect(req *http.Request, via []*http.Request) error {
	// don't follow the redirects
	return http.ErrUseLastResponse
}

// does http get to token serv url, and parses redirect parameter
func GetTokenFromTokSrv(url string) string {
	client := &http.Client{
		CheckRedirect: noRedirect,
	}
	resp, err := client.Get(url)

	if err != nil {
		log.Printf("Token Client error in POST to loc service error %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	log.Printf("Received response: %+v\n", resp)

	switch resp.StatusCode {
	case http.StatusSeeOther: //303
		r, _ := regexp.Compile("dt-id=(\\S+)")
		//find the redirect location response and extract the token
		//the real mobile client will have to do something similar.
		lochdr, ok := resp.Header["Location"]
		if ok && len(lochdr) > 0 {
			m := r.FindStringSubmatch(lochdr[0])
			if len(m) == 2 {
				token := m[1]
				log.Printf("Found token %s\n", token)
				return token
			}
		}
		log.Println("Did not find match for token in response")
		return ""

	default:
		log.Printf("Error: expected 303, got %v\n", resp.StatusCode)
		return ""
	}

}
