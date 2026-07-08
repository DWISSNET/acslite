package soap_test

import (
	"strings"
	"testing"

	"github.com/DWISSNET/acsgo/services/cwmp/soap"
)

func TestParseEnvelope_Inform(t *testing.T) {
	rawSOAP := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">42</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <cwmp:Inform>
      <DeviceId>
        <Manufacturer>Huawei</Manufacturer>
        <OUI>AABBCC</OUI>
        <ProductClass>HG8245H</ProductClass>
        <SerialNumber>12345678</SerialNumber>
      </DeviceId>
      <Event SOAP-ENC:arrayType="cwmp:EventStruct[1]">
        <EventStruct>
          <EventCode>0 BOOTSTRAP</EventCode>
          <CommandKey></CommandKey>
        </EventStruct>
      </Event>
      <MaxEnvelopes>1</MaxEnvelopes>
      <CurrentTime>2026-01-01T00:00:00Z</CurrentTime>
      <RetryCount>0</RetryCount>
      <ParameterList SOAP-ENC:arrayType="cwmp:ParameterValueStruct[2]">
        <ParameterValueStruct>
          <Name>Device.DeviceInfo.SoftwareVersion</Name>
          <Value>V300R017C10SPC150</Value>
        </ParameterValueStruct>
        <ParameterValueStruct>
          <Name>Device.DeviceInfo.HardwareVersion</Name>
          <Value>VER.B</Value>
        </ParameterValueStruct>
      </ParameterList>
    </cwmp:Inform>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	env, err := soap.ParseEnvelope([]byte(rawSOAP))
	if err != nil {
		t.Fatalf("ParseEnvelope failed: %v", err)
	}

	if env.Header.ID != "42" {
		t.Errorf("expected ID=42, got %q", env.Header.ID)
	}
	if env.Body.Inform == nil {
		t.Fatal("expected Inform, got nil")
	}
	inf := env.Body.Inform
	if inf.DeviceId.Manufacturer != "Huawei" {
		t.Errorf("expected Manufacturer=Huawei, got %q", inf.DeviceId.Manufacturer)
	}
	if inf.DeviceId.SerialNumber != "12345678" {
		t.Errorf("expected SerialNumber=12345678, got %q", inf.DeviceId.SerialNumber)
	}
	if len(inf.ParameterList) != 2 {
		t.Errorf("expected 2 parameters, got %d", len(inf.ParameterList))
	}
	if inf.ParameterList[0].Name != "Device.DeviceInfo.SoftwareVersion" {
		t.Errorf("unexpected parameter name: %q", inf.ParameterList[0].Name)
	}
}

func TestInformResponse(t *testing.T) {
	resp := soap.InformResponse("test-id-123")
	if !strings.Contains(resp, "InformResponse") {
		t.Error("InformResponse should contain InformResponse element")
	}
	if !strings.Contains(resp, "test-id-123") {
		t.Error("InformResponse should contain the CWMP ID")
	}
	if !strings.Contains(resp, "MaxEnvelopes") {
		t.Error("InformResponse should contain MaxEnvelopes")
	}
}

func TestRebootRequest(t *testing.T) {
	req := soap.RebootRequest("id-001", "cmd-key-001")
	if !strings.Contains(req, "Reboot") {
		t.Error("RebootRequest should contain Reboot element")
	}
	if !strings.Contains(req, "cmd-key-001") {
		t.Error("RebootRequest should contain command key")
	}
}

func TestSetParameterValuesRequest(t *testing.T) {
	params := map[string]string{
		"Device.WiFi.SSID": "MyNetwork",
	}
	req := soap.SetParameterValuesRequest("id-002", "key-002", params)
	if !strings.Contains(req, "SetParameterValues") {
		t.Error("should contain SetParameterValues element")
	}
	if !strings.Contains(req, "Device.WiFi.SSID") {
		t.Error("should contain parameter name")
	}
	if !strings.Contains(req, "MyNetwork") {
		t.Error("should contain parameter value")
	}
}

func TestGetRPCMethodsResponse(t *testing.T) {
	resp := soap.GetRPCMethodsResponse("id-003")
	if !strings.Contains(resp, "GetRPCMethodsResponse") {
		t.Error("should contain GetRPCMethodsResponse element")
	}
	for _, method := range []string{"Reboot", "FactoryReset", "Download", "GetParameterValues"} {
		if !strings.Contains(resp, method) {
			t.Errorf("GetRPCMethodsResponse should list %s", method)
		}
	}
}

func TestEmptyResponse(t *testing.T) {
	resp := soap.EmptyResponse("empty-id")
	if !strings.Contains(resp, "empty-id") {
		t.Error("EmptyResponse should include the CWMP ID")
	}
}

func TestGenerateID(t *testing.T) {
	id1 := soap.GenerateID()
	id2 := soap.GenerateID()
	if id1 == "" {
		t.Error("GenerateID should return non-empty string")
	}
	if id1 == id2 {
		t.Error("consecutive GenerateID calls should return different IDs")
	}
}
