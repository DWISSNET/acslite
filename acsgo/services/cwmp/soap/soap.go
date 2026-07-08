// Package soap provides SOAP/CWMP XML marshaling and unmarshaling for TR-069.
package soap

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

// Namespace constants
const (
	NSEnvelope = "http://schemas.xmlsoap.org/soap/envelope/"
	NSEncoding = "http://schemas.xmlsoap.org/soap/encoding/"
	NSCWMP     = "urn:dslforum-org:cwmp-1-0"
	NSXSDInt   = "xsd:int"
	NSXSDStr   = "xsd:string"
	NSXSDBool  = "xsd:boolean"
)

// ---- Incoming SOAP Structures ----

// Envelope is the root SOAP envelope.
type Envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Header  Header   `xml:"Header"`
	Body    Body     `xml:"Body"`
}

// Header contains the CWMP ID.
type Header struct {
	ID             string `xml:"ID"`
	HoldRequests   string `xml:"HoldRequests,omitempty"`
	NoMoreRequests string `xml:"NoMoreRequests,omitempty"`
}

// Body is the SOAP body containing an RPC method.
type Body struct {
	// Incoming
	Inform                    *Inform                    `xml:"Inform"`
	GetRPCMethods             *GetRPCMethods             `xml:"GetRPCMethods"`
	GetParameterValuesResponse *GetParameterValuesResponse `xml:"GetParameterValuesResponse"`
	SetParameterValuesResponse *SetParameterValuesResponse `xml:"SetParameterValuesResponse"`
	GetParameterNamesResponse  *GetParameterNamesResponse  `xml:"GetParameterNamesResponse"`
	RebootResponse            *RebootResponse            `xml:"RebootResponse"`
	FactoryResetResponse      *FactoryResetResponse      `xml:"FactoryResetResponse"`
	TransferCompleteRequest   *TransferComplete           `xml:"TransferComplete"`
	Fault                     *Fault                     `xml:"Fault"`
}

// DeviceID holds the TR-069 device identification.
type DeviceID struct {
	Manufacturer string `xml:"Manufacturer"`
	OUI          string `xml:"OUI"`
	ProductClass string `xml:"ProductClass"`
	SerialNumber string `xml:"SerialNumber"`
}

// EventStruct holds a single CWMP event code and command key.
type EventStruct struct {
	EventCode  string `xml:"EventCode"`
	CommandKey string `xml:"CommandKey"`
}

// ParameterValue holds a single TR-069 parameter name/value pair.
type ParameterValue struct {
	Name  string `xml:"Name"`
	Value string `xml:"Value"`
}

// Inform is the CWMP Inform RPC from CPE to ACS.
type Inform struct {
	DeviceId         DeviceID      `xml:"DeviceId"`
	Event            []EventStruct `xml:"Event>EventStruct"`
	MaxEnvelopes     int           `xml:"MaxEnvelopes"`
	CurrentTime      string        `xml:"CurrentTime"`
	RetryCount       int           `xml:"RetryCount"`
	ParameterList    []ParameterValue `xml:"ParameterList>ParameterValueStruct"`
}

// GetRPCMethods is the CWMP GetRPCMethods request.
type GetRPCMethods struct{}

// GetParameterValuesResponse contains values returned by the CPE.
type GetParameterValuesResponse struct {
	ParameterList []ParameterValue `xml:"ParameterList>ParameterValueStruct"`
}

// SetParameterValuesResponse is returned by CPE after SetParameterValues.
type SetParameterValuesResponse struct {
	Status int `xml:"Status"`
}

// GetParameterNamesResponse lists available parameter names.
type GetParameterNamesResponse struct {
	ParameterList []struct {
		Name     string `xml:"Name"`
		Writable bool   `xml:"Writable"`
	} `xml:"ParameterList>ParameterInfoStruct"`
}

// RebootResponse is the CPE's acknowledgement of a Reboot command.
type RebootResponse struct{}

// FactoryResetResponse is the CPE's acknowledgement of a FactoryReset command.
type FactoryResetResponse struct{}

// TransferComplete is the CPE notification that a firmware download completed.
type TransferComplete struct {
	CommandKey  string `xml:"CommandKey"`
	FaultStruct struct {
		FaultCode   int    `xml:"FaultCode"`
		FaultString string `xml:"FaultString"`
	} `xml:"FaultStruct"`
	StartTime   string `xml:"StartTime"`
	CompleteTime string `xml:"CompleteTime"`
}

// Fault represents a SOAP fault from the CPE.
type Fault struct {
	FaultCode   string `xml:"faultcode"`
	FaultString string `xml:"faultstring"`
	Detail      struct {
		CWMPFault struct {
			FaultCode   int    `xml:"FaultCode"`
			FaultString string `xml:"FaultString"`
		} `xml:"cwmp:Fault"`
	} `xml:"detail"`
}

// ---- Outgoing SOAP Builders ----

