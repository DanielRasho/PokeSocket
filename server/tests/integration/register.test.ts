import { describe, test, expect, beforeEach, afterEach } from "vitest";
import {
  CLIENT_MESSAGE_TYPE,
  CONNECT_REQUEST,
  CONNECT_SCHEMA,
  createMessage,
  SERVER_MESSAGE_TYPE,
  validateResponse,
  waitForMessage,
  WS_URL,
  WSTestClient,
} from "../helpers";
import { object, string, number, array, bool, Schema } from "yup";

describe("Testing Basic connections", () => {

  test("should handle graceful disconnection", async () => {
    const client = new WSTestClient(WS_URL);
    await client.connect();

    await client.send(CONNECT_REQUEST("persona 1", [1, 2, 3]),);

    const response = await waitForMessage(client);
    expect(response.type).toBe(SERVER_MESSAGE_TYPE.AcceptConnection);

    // Close connection gracefully
    await client.close();
    expect(client.isOpen()).toBe(false);
  });

  test("should support concurrent connections", async () => {
    const client1 = new WSTestClient(WS_URL);
    const client2 = new WSTestClient(WS_URL);


    await Promise.all([client1.connect(), client2.connect()]);

    client1.send(CONNECT_REQUEST("persona 1", [1, 2, 3]),);
    client2.send(CONNECT_REQUEST("persona 2", [1, 2, 3]),);

    const [response1, response2] = await Promise.all([
      waitForMessage(client1),
      waitForMessage(client2),
    ]);
    
    validateResponse(response1.payload, CONNECT_SCHEMA)
    validateResponse(response2.payload, CONNECT_SCHEMA)

    await Promise.all([client1.close(), client2.close()]);
  });
});
