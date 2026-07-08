package pkg

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// --------------------------------------------------------------------------
// TR-069 SOAP Envelope structures
// --------------------------------------------------------------------------

type SOAPEnvelope struct {
	XMLName xml.Name   `xml:"Envelope"`
	Header  SOAPHeader `xml:"Header"`
	Body    SOAPBody   `xml:"Body"`
}

type SOAPHeader struct {
	ID CWMPID `xml:"ID"`
}

type CWMPID struct {
	MustUnderstand string `xml:"mustUnderstand,attr"`
	Value          string `xml:",chardata"`
}

type SOAPBody struct {
	Inform                    *CWMPInform                    `xml:"Inform"`
	GetRPCMethodsResponse     *GetRPCMethodsResponse         `xml:"GetRPCMethodsResponse"`
	SetParameterValuesResponse *SetParameterValuesResponse   `xml:"SetParameterValuesResponse"`
	GetParameterValuesResponse *GetParameterValuesResponse   `xml:"GetParameterValuesResponse"`
	TransferComplete           *TransferComplete              `xml:"TransferComplete"`
	// ACS → CPE methods
	InformResponse             *InformResponse               `xml:"InformResponse"`
	GetRPCMethods              *GetRPCMethods                `xml:"GetRPCMethods"`
	Reboot                     *Reboot                       `xml:"Reboot"`
	FactoryReset               *FactoryReset                 `xml:"FactoryReset"`
	SetParameterValues         *SetParameterValues           `xml:"SetParameterValues"`
	GetParameterValues         *GetParameterValues           `xml:"GetParameterValues"`
}

// --------------------------------------------------------------------------
// CPE → ACS: Inform
// --------------------------------------------------------------------------

type CWMPInform struct {
	DeviceId    InformDeviceID `xml:"DeviceId"`
	Event       InformEvent    `xml:"Event"`
	MaxEnvelopes int           `xml:"MaxEnvelopes"`
	CurrentTime string         `xml:"CurrentTime"`
	RetryCount  int            `xml:"RetryCount"`
	ParameterList InformParamList `xml:"ParameterList"`
}

type InformDeviceID struct {
	Manufacturer string `xml:"Manufacturer"`
	OUI          string `xml:"OUI"`
	ProductClass string `xml:"ProductClass"`
	SerialNumber string `xml:"SerialNumber"`
}

type InformEvent struct {
	EventStruct []EventStruct `xml:"EventStruct"`
}

type EventStruct struct {
	EventCode  string `xml:"EventCode"`
	CommandKey string `xml:"CommandKey"`
}

type InformParamList struct {
	ParameterValueStruct []ParameterValueStruct `xml:"ParameterValueStruct"`
}

type ParameterValueStruct struct {
	Name  string `xml:"Name"`
	Value string `xml:"Value"`
}

// --------------------------------------------------------------------------
// ACS → CPE: InformResponse
// --------------------------------------------------------------------------

type InformResponse struct {
	MaxEnvelopes int `xml:"MaxEnvelopes"`
}

// --------------------------------------------------------------------------
// ACS → CPE: GetRPCMethods / GetRPCMethodsResponse
// --------------------------------------------------------------------------

type GetRPCMethods struct{}

type GetRPCMethodsResponse struct {
	MethodList MethodList `xml:"MethodList"`
}

type MethodList struct {
	Methods []string `xml:"string"`
}

// --------------------------------------------------------------------------
// ACS → CPE: Reboot / FactoryReset
// --------------------------------------------------------------------------

type Reboot struct {
	CommandKey string `xml:"CommandKey"`
}

type FactoryReset struct{}

// --------------------------------------------------------------------------
// ACS ↔ CPE: SetParameterValues
// --------------------------------------------------------------------------

type SetParameterValues struct {
	ParameterList  SetParamList `xml:"ParameterList"`
	ParameterKey   string       `xml:"ParameterKey"`
}

type SetParamList struct {
	ParameterValueStruct []ParameterValueStruct `xml:"ParameterValueStruct"`
}

type SetParameterValuesResponse struct {
	Status int `xml:"Status"`
}

// --------------------------------------------------------------------------
// ACS ↔ CPE: GetParameterValues
// --------------------------------------------------------------------------

