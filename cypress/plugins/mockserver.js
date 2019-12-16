const superagent = require("superagent");

class MockServer {
  constructor(port) {
    this.url = `http://localhost:${port}`;
  }

  reset = async () => {
    await superagent.put(`${this.url}/mockserver/reset`);
  };

  addMock = async ({
    request: { path, method = "GET" },
    response: { statusCode = 200, body = {} }
  }) => {
    await superagent.put(`${this.url}/mockserver/expectation`).send({
      httpRequest: {
        method,
        path
      },
      httpResponse: {
        statusCode,
        headers: {
          "content-type": ["application/json"]
        },
        body: JSON.stringify(body)
      }
    });
  };

  getRequests = async () => {
    const response = await superagent.put(
      `${this.url}/mockserver/retrieve?type=REQUESTS`
    );
    return response.body;
  };
}

module.exports = MockServer;
