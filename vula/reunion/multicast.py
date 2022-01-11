"""
This is using two sockets, so that we can listen on a fixed port while using a
randomized port to send (which is then used to track the to/from linkages
between message types).

Each client has its own epoch-like schedule, which will of course never be
perfectly aligned and could in fact be wildly different. When an epoch starts,
the t1 is multicast. When a multicast t1 is received, if we haven't seen that
t1 before, we unicast both our t1 (again) and our t2 to the sender.

This flow is not ideal but it works for the moment. Setting the interval to a
small value, near 2x the cost of a csidh operation, results in it only working
when both clients are started at approximately the same moment (otherwise they
stay out-of-sync across their epochs).

In the course of using this, it has become apparent that the reveal_once mode
might need more work. Logically it should be applied across epochs, in case
Mallory arrived early and Bob doesn't arrive until epoch N+1... but if both
parties keep running the protocol then we'll get spurious alerts about 3rd
parties that aren't really there. 🤔

"""
import socket
import struct
import time

import ifaddr
import click

from reunion.session import ReunionSession, T1, SessionPeer


class UDPListener(object):
    def __init__(self, bind_addr, port=0):
        self.bind_addr = bind_addr
        self.session = None
        self.old_session = None
        self.addr_map = {}
        self.old_session = None
        self.old_addr_map = {}

        self.clear_callback()

        # Create the TX socket
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)

        # Set a timeout so the socket does not block
        # indefinitely when trying to receive data.
        self.sock.settimeout(0.2)

        # Set the time-to-live for messages to 1 so they do not
        # go past the local network segment.
        ttl = struct.pack("b", 1)
        self.sock.setsockopt(socket.IPPROTO_IP, socket.IP_MULTICAST_TTL, ttl)
        self.sock.bind((bind_addr, port))

    def bind_callback(self, callback: object):
        self.callbacks.append(callback)

    def clear_callback(self):
        callbacks = self.callbacks
        self.callbacks = []
        return callbacks

    def callback(self, value):
        for callback in self.callbacks:
            callback(value)

    def bind_multicast(self, multicast_group, port):
        ifaces_by_ip = {
            ip.ip: (a.index, a.name)
            for a in reversed(ifaddr.get_adapters())
            for ip in a.ips
        }

        if_idx, if_name = ifaces_by_ip.get(self.bind_addr, (socket.INADDR_ANY, "any"))

        print(
            "Using %s on iface %s (%r) for multicast to %s:%s"
            % (self.bind_addr, if_name, if_idx, multicast_group, port)
        )

        # Create the RX socket
        self.msock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self.msock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)

        # Bind to the server address
        self.msock.bind(("", port))
        self.msock.settimeout(0.2)

        # Tell the operating system to add the socket to
        # the multicast group on all interfaces.
        group = socket.inet_aton(multicast_group)
        mreq = struct.pack("4sL", group, if_idx)
        self.msock.setsockopt(socket.IPPROTO_IP, socket.IP_ADD_MEMBERSHIP, mreq)

    def send(self, prefix, message, dest):
        payload = prefix + message
        print("Sending", prefix)
        #        print("Sending %s byte message %r to %s:%s" % (len(payload), payload[:5], *dest))
        sent = self.sock.sendto(payload, dest)
        if sent != len(payload):
            print("Tried to send %s bytes but only sent %s!" % (len(payload), sent))

    def poll(self):
        try:
            data, addr = self.sock.recvfrom(1024)
        except socket.timeout:
            return
        return self.process_message(data, addr)

    def poll_multicast(self):
        try:
            data, addr = self.msock.recvfrom(1024)
        except socket.timeout:
            return
        if data.startswith(b"t1"):
            return self.process_message(data, addr)

    def process_message(self, data, addr):

        #        print("Receivd %s byte message %r from %s:%s" % (len(data), data[:5], *addr))
        peer: SessionPeer = self.addr_map.get(addr)
        if not peer:
            peer = self.old_addr_map.get(addr)
            if peer and data[:2] != b"t1":
                print(
                    "!!! received %s message for %r from previous epoch"
                    % (data[:4], peer.t1)
                )

        if data.startswith(b"t1"):
            t1 = T1(data[3:])
            if t1.id in self.session.peers:
                print("ignoring replay of %r from %r" % (t1, addr))
                return
            elif t1.id == self.session.t1.id:
                #                print("ignoring our own %r from %r" % (t1, addr))
                return
            if data[2:3] == b"r":
                print("received t1r %r from %r" % (t1, addr))
            else:
                print("received t1_ %r from %r %r" % (t1, addr, data[:5]))
                self.send(b"t1r", self.session.t1, addr)
            t2 = self.session.process_t1(t1)
            self.send(b"t2", t2, addr)
            self.addr_map[addr] = self.session.peers[t1.id]

        elif data.startswith(b"t2"):

            if not peer:
                print("ignoring t2 message from unknown peer %r" % (addr,))
                return

            t2 = data[2:]
            t3, is_dummy = peer.process_t2(t2)
            if is_dummy:
                print("decryption of t2 from %r failed; sending dummy" % (addr,))
            else:
                print("successful decryption of t2 from %r; sending reveal" % (addr,))
            self.send(b"t3", t3, addr)

        elif data.startswith(b"t3"):

            if not peer:
                print("ignoring t3 message from unknown peer %r" % (addr,))
                return
            t3 = data[2:]
            print("received t3 from %r (%r)" % (peer.t1, addr))
            res = self.process_result(peer.process_t3(t3), addr, peer)

            self.callback(addr)

            print("t3 result:", res)

        else:
            print("%r sent us something weird: %r" % (addr, data[:50]))

    def process_result(self, payload, addr, peer):
        if payload is not None:
            print(
                "Decrypted message from %r on %r:\n   %s\n\n"
                % (peer.t1, addr, payload.decode())
            )

    def new_session(self, passphrase, message):
        print("creating new reunion session... ", flush=True, end="")
        started = time.time()
        self.old_session, self.session = self.session, ReunionSession.create(
            passphrase, message
        )
        print(
            "OK: creating {!r} took {:.1f} seconds)".format(
                self.session.t1, time.time() - started
            )
        )
        self.old_addr_map, self.addr_map = self.addr_map, {}