type GetParameterValues struct {
	ParameterNames GetParamNames `xml:"ParameterNames"`
}

type GetParamNames struct {
	Names []string `xml:"string"`
}

type GetParameterValuesResponse struct {
	ParameterList InformParamList `xml:"ParameterList"`
}

// --------------------------------------------------------------------------
// TransferComplete (firmware upgrade etc.)
// --------------------------------------------------------------------------

type TransferComplete struct {
	CommandKey  string `xml:"CommandKey"`
	FaultStruct struct {
		FaultCode   int    `xml:"FaultCode"`
		FaultString string `xml:"FaultString"`
	} `xml:"FaultStruct"`
	StartTime   string `xml:"StartTime"`
	CompleteTime string `xml:"CompleteTime"`
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

// ParseEnvelope decodes a raw SOAP XML body into SOAPEnvelope.
// It tries standard CWMP namespace prefixes used by most CPE devices.
func ParseEnvelope(data []byte) (*SOAPEnvelope, error) {
	// Normalise common SOAP namespace prefixes so stdlib xml can decode
	normalized := normalizeNamespaces(string(data))
	var env SOAPEnvelope
	if err := xml.Unmarshal([]byte(normalized), &env); err != nil {
		return nil, fmt.Errorf("soap parse error: %w", err)
	}
	return &env, nil
}

// normalizeNamespaces strips namespace prefixes so encoding/xml can match
// the local element names we use in struct tags.
func normalizeNamespaces(s string) string {
	prefixes := []string{
		"SOAP-ENV:", "soapenv:", "soap:",
		"cwmp:", "xsd:", "xsi:",
	}
	for _, p := range prefixes {
		s = strings.ReplaceAll(s, p, "")
		s = strings.ReplaceAll(s, strings.ToLower(p), "")
	}
	// Remove namespace declarations too
	for strings.Contains(s, " xmlns") {
		start := strings.Index(s, " xmlns")
		if start == -1 {
			break
		}
		end := strings.Index(s[start+1:], `"`)
		if end == -1 {
			break
		}
		end2 := strings.Index(s[start+1+end+1:], `"`)
		if end2 == -1 {
			break
		}
		total := start + 1 + end + 1 + end2 + 1
		s = s[:start] + s[total:]
	}
	return s
}

// BuildInformResponse returns the SOAP InformResponse XML.
func BuildInformResponse(id string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<SOAP-ENV:Envelope
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">%s</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <cwmp:InformResponse>
      <cwmp:MaxEnvelopes>1</cwmp:MaxEnvelopes>
    </cwmp:InformResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`, id)
}

// BuildRebootRequest returns a SOAP Reboot RPC XML.
func BuildRebootRequest(id, commandKey string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<SOAP-ENV:Envelope
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">%s</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <cwmp:Reboot>
      <cwmp:CommandKey>%s</cwmp:CommandKey>
    </cwmp:Reboot>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`, id, commandKey)
}

// BuildFactoryResetRequest returns a SOAP FactoryReset RPC XML.
func BuildFactoryResetRequest(id string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<SOAP-ENV:Envelope
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">%s</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <cwmp:FactoryReset/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`, id)
}

// BuildSetParameterValuesRequest builds a SetParameterValues SOAP request.
func BuildSetParameterValuesRequest(id, paramKey string, params map[string]string) string {
	var sb strings.Builder
	for name, value := range params {
		sb.WriteString(fmt.Sprintf(`      <cwmp:ParameterValueStruct>
        <cwmp:Name>%s</cwmp:Name>
        <cwmp:Value>%s</cwmp:Value>
      </cwmp:ParameterValueStruct>
`, name, xmlEscape(value)))
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<SOAP-ENV:Envelope
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">%s</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <cwmp:SetParameterValues>
      <cwmp:ParameterList>
%s      </cwmp:ParameterList>
      <cwmp:ParameterKey>%s</cwmp:ParameterKey>
    </cwmp:SetParameterValues>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`, id, sb.String(), paramKey)
}

// BuildEmptyResponse returns a SOAP envelope with an empty body (HTTP 204-equivalent in CWMP).
func BuildEmptyResponse(id string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<SOAP-ENV:Envelope
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">%s</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body/>
</SOAP-ENV:Envelope>`, id)
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
