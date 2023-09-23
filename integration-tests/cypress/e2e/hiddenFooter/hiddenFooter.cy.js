describe('Hidden Footer', () => {
  beforeEach(() => {
    cy.visit("/")
    cy.wait(500)
  });

  it('Footer is hidden', () => {
    cy.get('footer').then($el => {
      expect(Cypress.dom.isHidden($el)).to.be.true
    })
  })
});
