# HTTP API – WhatsApp Connect

## `GET /api/connect/whatsapp`

Description: Returns whether a WhatsApp session is currently connected.

Request:

- Empty.

Response:

- `connected`: Boolean indicating if a session exists.

## `DELETE /api/connect/whatsapp`

Description: Removes the stored WhatsApp session and history cache.

Request:

- Empty.

Response:

- Empty with status `204 No Content`.

## `GET /api/connect/whatsapp?connect=true`

Description: Starts the WhatsApp pairing flow and streams QR data via Server-Sent Events (SSE).

Request:

- `connect=true`: Required to start pairing.

Response (SSE data events):

- `{"code":"<base64 PNG>","timeout":<ms>}`: QR code image (base64-encoded PNG) and timeout in milliseconds.
- `{"scanned": true}`: Emitted after the QR code is successfully scanned.
- `{"error":"<message>"}`: Emitted on errors before the stream closes.

Behavior:

- If a session already exists, the endpoint returns `400`.
- The stream remains open for up to 5 minutes or until the client disconnects.
