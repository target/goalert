#!/usr/bin/env node

// Component testing and E2E testing share the same CLI binary name, but require
// different packages. This means only one will work depending on which package
// is installed. To work around this, we have two binaries, one for each package,
// and we require the correct one based on which set of tests we're running.

// Component testing uses the "playwright" package.

require('playwright/cli')
