import socket
import numpy as np
import cv2
import struct

WIDTH, HEIGHT = 1512, 982
CHANNELS = 3

tcp_sock = socket.create_connection(('localhost', 8083))
frame = np.zeros((HEIGHT, WIDTH, CHANNELS), dtype=np.uint8)

def read_n(sock, n):
    buf = b''
    while len(buf) < n:
        chunk = sock.recv(n - len(buf))
        if not chunk:
            raise RuntimeError("Socket closed")
        buf += chunk
    return buf

def parse_rect(b):
    return struct.unpack('<4i', b)  # little-endian int32 x 4

print("Streaming video from TCP... (ESC to quit)")

while True:
    # Read 4-byte length (big-endian)
    lbuf = read_n(tcp_sock, 4)
    size = struct.unpack('>I', lbuf)[0]
    data = read_n(tcp_sock, size)
    if len(data) < 17: continue
    ftype = data[0:1]
    rect = parse_rect(data[1:17])
    img_jpg = data[17:]
    x, y, w, h = rect
    img_patch = cv2.imdecode(np.frombuffer(img_jpg, dtype=np.uint8), cv2.IMREAD_COLOR)
    if img_patch is None:
        print("Failed decode, skipping frame")
    elif ftype == b'F':
        frame = img_patch.copy()
        print("← FULL frame applied")
    elif ftype == b'D':
        frame[y:y+h, x:x+w] = img_patch
        print(f"← Patched DIRTY rect x={x} y={y} w={w} h={h}")
    cv2.imshow("Screen Stream", frame)
    if cv2.waitKey(1) == 27: break

tcp_sock.close()
cv2.destroyAllWindows()