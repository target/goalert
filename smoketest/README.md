# Notification Smoketest Suite

A suite of tests to check that notifications get sent properly.

## Setup

1. Ensure you have postgres running locally, the test suite will create timestamped databases while running.
1. Make sure you have the `goalert` binary installed. (run `make install` from the root of the repo)
1. If you want codesigning (Mac OS, prevents the firewall popup) create a trusted cert named `localdev` (detailed steps are below).

## Running Tests

Run `make smoketest` from the root of the repo to run all tests.
The script will automatically rebuild goalert, including any change migrations.

To run a single test you can use `go test` from the smoketest directory.

## Creating Tests

- Try to keep the test under 5 minutes, if possible.
- Specify the current migration level when the test is created. This way your initial SQL won't break as migrations are applied in the future, and your test will ensure behavior against newer migrations at the same time.
- Make sure to call `t.Parallel()` and `defer h.Close()` in your test.

### Creating a Code Signing Cert (Mac OS Only)

1. Open `Keychain Access` (press Command+Space and start typing the name)
1. Click the menu `Keychain Access` -> `Certificate Assistant` -> `Create a Certificate ...`
1. Enter `localdev` as the name.
1. Set "Certificate Type" to `Code Signing`
1. Click continue, hit yes on the confirmation window.
1. Go to the `login` keychain and find the new cert.
1. Right-click (or control-click) and select `Get Info`.
1. Click the arrow next to `Trust`.
1. Next to "When using this certificate" select `Always Trust`.
