# SendIt

SendIt is a tool for setting up a reverse proxy to localhost that's accesible externally. It's similar to tools like `ngrok` and `serveo`.

It was created to aid in running/testing GoAlert integrations.

## Server

Usage: `sendit-server -secret <SECRET_STRING>`

If exposing publicly, it's recommended to start in secure mode with the `-secret` flag set. This will require clients to provide a token generated with the `sendit-token` command.

## Client

Usage: `sendit -token <TOKEN> <SERVER_URL>/<DESIRED_PREFIX> <LOCAL_URL>`

Example: `sendit https://example.com/foobar http://localhost:3030`

Make sure GoAlert is started with the appropriate prefix. For the above example: `make start GOALERT_HTTP_PREFIX=/foobar`

I you are testing Twilio functionality, you will also need to set your `General.PublicURL` config to the source URL (example above: `https://example.com/foobar`)

## Generate Tokens

Usage: `sendit-token -secret <SECRET_STRING>`

The `sendit-token` command will generate a token that will work with a server that has the provided `SECRET_STRING` set as it's `-secret` parameter.

### How It Works

A session is established on the `open` endpoint reserving a desired prefix, and providing the client with a token to establish two HTTP streams.

- A POST request is made to `read` which the client will use by reading the _response_ body
- A POST request is made to `write` which the client will use by writing the _request_ body

With both read and write the two requests are used together to provide a bi-directional stream. On top of said stream, a multiplexing library is used that allows multiple "streams" to be emulated within a single `ReadWriteCloser`. New HTTP requests on the server end can then be proxied back to the client, and from the client to the target local URL.

After the `max-ttl` is reached, the client will initiate new `read` and `write` requests to the server, using the existing token. A 4-way handshake is used to transition traffic from the old requests to the new. Afterwards the old requests complete normally and the new ones carry all traffic. Only one request is used for each half of the "pipe" at a time.

The `max-ttl` can be adjusted lower, for environments that have a shorter request timeouts than the default of 15 seconds.
