#!/usr/bin/env node

// Component testing and E2E testing share the same CLI binary name, but require
// different packages. This means only one will work depending on which package
// is installed. To work around this, we can use the same trick as the
// playwright-ct.bin.js file, but instead require the CLI from the
// @playwright/test package.

require('@playwright/test/cli')
