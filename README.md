# 📺 Screenshare Streaming System(Apache Pulsar,Kotlin ,Go,Python)
A cross-platform, end-to-end screen streaming demo using:
- Go: screen grabber/sender
- Kotlin: two relays (TCP→Pulsar and Pulsar→TCP, multi-client support)
- Python: OpenCV real-time viewer
- Apache Pulsar: robust pub/sub message bus

---
## 🏗️ Architecture
[Go Sender]
|
[TCP socket 8082]
|
[Kotlin TCP→Pulsar Relay]
|
[Apache Pulsar topic]
|
[Kotlin Pulsar→TCP Relay]-----(many clients ok)
|
[Python Viewer(s) with OpenCV]
---

## 🗂 Files Overview

| File                       | Language | Purpose                                    |
|----------------------------|----------|--------------------------------------------|
| `go_sender.go`             | Go       | Screen capture & TCP sender (to relay)     |
| `kotlin_tcp_to_pulsar.kt`  | Kotlin   | Receives TCP data, relays to Pulsar topic  |
| `kotlin_pulsar_to_tcp.kt`  | Kotlin   | Pulsar-consuming relay, streams to clients |
| `python_viewer.py`         | Python   | Receives TCP frames, displays with OpenCV  |
---

## 🔧 Pre-requisites

- **Go** installed [golang.org](https://golang.org/)
- **Kotlin** (with Gradle or `kotlinc`), and [Apache Pulsar client JARs](https://search.maven.org/artifact/org.apache.pulsar/pulsar-client)
- **Python 3.7+** with `numpy`, `opencv-python`
- **Apache Pulsar** broker (local or accessible)

---

## ⚡ Setup

1. **Start Apache Pulsar:**  
   Download and run locally:
   ```sh
   bin/pulsar standalone
   
2. Add Requirements (on dev machine):
# Python
pip install numpy opencv-python
# Go
go get github.com/kbinani/screenshot

# Kotlin clients: Add Pulsar client dependency to your build.gradle or classpath.


🚦 Run Order (for LIVE real-time streaming)
-----

Important: Start in this order so nothing is lost!

Kotlin TCP → Pulsar Relay (port 8082):

Receives frames from Go, publishes to Pulsar.

Kotlin Pulsar → TCP Relay (port 8083):

Subscribes to Pulsar, broadcasts to all connecting viewers.

Python OpenCV Viewer:

Connects to relay, displays screen in real time.

Go Sender:

Captures screen, streams to relay.


🖥️ Protocol Details
----------

Frames are sent as [4-byte length in big-endian][1-byte flag][16 bytes rect][JPEG bytes]
flag: 'F' (full frame) or 'D' (dirty rectangle)
rect: x, y, w, h — each as little-endian int32
Each client reconstructs video via keyframes + patches


For multiple viewers make sure to send to all connected socket connections in ViewerSocket.kt : 
---
```
    val server = ServerSocket(8083)
    val clients = CopyOnWriteArrayList<DataOutputStream>()

    // Accept new TCP viewer clients
    Thread {
        while (true) {
            val socket = server.accept()
            println("New client connected: ${socket.inetAddress}")
            val out = DataOutputStream(socket.getOutputStream())
            clients.add(out)
            // Optionally: Add a mechanism to remove clients if they disconnect.
        }
    }.start()

    // Main loop: Every frame from Pulsar is broadcast to all clients
    while (true) {
        val msg = consumer.receive()
        val data = msg.data
        val deadClients = mutableListOf<DataOutputStream>()
        clients.forEach { out ->
            try {
                out.writeInt(data.size)
                out.write(data)

```

🕵️ If you run into issues...
---
Check Pulsar is running and reachable on pulsar://localhost:6650
Try running each step from terminal to see all errors.
If ports 8082/8083 are busy, change them in code and client.
If "Unresolved reference: PulsarClient" in Kotlin, add Pulsar dependency to your Gradle/Maven.


📖 Reference Links
---
## Apache Pulsar https://raghurambtechit.medium.com/apache-pulsar-basics-019d58b61fb8
