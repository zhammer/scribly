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

  getEmails = async () => {
    const requests = await this.mockserver.getRequests();
    const requestBodies = requests.map(request =>
      JSON.parse(request.body.string)
    );
    return requestBodies;
  };
}

module.exports = MockSendGrid;
