describe('Homepage rendering', () => {
  beforeEach(() => {
    cy.visit("/")
  });

  it("Footer contains promo", () => {
    cy.get('footer').contains("OliveTin")
  })

  it('Default buttons are rendered', () => {
    cy.get("#rootGroup button").should('have.length', 6)
  })

  it('Switcher navigation is visible', () => {
    cy.get('#sectionSwitcher').then($el => {
      expect(Cypress.dom.isHidden($el)).to.be.false
    })
  })
});


