import { test, expect } from '@playwright/test'
import { userSessionFile } from './lib'

test.describe.configure({ mode: 'parallel' })
test.use({ storageState: userSessionFile })
