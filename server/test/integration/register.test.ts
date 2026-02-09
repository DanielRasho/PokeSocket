import { describe, it, expect, beforeEach, afterEach } from "vitest";
import {
  CLIENT_MESSAGE_TYPE,
  createMessage,
  SERVER_MESSAGE_TYPE,
  waitForMessage,
  WS_URL,
  WSTestClient,
} from "../helpers";

describe("WebSocket Connection Tests", () => {
  let client: WSTestClient;

  beforeEach(() => {
    client = new WSTestClient(WS_URL);
  });

  afterEach(async () => {
    if (client.isOpen()) {
      await client.close();
    }
  });

  it("should connect to the server", async () => {
    await client.connect();
    expect(client.isOpen()).toBe(true);
  });

  it("should successfully connect with valid username", async () => {
    await client.connect();

    const connectRequest = {
      username: "TestPlayer",
    };

    await client.send(
      createMessage(CLIENT_MESSAGE_TYPE.Connect, connectRequest),
    );

    const response = await waitForMessage(client);

    expect(response.type).toBe(SERVER_MESSAGE_TYPE.AcceptConnection);
    expect(response.payload.username).toBe("TestPlayer");
    expect(response.payload.uuid).toBeDefined();
    expect(response.payload.uuid).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
    );
  });

  it("should reject connection without username", async () => {
    await client.connect();

    const invalidRequest = {
      // Missing username field
    };

    await client.send(
      createMessage(CLIENT_MESSAGE_TYPE.Connect, invalidRequest),
    );

    const response = await waitForMessage(client);

    expect(response.type).toBe(SERVER_MESSAGE_TYPE.Error);
    expect(response.payload.code).toBeDefined();
    expect(response.payload.msg).toBeDefined();
  });

  it("should reject duplicate connection attempts", async () => {
    await client.connect();

    const connectRequest = {
      username: "TestPlayer",
    };

    // First connection
    await client.send(
      createMessage(CLIENT_MESSAGE_TYPE.Connect, connectRequest),
    );
    const firstResponse = await waitForMessage(client);
    expect(firstResponse.type).toBe(SERVER_MESSAGE_TYPE.AcceptConnection);

    // Try to connect again
    await client.send(
      createMessage(CLIENT_MESSAGE_TYPE.Connect, connectRequest),
    );
    const secondResponse = await waitForMessage(client);

    expect(secondResponse.type).toBe(SERVER_MESSAGE_TYPE.Error);
    expect(secondResponse.payload.details?.type).toBe("Already connected");
  });
});

describe("Player with Pokemon Tests", () => {
  let client: WSTestClient;

  beforeEach(async () => {
    client = new WSTestClient(WS_URL);
    await client.connect();
  });

  afterEach(async () => {
    if (client.isOpen()) {
      await client.close();
    }
  });

  it("should connect player with username and pokemon list", async () => {
    // Player data
    const playerData = {
      username: "AshKetchum",
      pokemons: [
        {
          id: 25,
          name: "Pikachu",
          hp: 100,
          attack: 55,
          defense: 40,
        },
        {
          id: 6,
          name: "Charizard",
          hp: 150,
          attack: 84,
          defense: 78,
        },
        {
          id: 9,
          name: "Blastoise",
          hp: 140,
          attack: 83,
          defense: 100,
        },
      ],
    };

    // Send connection request
    await client.send(
      createMessage(CLIENT_MESSAGE_TYPE.Connect, {
        username: playerData.username,
        // Note: You may need to extend your Go backend to accept pokemon data
        // This is just an example of what the client might send
      }),
    );

    const response = await waitForMessage(client);

    expect(response.type).toBe(SERVER_MESSAGE_TYPE.AcceptConnection);
    expect(response.payload.username).toBe("AshKetchum");
    expect(response.payload.uuid).toBeDefined();

    // After successful connection, you could send pokemon data separately
    // or include it in a different message type
    console.log("Connected player:", {
      username: response.payload.username,
      uuid: response.payload.uuid,
      pokemons: playerData.pokemons.map((p) => p.name),
    });
  });

  it("should handle multiple pokemon and validate data", async () => {
    const connectRequest = {
      username: "GaryOak",
    };

    await client.send(
      createMessage(CLIENT_MESSAGE_TYPE.Connect, connectRequest),
    );

    const response = await waitForMessage(client);
    expect(response.type).toBe(SERVER_MESSAGE_TYPE.AcceptConnection);

    // Example: You might have a separate message to set up pokemon team
    // This would depend on your actual protocol
    const pokemonTeam = [
      { id: 1, name: "Bulbasaur", hp: 100, attack: 49, defense: 49 },
      { id: 4, name: "Charmander", hp: 95, attack: 52, defense: 43 },
      { id: 7, name: "Squirtle", hp: 94, attack: 48, defense: 65 },
    ];

    expect(pokemonTeam).toHaveLength(3);
    expect(pokemonTeam.every((p) => p.hp > 0)).toBe(true);
  });
});

describe("Connection Lifecycle Tests", () => {
  it("should handle graceful disconnection", async () => {
    const client = new WSTestClient(WS_URL);
    await client.connect();

    const connectRequest = {
      username: "TemporaryPlayer",
    };

    await client.send(
      createMessage(CLIENT_MESSAGE_TYPE.Connect, connectRequest),
    );

    const response = await waitForMessage(client);
    expect(response.type).toBe(SERVER_MESSAGE_TYPE.AcceptConnection);

    // Close connection gracefully
    await client.close();
    expect(client.isOpen()).toBe(false);
  });

  it("should support concurrent connections", async () => {
    const client1 = new WSTestClient(WS_URL);
    const client2 = new WSTestClient(WS_URL);

    await Promise.all([client1.connect(), client2.connect()]);

    await client1.send(
      createMessage(CLIENT_MESSAGE_TYPE.Connect, { username: "Player1" }),
    );
    await client2.send(
      createMessage(CLIENT_MESSAGE_TYPE.Connect, { username: "Player2" }),
    );

    const [response1, response2] = await Promise.all([
      waitForMessage(client1),
      waitForMessage(client2),
    ]);

    expect(response1.payload.username).toBe("Player1");
    expect(response2.payload.username).toBe("Player2");
    expect(response1.payload.uuid).not.toBe(response2.payload.uuid);

    await Promise.all([client1.close(), client2.close()]);
  });
});
