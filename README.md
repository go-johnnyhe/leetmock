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