// InformResponse generates the SOAP response to a CPE Inform.
func InformResponse(cwmpID string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
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
</SOAP-ENV:Envelope>`, cwmpID)
}

// GetRPCMethodsResponse lists the ACS-supported RPC methods.
func GetRPCMethodsResponse(cwmpID string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">%s</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <cwmp:GetRPCMethodsResponse>
      <MethodList SOAP-ENC:arrayType="xsd:string[8]"
        xmlns:SOAP-ENC="http://schemas.xmlsoap.org/soap/encoding/">
        <string>GetRPCMethods</string>
        <string>SetParameterValues</string>
        <string>GetParameterValues</string>
        <string>GetParameterNames</string>
        <string>Reboot</string>
        <string>FactoryReset</string>
        <string>Download</string>
        <string>Upload</string>
      </MethodList>
    </cwmp:GetRPCMethodsResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`, cwmpID)
}

// RebootRequest builds a SOAP Reboot RPC.
func RebootRequest(cwmpID, commandKey string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
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
</SOAP-ENV:Envelope>`, cwmpID, commandKey)
}

// FactoryResetRequest builds a SOAP FactoryReset RPC.
func FactoryResetRequest(cwmpID string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">%s</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <cwmp:FactoryReset/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`, cwmpID)
}

// SetParameterValuesRequest builds a SOAP SetParameterValues RPC.
func SetParameterValuesRequest(cwmpID, commandKey string, params map[string]string) string {
	var sb strings.Builder
	for name, value := range params {
		sb.WriteString(fmt.Sprintf(`
        <ParameterValueStruct>
          <Name>%s</Name>
          <Value xsi:type="xsd:string">%s</Value>
        </ParameterValueStruct>`, xmlEscape(name), xmlEscape(value)))
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:SOAP-ENC="http://schemas.xmlsoap.org/soap/encoding/"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
  xmlns:xsd="http://www.w3.org/2001/XMLSchema"
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">%s</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <cwmp:SetParameterValues>
      <ParameterList SOAP-ENC:arrayType="cwmp:ParameterValueStruct[%d]">%s
      </ParameterList>
      <ParameterKey>%s</ParameterKey>
    </cwmp:SetParameterValues>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`, cwmpID, len(params), sb.String(), commandKey)
}

// GetParameterValuesRequest builds a SOAP GetParameterValues RPC.
func GetParameterValuesRequest(cwmpID string, paths []string) string {
	var sb strings.Builder
	for _, p := range paths {
		sb.WriteString(fmt.Sprintf("        <string>%s</string>\n", xmlEscape(p)))
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:SOAP-ENC="http://schemas.xmlsoap.org/soap/encoding/"
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">%s</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <cwmp:GetParameterValues>
      <ParameterNames SOAP-ENC:arrayType="xsd:string[%d]">
%s      </ParameterNames>
    </cwmp:GetParameterValues>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`, cwmpID, len(paths), sb.String())
}

// DownloadRequest builds a SOAP Download (firmware update) RPC.
func DownloadRequest(cwmpID, commandKey, fileType, url, username, password, fileSize, targetFileName string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">%s</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <cwmp:Download>
      <CommandKey>%s</CommandKey>
      <FileType>%s</FileType>
      <URL>%s</URL>
      <Username>%s</Username>
      <Password>%s</Password>
      <FileSize>%s</FileSize>
      <TargetFileName>%s</TargetFileName>
      <DelaySeconds>0</DelaySeconds>
    </cwmp:Download>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`, cwmpID, commandKey, fileType, url, username, password, fileSize, targetFileName)
}

// EmptyResponse generates an empty SOAP body (ACS has no pending RPC).
func EmptyResponse(cwmpID string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">%s</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body/>
</SOAP-ENV:Envelope>`, cwmpID)
}

// FaultResponse generates a SOAP fault response.
func FaultResponse(cwmpID, code, text string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">%s</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <SOAP-ENV:Fault>
      <faultcode>%s</faultcode>
      <faultstring>%s</faultstring>
    </SOAP-ENV:Fault>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`, cwmpID, code, text)
}

// TransferCompleteResponse acknowledges a TransferComplete notification.
func TransferCompleteResponse(cwmpID string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">%s</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <cwmp:TransferCompleteResponse/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`, cwmpID)
}

// ParseEnvelope parses raw SOAP XML into an Envelope struct.
func ParseEnvelope(data []byte) (*Envelope, error) {
	// Strip namespace prefixes that xml.Unmarshal may choke on
	normalized := normalizeNamespaces(data)
	var env Envelope
	if err := xml.Unmarshal(normalized, &env); err != nil {
		return nil, fmt.Errorf("parse SOAP: %w", err)
	}
	return &env, nil
}

// GenerateID creates a unique CWMP message ID.
func GenerateID() string {
	return fmt.Sprintf("ACS%d", time.Now().UnixNano())
}

// normalizeNamespaces strips namespace prefixes so xml.Unmarshal handles
// elements like <cwmp:Inform> as just <Inform>.
func normalizeNamespaces(data []byte) []byte {
	s := string(data)
	// Replace namespace-prefixed element names
	prefixes := []string{"cwmp:", "SOAP-ENV:", "SOAP-ENC:", "soap:", "soapenv:", "xsi:"}
	for _, p := range prefixes {
		s = strings.ReplaceAll(s, "<"+p, "<")
		s = strings.ReplaceAll(s, "</"+p, "</")
	}
	return []byte(s)
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
