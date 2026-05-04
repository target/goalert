import { Chance } from 'chance'
import { testScreen } from '../support/e2e'
const c = new Chance()

function testIMAPMonitoring(screen: ScreenFormat): void {
  describe('IMAP Email Monitoring', () => {
    let svc: Service

    beforeEach(() => {
      cy.createService().then((s: Service) => {
        svc = s
      })
    })

    it('should configure IMAP and manage filter rules', () => {
      const username = `test-${c.word({ length: 8 })}@example.com`
      const filterName1 = 'Error ' + c.word({ length: 5 })
      const filterName2 = 'Critical ' + c.word({ length: 5 })

      // Navigate to service IMAP page
      cy.visit(`/services/${svc.id}`)
      cy.get('body').should('contain', svc.name)

      // Click IMAP Email Monitoring link
      cy.get('a[href*="imap-email-monitoring"]')
        .should('contain', 'IMAP Email Monitoring')
        .click()

      cy.url().should('include', '/imap-email-monitoring')

      // Should show empty state initially
      cy.get('body').should('contain', 'No IMAP configuration')

      // Configure IMAP
      cy.get('button').contains('Configure IMAP').click()

      // Fill in IMAP configuration form
      cy.dialogForm({
        enabled: true,
        username,
        host: 'imap.gmail.com',
        port: '993',
        mailbox: 'INBOX',
        pollIntervalMinutes: '5',
        useTLS: true,
        markAsRead: false,
        deleteAfter: false,
        includeHeaders: false,
        includeFrom: true,
        includeTo: true,
        includeSubject: true,
        includeBody: true,
        oauthClientID: 'test-client-id',
        oauthClientSecret: 'test-client-secret',
        oauthRefreshToken: 'test-refresh-token',
      })

      cy.dialogFinish('Submit')

      // Verify IMAP configuration is displayed
      cy.get('body').should('contain', username)
      cy.get('body').should('contain', 'imap.gmail.com')
      cy.get('body').should('contain', 'INBOX')

      // Create first filter rule
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button[aria-label="Create Filter Rule"]').click()
      }

      cy.dialogForm({
        name: filterName1,
        subjectPattern: 'ERROR',
        matchMode: 'contains',
        excludeReplies: true,
      })

      cy.dialogFinish('Submit')

      // Verify filter rule is displayed
      cy.get('body').should('contain', filterName1)
      cy.get('body').should('contain', 'ERROR')
      cy.get('body').should('contain', 'contains')

      // Create second filter rule with from pattern
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button[aria-label="Create Filter Rule"]').click()
      }

      cy.dialogForm({
        name: filterName2,
        fromPattern: 'alerts@example.com',
        subjectPattern: 'CRITICAL',
        matchMode: 'exact',
        excludeReplies: false,
      })

      cy.dialogFinish('Submit')

      // Verify second filter rule
      cy.get('body').should('contain', filterName2)
      cy.get('body').should('contain', 'alerts@example.com')
      cy.get('body').should('contain', 'CRITICAL')

      // Edit first filter rule
      cy.get(`[data-cy="card-list-item"]`)
        .contains(filterName1)
        .parent()
        .parent()
        .parent()
        .find('button[aria-label="Other Actions"]')
        .click()

      cy.get('[role="menuitem"]').contains('Edit').click()

      // Update the filter rule
      cy.dialogForm({
        subjectPattern: 'CRITICAL ERROR',
        matchMode: 'regex',
      })

      cy.dialogFinish('Submit')

      // Verify the update
      cy.get('body').should('contain', filterName1)
      cy.get('body').should('contain', 'CRITICAL ERROR')
      cy.get('body').should('contain', 'regex')

      // Disable filter rule
      cy.get(`[data-cy="card-list-item"]`)
        .contains(filterName1)
        .parent()
        .parent()
        .parent()
        .find('button[aria-label="Other Actions"]')
        .click()

      cy.get('[role="menuitem"]').contains('Edit').click()

      cy.dialogForm({
        enabled: false,
      })

      cy.dialogFinish('Submit')

      // Delete second filter rule
      cy.get(`[data-cy="card-list-item"]`)
        .contains(filterName2)
        .parent()
        .parent()
        .parent()
        .find('button[aria-label="Other Actions"]')
        .click()

      cy.get('[role="menuitem"]').contains('Delete').click()

      cy.dialogTitle('Are you sure?')
      cy.dialogContains(filterName2)
      cy.dialogFinish('Confirm')

      // Verify deletion
      cy.get('body').should('not.contain', filterName2)
      cy.get('body').should('contain', filterName1) // First rule should still exist

      // Edit IMAP configuration
      cy.get('button').contains('Edit Configuration').click()

      cy.dialogForm({
        pollIntervalMinutes: '10',
        markAsRead: true,
      })

      cy.dialogFinish('Submit')

      // Verify configuration update
      cy.get('body').should('contain', '10 minutes')

      // Disable IMAP
      cy.get('button').contains('Edit Configuration').click()

      cy.dialogForm({
        enabled: false,
      })

      cy.dialogFinish('Submit')

      // Verify IMAP is disabled
      cy.get('body').should('contain', 'IMAP monitoring is currently disabled')
    })

    it('should validate filter rule form', () => {
      const username = `test-${c.word({ length: 8 })}@example.com`

      // Navigate to service IMAP page
      cy.visit(`/services/${svc.id}/imap-email-monitoring`)

      // Configure IMAP first
      cy.get('button').contains('Configure IMAP').click()

      cy.dialogForm({
        enabled: true,
        username,
        oauthClientID: 'test-client-id',
        oauthClientSecret: 'test-client-secret',
        oauthRefreshToken: 'test-refresh-token',
      })

      cy.dialogFinish('Submit')

      // Try to create filter rule without name
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button[aria-label="Create Filter Rule"]').click()
      }

      // Submit without filling required fields
      cy.dialogFinish('Submit')

      // Should show validation error
      cy.get('body').should('contain', 'Required')

      // Cancel dialog
      cy.dialogFinish('Cancel')

      // Try to create filter rule without any pattern
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button[aria-label="Create Filter Rule"]').click()
      }

      cy.dialogForm({
        name: 'Invalid Rule',
      })

      cy.dialogFinish('Submit')

      // Should show validation error for missing patterns
      cy.get('body').should(
        'contain',
        'At least one pattern (From, Subject, or To) must be provided',
      )

      cy.dialogFinish('Cancel')
    })

    it('should test regex pattern validation', () => {
      const username = `test-${c.word({ length: 8 })}@example.com`

      // Navigate to service IMAP page
      cy.visit(`/services/${svc.id}/imap-email-monitoring`)

      // Configure IMAP
      cy.get('button').contains('Configure IMAP').click()

      cy.dialogForm({
        enabled: true,
        username,
        oauthClientID: 'test-client-id',
        oauthClientSecret: 'test-client-secret',
        oauthRefreshToken: 'test-refresh-token',
      })

      cy.dialogFinish('Submit')

      // Create filter rule with regex pattern
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button[aria-label="Create Filter Rule"]').click()
      }

      cy.dialogForm({
        name: 'Regex Test',
        subjectPattern: '(ERROR|CRITICAL|URGENT)',
        matchMode: 'regex',
        excludeReplies: true,
      })

      cy.dialogFinish('Submit')

      // Verify regex pattern is saved
      cy.get('body').should('contain', 'Regex Test')
      cy.get('body').should('contain', '(ERROR|CRITICAL|URGENT)')
      cy.get('body').should('contain', 'regex')
    })

    it('should handle OAuth credentials', () => {
      const username = `test-${c.word({ length: 8 })}@example.com`

      // Navigate to service IMAP page
      cy.visit(`/services/${svc.id}/imap-email-monitoring`)

      // Configure IMAP with OAuth
      cy.get('button').contains('Configure IMAP').click()

      // Fill OAuth fields
      cy.dialogForm({
        enabled: true,
        username,
        oauthClientID: 'test-client-id-123',
        oauthClientSecret: 'test-client-secret-456',
      })

      // Verify "Get OAuth Refresh Token" button is enabled when credentials are filled
      cy.get('button')
        .contains('Get OAuth Refresh Token')
        .should('not.be.disabled')

      // Add refresh token
      cy.dialogForm({
        oauthRefreshToken: 'test-refresh-token-789',
      })

      cy.dialogFinish('Submit')

      // Verify configuration saved
      cy.get('body').should('contain', username)
    })
  })
}

testScreen('IMAP Email Monitoring', testIMAPMonitoring)
