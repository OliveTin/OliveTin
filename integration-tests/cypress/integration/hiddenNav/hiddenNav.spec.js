describe('Hidden Nav', () => {
  beforeEach(() => {
    cy.visit("/")
  });

  it("Footer contains promo", () => {
    cy.get('footer').contains("OliveTin")
  })

  it('Switcher navigation is hidden', () => {
    cy.get('#switcher').then($el => {
      expect(Cypress.dom.isHidden($el)).to.be.true
    })
  })
});


