describe('Homepage rendering', () => {
  beforeEach(() => {
    cy.visit("/")
  });

  it("Footer contains promo", () => {
    cy.get('footer').contains("OliveTin")
  })

  it('Default buttons are rendered', () => {
    cy.get("#root-group button").should('have.length', 8)
  })

  /*
  it('Switcher navigation is visible', () => {
    cy.get('#section-switcher').then($el => {
      expect(Cypress.dom.isHidden($el)).to.be.false
    })
  })
  */
});
