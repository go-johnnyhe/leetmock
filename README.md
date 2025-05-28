# leetcode-session
Pair-program in terminal using Websockets

Structure
- server setup on docker (certain layer of complexity)
- clients connect to server 
- file notification system (fsnotify)
- diffing algorithm

Server
- websocket (gorilla)
- handle connections
- send diff message

Client
- connect to server
- create file watcher
- send file content
- handle file content
- read diff message


Thoughts
- latency
- resolving conflicts
    - notify the user??
    - choose automatically last write?
- logging for rollback? NoSQL db?
- show diff content, +/- operations like Git?

- Person A "file watch changes <--> ws client <--> ws server" Person B "<--> ws client <--> file watch changes"

Objective:
WebSocket Echo Test

- Person A runs leetcode-mock start (WebSocket server)
- Person B runs leetcode-mock join <ip> (client)
- B sends "hello", A prints it, replies "world", B prints it
Success = working bi-directional connection
