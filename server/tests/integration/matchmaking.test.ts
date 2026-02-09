import { describe, test, expect } from "vitest";
import {
  CLIENT_MESSAGE_TYPE,
  CONNECT_REQUEST,
  CONNECT_SCHEMA,
  MATCH_REQUEST,
  MATCH_FOUND_SCHEMA,
  QUEUE_JOINED_SCHEMA,
  SERVER_MESSAGE_TYPE,
  validateResponse,
  waitForMessage,
  WS_URL,
  WSTestClient,
} from "../helpers";

describe("Matchmaking System", () => {
  test("should add player to queue when no opponent is waiting", async () => {
    const client = new WSTestClient(WS_URL);
    await client.connect();

    // Connect first
    await client.send(CONNECT_REQUEST("Player1", [1, 2, 3]));
    const connectResponse = await waitForMessage(client);
    expect(connectResponse.type).toBe(SERVER_MESSAGE_TYPE.AcceptConnection);
    validateResponse(connectResponse.payload, CONNECT_SCHEMA);

    // Request matchmaking
    await client.send(MATCH_REQUEST());
    const queueResponse = await waitForMessage(client);

    // Should receive QueueJoined message
    expect(queueResponse.type).toBe(SERVER_MESSAGE_TYPE.QueueJoined);
    validateResponse(queueResponse.payload, QUEUE_JOINED_SCHEMA);
    expect(queueResponse.payload.queue_size).toBe(1);
    expect(queueResponse.payload.message).toContain("waiting");

    await client.close();
  });

  test("should match two players when both request matchmaking", async () => {
    const client1 = new WSTestClient(WS_URL);
    const client2 = new WSTestClient(WS_URL);

    await Promise.all([client1.connect(), client2.connect()]);

    // Connect both players
    await client1.send(CONNECT_REQUEST("Player1", [1, 2, 3]));
    await client2.send(CONNECT_REQUEST("Player2", [4, 5, 6]));

    const [connect1, connect2] = await Promise.all([
      waitForMessage(client1),
      waitForMessage(client2),
    ]);

    expect(connect1.type).toBe(SERVER_MESSAGE_TYPE.AcceptConnection);
    expect(connect2.type).toBe(SERVER_MESSAGE_TYPE.AcceptConnection);

    // First player enters queue
    await client1.send(MATCH_REQUEST());
    const queue1 = await waitForMessage(client1);
    expect(queue1.type).toBe(SERVER_MESSAGE_TYPE.QueueJoined);
    expect(queue1.payload.queue_size).toBe(1);

    // Second player enters queue - should trigger match
    await client2.send(MATCH_REQUEST());

    const [match1, match2] = await Promise.all([
      waitForMessage(client1),
      waitForMessage(client2),
    ]);

    // Both players should receive MatchFound
    expect(match1.type).toBe(SERVER_MESSAGE_TYPE.MatchFound);
    expect(match2.type).toBe(SERVER_MESSAGE_TYPE.MatchFound);

    validateResponse(match1.payload, MATCH_FOUND_SCHEMA);
    validateResponse(match2.payload, MATCH_FOUND_SCHEMA);

    // Verify battle IDs match
    expect(match1.payload.battle_id).toBe(match2.payload.battle_id);

    // Player1 should receive correct info
    expect(match1.payload.your_info.username).toBe("Player1");
    expect(match1.payload.opponent_info.username).toBe("Player2");
    expect(match1.payload.your_info.team).toHaveLength(3);
    expect(match1.payload.opponent_info.team).toHaveLength(3);

    // Player2 should receive correct info
    expect(match2.payload.your_info.username).toBe("Player2");
    expect(match2.payload.opponent_info.username).toBe("Player1");
    expect(match2.payload.your_info.team).toHaveLength(3);
    expect(match2.payload.opponent_info.team).toHaveLength(3);

    await Promise.all([client1.close(), client2.close()]);
  });

  test("should handle multiple players in queue (FIFO)", async () => {
    const client1 = new WSTestClient(WS_URL);
    const client2 = new WSTestClient(WS_URL);
    const client3 = new WSTestClient(WS_URL);

    await Promise.all([
      client1.connect(),
      client2.connect(),
      client3.connect(),
    ]);

    // Connect all players
    await Promise.all([
      client1.send(CONNECT_REQUEST("Player1", [1, 2, 3])),
      client2.send(CONNECT_REQUEST("Player2", [4, 5, 6])),
      client3.send(CONNECT_REQUEST("Player3", [7, 8, 9])),
    ]);

    // Wait for all connections
    await Promise.all([
      waitForMessage(client1),
      waitForMessage(client2),
      waitForMessage(client3),
    ]);

    // Player1 enters queue
    await client1.send(MATCH_REQUEST());
    const queue1 = await waitForMessage(client1);
    expect(queue1.type).toBe(SERVER_MESSAGE_TYPE.QueueJoined);

    // Player2 enters queue - should match with Player1
    await client2.send(MATCH_REQUEST());
    const [match1, match2] = await Promise.all([
      waitForMessage(client1),
      waitForMessage(client2),
    ]);

    expect(match1.type).toBe(SERVER_MESSAGE_TYPE.MatchFound);
    expect(match2.type).toBe(SERVER_MESSAGE_TYPE.MatchFound);
    expect(match1.payload.opponent_info.username).toBe("Player2");
    expect(match2.payload.opponent_info.username).toBe("Player1");

    // Player3 enters queue - should be alone in queue
    await client3.send(MATCH_REQUEST());
    const queue3 = await waitForMessage(client3);
    expect(queue3.type).toBe(SERVER_MESSAGE_TYPE.QueueJoined);
    expect(queue3.payload.queue_size).toBe(1);

    await Promise.all([client1.close(), client2.close(), client3.close()]);
  });

  test("should remove player from queue on disconnect", async () => {
    const client1 = new WSTestClient(WS_URL);
    const client2 = new WSTestClient(WS_URL);

    await Promise.all([client1.connect(), client2.connect()]);

    // Connect both players
    await client1.send(CONNECT_REQUEST("Player1", [1, 2, 3]));
    await client2.send(CONNECT_REQUEST("Player2", [4, 5, 6]));

    await Promise.all([waitForMessage(client1), waitForMessage(client2)]);

    // Player1 enters queue
    await client1.send(MATCH_REQUEST());
    const queue1 = await waitForMessage(client1);
    expect(queue1.type).toBe(SERVER_MESSAGE_TYPE.QueueJoined);

    // Player1 disconnects (should be removed from queue)
    await client1.close();

    // Small delay to ensure cleanup happens
    await new Promise((resolve) => setTimeout(resolve, 100));

    // Player2 enters queue - should be alone (not matched with disconnected Player1)
    await client2.send(MATCH_REQUEST());
    const queue2 = await waitForMessage(client2);
    expect(queue2.type).toBe(SERVER_MESSAGE_TYPE.QueueJoined);
    expect(queue2.payload.queue_size).toBe(1);

    await client2.close();
  });

  test("should handle rapid matchmaking requests", async () => {
    const clients = Array.from({ length: 4 }, () => new WSTestClient(WS_URL));

    // Connect all clients
    await Promise.all(clients.map((c) => c.connect()));

    // Send connection requests
    await Promise.all(
      clients.map((c, i) =>
        c.send(CONNECT_REQUEST(`Player${i + 1}`, [1, 2, 3]))
      )
    );

    // Wait for all connections
    await Promise.all(clients.map((c) => waitForMessage(c)));

    // All players request matchmaking simultaneously
    await Promise.all(clients.map((c) => c.send(MATCH_REQUEST())));

    // Collect all responses
    const responses = await Promise.all(clients.map((c) => waitForMessage(c)));

    // Count matches and queued
    const matched = responses.filter(
      (r) => r.type === SERVER_MESSAGE_TYPE.MatchFound
    );
    const queued = responses.filter(
      (r) => r.type === SERVER_MESSAGE_TYPE.QueueJoined
    );

    // With 4 players, we should have 2 matches (4 MatchFound messages) or some in queue
    // The exact outcome depends on timing, but we should have valid responses
    expect(matched.length + queued.length).toBe(4);

    // If there are matches, validate their structure
    matched.forEach((response) => {
      validateResponse(response.payload, MATCH_FOUND_SCHEMA);
    });

    queued.forEach((response) => {
      validateResponse(response.payload, QUEUE_JOINED_SCHEMA);
    });

    await Promise.all(clients.map((c) => c.close()));
  });
});