const MockServer = require("./mockserver");

class MockResend {
  constructor(port) {
    this.mockserver = new MockServer(port);
  }

  listenForEmails = async () => {
    await this.mockserver.reset();
    await this.mockserver.addMock({
      request: {
        path: "/emails",
        method: "POST"
      },
      response: {
        statusCode: 200
      }
    });
    return null;
  };

  getEmails = async () => {
    const requests = await this.mockserver.getRequests();
    const requestBodies = requests.map(request => request.body.json);
    return requestBodies;
  };
}

module.exports = MockResend;
