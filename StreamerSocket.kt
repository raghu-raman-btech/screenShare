package org.rr

import java.net.ServerSocket
import java.io.DataInputStream
import org.apache.pulsar.client.api.*

fun main() {
    val pulsarClient = PulsarClient.builder()
        .serviceUrl("pulsar://localhost:6650")
        .build()
    val producer = pulsarClient.newProducer()
        .topic("persistent://public/default/screen-share")
        .create()

    val server = ServerSocket(8082)
    println("TCP server listening on 8082 for screen sender...")
    while (true) {
        val socket = server.accept()
        println("Client connected: ${socket.inetAddress}")
        Thread {
            val input = DataInputStream(socket.getInputStream())
            try {
                while (true) {
                    val size = input.readInt()
                    val buf = ByteArray(size)
                    var read = 0
                    while (read < size) {
                        val r = input.read(buf, read, size - read)
                        if (r == -1) throw RuntimeException("Disconnected mid-frame")
                        read += r
                    }
                    producer.send(buf)
                    println("Pushed frame of ${buf.size / 1024.0} KB to Pulsar")
                }
            } catch (e: Exception) {
                println("Connection closed: ${e.message}")
            } finally {
                input.close(); socket.close()
            }
        }.start()
    }
}