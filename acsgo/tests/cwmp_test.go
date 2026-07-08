package tests

import (
	"testing"

	"github.com/DWISSNET/acsgo/pkg"
)

func TestParseInformEnvelope(t *testing.T) {
	xml := `<?xml version="1.0" encoding="utf-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">100</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <cwmp:Inform>
      <cwmp:DeviceId>
        <cwmp:Manufacturer>Huawei Technologies Co., Ltd</cwmp:Manufacturer>
        <cwmp:OUI>00259E</cwmp:OUI>
        <cwmp:ProductClass>HG8245H</cwmp:ProductClass>
        <cwmp:SerialNumber>HWTXXXXXXXX001</cwmp:SerialNumber>
      </cwmp:DeviceId>
      <cwmp:Event>
        <cwmp:EventStruct>
          <cwmp:EventCode>2 PERIODIC</cwmp:EventCode>
          <cwmp:CommandKey></cwmp:CommandKey>
        </cwmp:EventStruct>
      </cwmp:Event>
      <cwmp:MaxEnvelopes>1</cwmp:MaxEnvelopes>
      <cwmp:CurrentTime>2026-01-01T00:00:00Z</cwmp:CurrentTime>
      <cwmp:RetryCount>0</cwmp:RetryCount>
      <cwmp:ParameterList>
        <cwmp:ParameterValueStruct>
          <cwmp:Name>Device.DeviceInfo.Manufacturer</cwmp:Name>
          <cwmp:Value>Huawei</cwmp:Value>
        </cwmp:ParameterValueStruct>
      </cwmp:ParameterList>
    </cwmp:Inform>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	env, err := pkg.ParseEnvelope([]byte(xml))
	if err != nil {
		t.Fatalf("ParseEnvelope error: %v", err)
	}
	if env.Body.Inform == nil {
		t.Fatal("expected Inform in body, got nil")
	}
	serial := env.Body.Inform.DeviceId.SerialNumber
	if serial != "HWTXXXXXXXX001" {
		t.Errorf("expected serial HWTXXXXXXXX001, got %s", serial)
	}
	manufacturer := env.Body.Inform.DeviceId.Manufacturer
	if manufacturer == "" {
		t.Error("expected non-empty Manufacturer")
	}
	if len(env.Body.Inform.Event.EventStruct) == 0 {
		t.Error("expected at least one EventStruct")
	}
	if env.Body.Inform.Event.EventStruct[0].EventCode != "2 PERIODIC" {
		t.Errorf("unexpected event code: %s", env.Body.Inform.Event.EventStruct[0].EventCode)
	}
}

func TestBuildInformResponse(t *testing.T) {
	resp := pkg.BuildInformResponse("42")
	if len(resp) == 0 {
		t.Fatal("expected non-empty InformResponse XML")
	}
	if !contains(resp, "InformResponse") {
		t.Error("response should contain InformResponse")
	}
	if !contains(resp, "42") {
		t.Error("response should contain ID=42")
	}
}

func TestBuildRebootRequest(t *testing.T) {
	req := pkg.BuildRebootRequest("55", "reboot-key-1")
	if !contains(req, "Reboot") {
		t.Error("reboot request should contain Reboot element")
	}
	if !contains(req, "reboot-key-1") {
		t.Error("reboot request should contain CommandKey")
	}
}

func TestBuildSetParameterValues(t *testing.T) {
	params := map[string]string{
		"Device.WiFi.SSID.1.SSID":          "MyNetwork",
		"Device.WiFi.SSID.1.KeyPassphrase":  "MyPassword",
	}
	req := pkg.BuildSetParameterValuesRequest("99", "pkey-1", params)
	if !contains(req, "SetParameterValues") {
		t.Error("expected SetParameterValues in body")
	}
	if !contains(req, "MyNetwork") {
		t.Error("expected SSID value in body")
	}
}

func TestExtractMACFromSerial(t *testing.T) {
	mac := pkg.ExtractMACFromSerial("HWTAABBCCDDEEFF")
	if mac == "00:00:00:00:00:00" {
		t.Error("expected non-zero MAC from serial")
	}
}

func TestDetectVendorFromManufacturer(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Huawei Technologies Co., Ltd", "huawei"},
		{"ZTE CORPORATION", "zte"},
		{"TP-LINK Technologies", "tplink"},
		{"MikroTik", "mikrotik"},
	}
	for _, c := range cases {
		got := pkg.DetectVendorFromManufacturer(c.input)
		if got != c.expected {
			t.Errorf("DetectVendor(%q) = %q, want %q", c.input, got, c.expected)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
