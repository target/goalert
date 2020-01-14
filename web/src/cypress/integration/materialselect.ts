import { Chance } from 'chance'
import { testScreen } from '../support'

const c = new Chance()

testScreen('MaterialSelect', testMaterialSelect)

function testMaterialSelect(screen: ScreenFormat) {
  describe('Clearable fields', () => {
    let ep: EP
    let svc: Service

    before(() => {
      cy.createEP().then(e => {
        ep = e
      })
      cy.createService().then(e => {
        svc = e
      })
    })

    it('should clear required fields', () => {
      cy.visit(`/escalation-policies/${ep.id}`)
      const name = 'SM EP ' + c.word({ length: 7 })
      const description = c.word({ length: 9 })
      const repeat = c.integer({ min: 0, max: 5 }).toString()

      cy.pageAction('Edit Escalation Policy')
      cy.dialogTitle('Edit Escalation Policy')

      //fill name and descr, but clear repeat; expect validation error
      cy.dialogForm({ name, description, repeat: null })
      cy.dialogClick('Submit')
      cy.dialogContains('Required field')

      // fill in repeat; expect success
      cy.dialogForm({ repeat })
      cy.dialogFinish('Submit')

      // old name and descr should not be present
      cy.get('body')
        .should('not.contain', ep.name)
        .should('not.contain', ep.description)

      // new ones should
      cy.get('body')
        .should('contain', name)
        .should('contain', description)
    })

    it('should clear optional fields', () => {
      cy.visit('/services')
      const name = 'SM Svc ' + c.word({ length: 8 })
      const description = c.word({ length: 10 })

      cy.pageFab()
      cy.dialogForm({
        name,
        'escalation-policy': svc.ep.name,
        description,
      })

      cy.dialogForm({
        'escalation-policy': null,
      })
      cy.dialogFinish('Submit')

      // should be on details page
      cy.get('body')
        .should('contain', name)
        .should('contain', description)
    })
  })
}
