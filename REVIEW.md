Review my Go implementation of Netrek and help me identify what's missing to create a minimum viable product. This is a modern rewrite of the classic multiplayer space strategy game with a focus on clean architecture.

The codebase is organized into these key packages:
- cmd/client: client application implementation
- cmd/server: server application implementation
- pkg/physics: Vector math and collision detection
- pkg/entity: Game objects (ships, planets, weapons)
- pkg/event: Event system for game state changes
- pkg/engine: Core game logic and state management
- pkg/network: Client-server communication over TCP
- pkg/config: Game configuration handling

Please focus on:

1. Unimplemented methods and functions - Identify any method stubs or functions that are referenced but not fully implemented
2. Missing struct fields - Are there any critical fields missing from the core data structures?
3. Incomplete interfaces - Are any interfaces missing methods needed for basic functionality?
4. Core gameplay gaps - What essential Netrek mechanics are missing for a playable MVP?
5. Critical paths - Is there a complete path from user input through network to game logic and back?

For each gap identified, suggest a straightforward implementation that prioritizes clarity over performance. I'm looking to get to a functional game first, with optimizations coming later.

Please be specific about which files and functions need attention, and provide simple, readable implementations that complete the minimum necessary functionality.