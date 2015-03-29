# UploadSession Resource

Provides information about a [large file upload](../items/upload_large_files.md)
session.

### JSON representation
<!-- { "blockType": "resource", "@odata.type": "oneDrive.uploadSession" } -->
```json
{
  "uploadUrl": "string",
  "expirationDateTime": "string (timestamp)",
  "nextExpectedRanges": ["string"]
}
```

| Property Name        | Value                               | Description                                                                                                     |
|:---------------------|:------------------------------------|:----------------------------------------------------------------------------------------------------------------|
| `uploadUrl`          | string                              | URL where fragment PUT requests should be directed.                                                             |
| `expirationDateTime` | [timestamp](../facets/timestamp.md) | Date and time when the upload session expires.                                                                  |
| `nextExpectedRanges` | string array                        | An array of byte ranges the server is missing. Not always a full list of the missing ranges.                    |
