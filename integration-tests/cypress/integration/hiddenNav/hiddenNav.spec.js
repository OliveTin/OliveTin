describe('Hidden Nav', () => {
  beforeEach(() => {
    cy.visit("/")
    cy.wait(500)
  });

  it("Footer contains promo", () => {
    cy.get('footer').contains("OliveTin")
  })

  it('Switcher navigation is hidden', () => {
    cy.get('#section-switcher').then($el => {
      expect(Cypress.dom.isHidden($el)).to.be.true
    })
  })
});
