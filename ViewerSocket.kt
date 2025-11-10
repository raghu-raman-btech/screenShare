package org.rr

import java.net.ServerSocket
import java.io.DataOutputStream
import org.apache.pulsar.client.api.*

fun main() {
    val pulsarClient = PulsarClient.builder()
        .serviceUrl("pulsar://localhost:6650")
        .build()
    val consumer = pulsarClient.newConsumer()
        .topic("persistent://public/default/screen-share")
        .subscriptionName("viewer")
        .subscribe()

    val server = ServerSocket(8083)
    println("TCP video relay listening on 8083, waiting for viewer client...")
    val socket = server.accept()
    println("Viewer client connected: ${socket.inetAddress}")
    val output = DataOutputStream(socket.getOutputStream())

    while (true) {
        val msg = consumer.receive()
        val data = msg.data
        output.writeInt(data.size)
        output.write(data)
        output.flush()
        consumer.acknowledge(msg)
        println("Pushed frame of ${data.size / 1024.0} KB to client")
    }
}