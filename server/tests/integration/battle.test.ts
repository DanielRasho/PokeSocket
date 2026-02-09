import { describe, test, expect } from "vitest";
import {
  CLIENT_MESSAGE_TYPE,
  CONNECT_REQUEST,
  CONNECT_SCHEMA,
  MATCH_REQUEST,
  MATCH_FOUND_SCHEMA,
  ATTACK_REQUEST,
  ATTACK_RESPONSE_SCHEMA,
  ERROR_SCHEMA,
  SERVER_MESSAGE_TYPE,
  validateResponse,
  waitForMessage,
  WS_URL,
  WSTestClient,
  type Message,
} from "../helpers";

// Helper function to setup a fresh battle for each test
async function setupBattle() {
  const client1 = new WSTestClient(WS_URL);
  const client2 = new WSTestClient(WS_URL);

  await Promise.all([client1.connect(), client2.connect()]);

  // Connect both players
  await client1.send(CONNECT_REQUEST("Player1", [1, 2, 3]));
  await client2.send(CONNECT_REQUEST("Player2", [1, 5, 6]));

  await Promise.all([waitForMessage(client1), waitForMessage(client2)]);

  // Enter matchmaking
  await client1.send(MATCH_REQUEST());
  await waitForMessage(client1); // Queue joined

  await client2.send(MATCH_REQUEST());

  const [match1, match2] = await Promise.all([
    waitForMessage(client1),
    waitForMessage(client2),
  ]);

  expect(match1.type).toBe(SERVER_MESSAGE_TYPE.MatchFound);
  expect(match2.type).toBe(SERVER_MESSAGE_TYPE.MatchFound);

  const battleId = match1.payload.battle_id;

  return { client1, client2, battleId };
}

