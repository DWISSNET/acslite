// Builder untuk membuat pesan CWMP/SOAP yang dikirim ke device
export function createSetParameterValuesRequest(parameters: Array<{name: string, value: any}>) {
  const paramList = parameters.map(p => 
    `<cwmp:ParameterValueStruct>
      <cwmp:Name>${p.name}</cwmp:Name>
      <cwmp:Value xsi:type="xsd:string">${p.value}</cwmp:Value>
    </cwmp:ParameterValueStruct>`
  ).join('');

  return `<?xml version="1.0" encoding="utf-8"?>
<SOAP-ENV:Envelope 
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
  xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">${Date.now()}</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <cwmp:SetParameterValues>
      <cwmp:ParameterList>${paramList}</cwmp:ParameterList>
      <cwmp:ParameterKey></cwmp:ParameterKey>
    </cwmp:SetParameterValues>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`;
}

export function createRebootRequest() {
  return `<?xml version="1.0" encoding="utf-8"?>
<SOAP-ENV:Envelope 
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">${Date.now()}</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <cwmp:Reboot/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`;
}

export function createFactoryResetRequest() {
  return `<?xml version="1.0" encoding="utf-8"?>
<SOAP-ENV:Envelope 
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">${Date.now()}</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <cwmp:FactoryReset/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`;
}

export function createDownloadRequest(fileType: string, fileUrl: string, filename: string) {
  return `<?xml version="1.0" encoding="utf-8"?>
<SOAP-ENV:Envelope 
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
  <SOAP-ENV:Header>
    <cwmp:ID SOAP-ENV:mustUnderstand="1">${Date.now()}</cwmp:ID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <cwmp:Download>
      <cwmp:FileType>${fileType}</cwmp:FileType>
      <cwmp:FileURL>${fileUrl}/${filename}</cwmp:FileURL>
      <cwmp:FileName>${filename}</cwmp:FileName>
      <cwmp:Username></cwmp:Username>
      <cwmp:Password></cwmp:Password>
      <cwmp:FileSize>0</cwmp:FileSize>
      <cwmp:TargetFileName></cwmp:TargetFileName>
    </cwmp:Download>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`;
}