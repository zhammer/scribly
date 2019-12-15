const MockServer = require("./mockserver");

class MockSendGrid {
  constructor(port) {
    this.mockserver = new MockServer(port);
  }

  listenForEmails = async () => {
    await this.mockserver.reset();
    await this.mockserver.addMock({
      request: {
        path: "/v3/mail/send",
        method: "POST"
      },
      response: {
        statusCode: 202
      }
    });
    return null;
  };
}

module.exports = MockSendGrid;