describe("Battle System", () => {

  test("should allow player1 to attack on turn 1", async () => {
    const { client1, client2, battleId } = await setupBattle();

    // Player1 attacks (turn 1 = player1's turn)
    await client1.send(ATTACK_REQUEST(battleId, 1));

    const [response1, response2] = await Promise.all([
      waitForMessage(client1),
      waitForMessage(client2),
    ]);

    // Both players should receive attack response
    expect(response1.type).toBe(SERVER_MESSAGE_TYPE.Attack);
    expect(response2.type).toBe(SERVER_MESSAGE_TYPE.Attack);

    validateResponse(response1.payload, ATTACK_RESPONSE_SCHEMA);
    validateResponse(response2.payload, ATTACK_RESPONSE_SCHEMA);
    
    console.log(JSON.stringify(response1.payload, null, "\t"))

    // Verify damage was dealt
    expect(response1.payload.message).toContain("damage");
    expect(response2.payload.message).toContain("damage");

    await Promise.all([client1.close(), client2.close()]);
  });

  test("should reject attack from wrong player on turn 1", async () => {
    const { client1, client2, battleId } = await setupBattle();

    // Player2 tries to attack on turn 1 (should fail - it's player1's turn)
    await client2.send(ATTACK_REQUEST(battleId, 1));

    const errorResponse = await waitForMessage(client2);

    expect(errorResponse.type).toBe(SERVER_MESSAGE_TYPE.Error);
    validateResponse(errorResponse.payload, ERROR_SCHEMA);
    expect(errorResponse.payload.msg).toContain("Bad request");
    expect(errorResponse.payload.details.error).toContain("not your turn");

    await Promise.all([client1.close(), client2.close()]);
  });

  test("should alternate turns between players", async () => {
    const { client1, client2, battleId } = await setupBattle();

    // Turn 1: Player1 attacks
    await client1.send(ATTACK_REQUEST(battleId, 1));
    const [turn1_p1, turn1_p2] = await Promise.all([
      waitForMessage(client1),
      waitForMessage(client2),
    ]);
    expect(turn1_p1.type).toBe(SERVER_MESSAGE_TYPE.Attack);
    expect(turn1_p2.type).toBe(SERVER_MESSAGE_TYPE.Attack);

    // Turn 2: Player2 attacks
    await client2.send(ATTACK_REQUEST(battleId, 1));
    const [turn2_p1, turn2_p2] = await Promise.all([
      waitForMessage(client1),
      waitForMessage(client2),
    ]);
    expect(turn2_p1.type).toBe(SERVER_MESSAGE_TYPE.Attack);
    expect(turn2_p2.type).toBe(SERVER_MESSAGE_TYPE.Attack);

    // Turn 3: Player1 attacks again
    await client1.send(ATTACK_REQUEST(battleId, 1));
    const [turn3_p1, turn3_p2] = await Promise.all([
      waitForMessage(client1),
      waitForMessage(client2),
    ]);
    expect(turn3_p1.type).toBe(SERVER_MESSAGE_TYPE.Attack);
    expect(turn3_p2.type).toBe(SERVER_MESSAGE_TYPE.Attack);

    // Turn 4: Player2 attacks again
    await client2.send(ATTACK_REQUEST(battleId, 1));
    const [turn4_p1, turn4_p2] = await Promise.all([
      waitForMessage(client1),
      waitForMessage(client2),
    ]);
    expect(turn4_p1.type).toBe(SERVER_MESSAGE_TYPE.Attack);
    expect(turn4_p2.type).toBe(SERVER_MESSAGE_TYPE.Attack);

    await Promise.all([client1.close(), client2.close()]);
  });

  test("should damage defender's active pokemon", async () => {
    const { client1, client2, battleId } = await setupBattle();

    // Get initial HP of player2's first pokemon (active by default)
    await client1.send(ATTACK_REQUEST(battleId, 1));
    const [response1, response2] = await Promise.all([
      waitForMessage(client1),
      waitForMessage(client2),
    ]);

    const initialHP =
      response2.payload.your_info.team.find((p) => p.position === 1)
        ?.current_hp || 0;

    console.log(JSON.stringify(response2.payload, null, "\t"))

    // Attack again on player1's next turn
    await client2.send(ATTACK_REQUEST(battleId, 1)); // Player2 turn 2
    await Promise.all([waitForMessage(client1), waitForMessage(client2)]);

    await client1.send(ATTACK_REQUEST(battleId, 1)); // Player1 turn 3
    const [response3, response4] = await Promise.all([
      waitForMessage(client1),
      waitForMessage(client2),
    ]);

    const newHP =
      response4.payload.your_info.team.find((p) => p.position === 1)
        ?.current_hp || 0;

    console.log(JSON.stringify(response4.payload, null, "\t"))

    // HP should have decreased further
    expect(newHP).toBeLessThan(initialHP);

    await Promise.all([client1.close(), client2.close()]);
  });

  test("should auto-switch to next pokemon when active pokemon faints", async () => {
    const { client1, client2, battleId } = await setupBattle();

    // Attack repeatedly until first pokemon faints
    let lastResponse1: Message;
    let lastResponse2: Message;
    let turn = 1;
    const maxTurns = 20; // Safety limit

    while (turn <= maxTurns) {
      if (turn % 2 === 1) {
        // Player1's turn
        await client1.send(ATTACK_REQUEST(battleId, 1));
      } else {
        // Player2's turn
        await client2.send(ATTACK_REQUEST(battleId, 1));
      }

      [lastResponse1, lastResponse2] = await Promise.all([
        waitForMessage(client1),
        waitForMessage(client2),
      ]);

      // Check if first pokemon fainted
      const player2Pokemon1 = lastResponse2.payload.your_info.team.find(
        (p) => p.position === 1
      );

      if (player2Pokemon1 && player2Pokemon1.is_fainted) {
        // Pokemon fainted! Check if auto-switch happened
        expect(player2Pokemon1.current_hp).toBe(0);
        expect(player2Pokemon1.is_fainted).toBe(true);

        // Should have switched to next available pokemon (position 2)
        const player2Pokemon2 = lastResponse2.payload.your_info.team.find(
          (p) => p.position === 2
        );
        expect(player2Pokemon2).toBeDefined();
        expect(player2Pokemon2.current_hp).toBeGreaterThan(0);

        break;
      }

      turn++;
    }

    expect(turn).toBeLessThan(maxTurns); // Ensure we didn't timeout

    await Promise.all([client1.close(), client2.close()]);
  });

  test("should end battle when all opponent pokemon are fainted", async () => {
    const { client1, client2, battleId } = await setupBattle();

    let turn = 1;
    const maxTurns = 50; // Safety limit
    let battleEnded = false;

    while (turn <= maxTurns && !battleEnded) {
      if (turn % 2 === 1) {
        // Player1's turn
        await client1.send(ATTACK_REQUEST(battleId, 1));
      } else {
        // Player2's turn
        await client2.send(ATTACK_REQUEST(battleId, 1));
      }

      const [response1, response2] = await Promise.all([
        waitForMessage(client1),
        waitForMessage(client2),
      ]);

      if (response1.payload.battle_ended) {
        battleEnded = true;

        // Verify battle ended properly
        expect(response1.payload.battle_ended).toBe(true);
        expect(response2.payload.battle_ended).toBe(true);

        // Winner should be set
        expect(response1.payload.winner).toBeDefined();
        expect(response2.payload.winner).toBeDefined();
        expect(response1.payload.winner).toBe(response2.payload.winner);

        // All pokemon of one team should be fainted
        const player1AllFainted = response1.payload.your_info.team.every(
          (p) => p.is_fainted
        );
        const player2AllFainted = response1.payload.opponent_info.team.every(
          (p) => p.is_fainted
        );

        expect(player1AllFainted || player2AllFainted).toBe(true);

        // Winner should be the player whose pokemon are NOT all fainted
        if (player1AllFainted) {
          expect(response1.payload.winner).toBe(
            response1.payload.opponent_info.player_id
          );
        } else if (player2AllFainted) {
          expect(response1.payload.winner).toBe(
            response1.payload.your_info.player_id
          );
        }
      }

      turn++;
    }

    expect(battleEnded).toBe(true);

    await Promise.all([client1.close(), client2.close()]);
  });

  test("should handle rapid attack requests from same player", async () => {
    const { client1, client2, battleId } = await setupBattle();

    // Player1 sends multiple attacks in quick succession
    await client1.send(ATTACK_REQUEST(battleId, 1));
    await client1.send(ATTACK_REQUEST(battleId, 1)); // Should fail - not their turn after first
    await client1.send(ATTACK_REQUEST(battleId, 1)); // Should fail - not their turn

    // First attack should succeed
    const [success1_p1, success1_p2] = await Promise.all([
      waitForMessage(client1),
      waitForMessage(client2),
    ]);
    expect(success1_p1.type).toBe(SERVER_MESSAGE_TYPE.Attack);
    expect(success1_p2.type).toBe(SERVER_MESSAGE_TYPE.Attack);

    // Next two should be errors (not player1's turn anymore)
    const error1 = await waitForMessage(client1);
    expect(error1.type).toBe(SERVER_MESSAGE_TYPE.Error);
    expect(error1.payload.details.error).toContain("not your turn");

    const error2 = await waitForMessage(client1);
    expect(error2.type).toBe(SERVER_MESSAGE_TYPE.Error);
    expect(error2.payload.details.error).toContain("not your turn");

    await Promise.all([client1.close(), client2.close()]);
  });

  test("should synchronize battle state between both players", async () => {
    const { client1, client2, battleId } = await setupBattle();

    // Execute several turns
    for (let i = 0; i < 4; i++) {
      if (i % 2 === 0) {
        await client1.send(ATTACK_REQUEST(battleId, 1));
      } else {
        await client2.send(ATTACK_REQUEST(battleId, 1));
      }

      const [response1, response2] = await Promise.all([
        waitForMessage(client1),
        waitForMessage(client2),
      ]);

      // Both players should see the same battle state (from their perspective)
      expect(response1.payload.battle_id).toBe(response2.payload.battle_id);

      // Player1's your_info should match Player2's opponent_info
      expect(response1.payload.your_info.player_id).toBe(
        response2.payload.opponent_info.player_id
      );
      expect(response1.payload.opponent_info.player_id).toBe(
        response2.payload.your_info.player_id
      );

      // Team states should match
      expect(response1.payload.your_info.team).toEqual(
        response2.payload.opponent_info.team
      );
      expect(response1.payload.opponent_info.team).toEqual(
        response2.payload.your_info.team
      );
    }

    await Promise.all([client1.close(), client2.close()]);
  });
});
