package suci

import (
	"fmt"
	"testing"
)

func TestToSupi(t *testing.T) {
	suciProfiles := []SuciProfile{
		{
			ProtectionScheme: "1", // Protect Scheme: Profile A
			PrivateKey:       "c53c22208b61860b06c62e5406a7b330c2b577aa5558981510d128247d38bd1d",
			PublicKey:        "5a8d38864820197c3394b92613b20b91633cbd897119273bf8e4a6f4eec0a650",
		},
		{
			ProtectionScheme: "2", // Protect Scheme: Profile B
			PrivateKey:       "F1AB1074477EBCC7F554EA1C5FC368B1616730155E0041AC447D6301975FECDA",
			PublicKey: "0472DA71976234CE833A6907425867B82E074D44EF907DFB4B3E21C1C2256EBCD" +
				"15A7DED52FCBB097A4ED250E036C7B9C8C7004C4EEDC4F068CD7BF8D3F900E3B4",
		},
	}
	testCases := []struct {
		suci         string
		expectedSupi string
		expectedErr  error
	}{
		{
			suci:         "suci-0-208-93-0-0-0-00007487",
			expectedSupi: "imsi-2089300007487",
			expectedErr:  nil,
		},
		{
			suci: "suci-0-208-93-0-1-1-b2e92f836055a255837debf850b528997ce0201cb82a" +
				"dfe4be1f587d07d8457dcb02352410cddd9e730ef3fa87",
			expectedSupi: "imsi-20893001002086",
			expectedErr:  nil,
		},
		{
			suci: "suci-0-208-93-0-2-2-039aab8376597021e855679a9778ea0b67396e68c66d" +
				"f32c0f41e9acca2da9b9d146a33fc2716ac7dae96aa30a4d",
			expectedSupi: "imsi-20893001002086",
			expectedErr:  nil,
		},
		{
			suci: "suci-0-208-93-0-2-2-0434a66778799d52fedd9326db4b690d092e05c9ba0ace5b413da" +
				"fc0a40aa28ee00a79f790fa4da6a2ece892423adb130dc1b30e270b7d0088bdd716b93894891d5221a74c810d6b9350cc067c76",
			expectedSupi: "",
			expectedErr:  fmt.Errorf("crypto/elliptic: attempted operation on invalid point"),
		},
	}
	for i, tc := range testCases {
		supi, err := ToSupi(tc.suci, suciProfiles)
		if err != nil {
			if err.Error() != tc.expectedErr.Error() {
				t.Errorf("TC%d fail: err[%s], expected[%s]\n", i, err, tc.expectedErr)
			}
		} else if supi != tc.expectedSupi {
			t.Errorf("TC%d fail: supi[%s], expected[%s]\n", i, supi, tc.expectedSupi)
		}
	}
}
