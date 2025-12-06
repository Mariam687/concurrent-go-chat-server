# Concurrent go chat server

## Description
A real-time multi-client chat application implemented in **Go**.  
It supports multiple clients connecting to a single server. Clients can send messages that are broadcast to all other clients in real-time.  

Features include:  
- Real-time message broadcasting between clients  
- Join notifications for new users  
- No self-message echo  
- Simple command-line interface  

### This project demonstrates Go concurrency, channels, mutex synchronization, and RPC-based client-server communication.

---

## Demo

![Chat Demo](assets/demo.png)  
*Screenshot of the running GoRPC-Chat showing multiple clients connected and messages broadcasted.*

---


## 💻 How to Run

1.  **Ensure Go is Installed:** Make sure you have Go installed on your system.
2.  **Clone the Repository:**
    ```bash
    git clone https://github.com/Mariam687/concurrent-go-chat-server.git
    cd concurrent-go-chat-server
    ```
3.  **Run the Server:** Execute the `server/server.go` file from the terminal.

    ```bash
    go run server.go
    ```
4.  **Run the Client:** Execute the `client/client.go` file from another terminal.

    ```bash
    go run client.go
    ```






