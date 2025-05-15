#!/usr/bin/env python3
"""
A simple mock main that launches the GUI with no real functionality.
It uses a fake DataProvider so that DBus/system calls are bypassed.
"""


# -----------------------------------------------------
# 1. A mock/fake DataProvider that returns dummy data
# -----------------------------------------------------
class MockDataProvider:
    def get_peers(self):
        # Return an empty list or minimal fake peers:
        return [
            {
                "name": "Alice",
                "id": 1,
                "other_names": ["Alicia"],
                "status": "active",
                "endpoint": "203.0.113.1:51820",
                "allowed_ip": "10.0.0.2/32",
                "latest_signature": "2025-05-09T08:15:27Z",
                "latest_handshake": "2025-05-09T08:20:13Z",
                "wg_pubkey": "DtBt1FvN3q2u6Yz9XHwRsT4pVl+AWK3U4vZxbr9CdE0="
            },
            {
                "name": "Bob",
                "id": 2,
                "other_names": ["Bobby", "Robert"],
                "status": "inactive",
                "endpoint": "203.0.113.2:51820",
                "allowed_ip": "10.0.1.0/24",
                "latest_signature": "2025-05-09T07:45:00Z",
                "latest_handshake": "2025-05-09T07:50:22Z",
                "wg_pubkey": "AuZx8VwYkL3PfQ9RjM1tN4uBvCxY2D5ZsH7cLp0JnK4="
            },
            {
                "name": "Carol",
                "id": 3,
                "other_names": [],
                "status": "active",
                "endpoint": "203.0.113.3:51820",
                "allowed_ip": "10.0.2.5/32",
                "latest_signature": "2025-05-09T09:05:11Z",
                "latest_handshake": "2025-05-09T09:10:45Z",
                "wg_pubkey": "Fq9P0nTmU3kR5sV2aXhYeWzZcBnLmOpQrStUvWxYzA1="
            },
            {
                "name": "Dave",
                "id": 4,
                "other_names": ["David"],
                "status": "paused",
                "endpoint": "198.51.100.4:51820",
                "allowed_ip": "10.0.3.0/24",
                "latest_signature": "2025-05-09T06:30:00Z",
                "latest_handshake": "2025-05-09T06:35:30Z",
                "wg_pubkey": "GhIjKlMnOpQrStUvWxYz0123456789+/=="
            },
            {
                "name": "Eve",
                "id": 5,
                "other_names": ["Evie"],
                "status": "active",
                "endpoint": "198.51.100.5:51820",
                "allowed_ip": "10.0.4.42/32",
                "latest_signature": "2025-05-09T10:12:34Z",
                "latest_handshake": "2025-05-09T10:15:00Z",
                "wg_pubkey": "ZxYvUtSrQpOnMlKjIhGfEdCbAa1234567890+=="
            },
            {
                "name": "Frank",
                "id": 6,
                "other_names": [],
                "status": "inactive",
                "endpoint": "198.51.100.6:51820",
                "allowed_ip": "10.0.5.0/24",
                "latest_signature": "2025-05-09T05:00:00Z",
                "latest_handshake": "2025-05-09T05:05:05Z",
                "wg_pubkey": "MnOpQrStUvWxYzZaBcDeFgHiJkLmNoPqRsTuVwXyZ="
            },
            {
                "name": "Grace",
                "id": 7,
                "other_names": ["Gracie"],
                "status": "active",
                "endpoint": "203.0.113.7:51820",
                "allowed_ip": "10.0.6.100/32",
                "latest_signature": "2025-05-09T11:11:11Z",
                "latest_handshake": "2025-05-09T11:15:15Z",
                "wg_pubkey": "UvWxYzZaBcDeFgHiJkLmNoPqRsTuVwXyZ0123456="
            },
            {
                "name": "Heidi",
                "id": 8,
                "other_names": [],
                "status": "paused",
                "endpoint": "203.0.113.8:51820",
                "allowed_ip": "10.0.7.0/24",
                "latest_signature": "2025-05-09T04:44:44Z",
                "latest_handshake": "2025-05-09T04:50:50Z",
                "wg_pubkey": "BcDeFgHiJkLmNoPqRsTuVwXyZ0123456UvWxYzZa="
            },
            {
                "name": "Ivan",
                "id": 9,
                "other_names": ["Iván"],
                "status": "active",
                "endpoint": "203.0.113.9:51820",
                "allowed_ip": "10.0.8.8/32",
                "latest_signature": "2025-05-09T03:33:33Z",
                "latest_handshake": "2025-05-09T03:35:35Z",
                "wg_pubkey": "FgHiJkLmNoPqRsTuVwXyZ0123456UvWxYzZaBcDe="
            },
            {
                "name": "Judy",
                "id": 10,
                "other_names": [],
                "status": "inactive",
                "endpoint": "203.0.113.10:51820",
                "allowed_ip": "10.0.9.0/24",
                "latest_signature": "2025-05-09T02:22:22Z",
                "latest_handshake": "2025-05-09T02:25:25Z",
                "wg_pubkey": "LmNoPqRsTuVwXyZ0123456UvWxYzZaBcDeFgHiJk="
            },
        ]

    def get_prefs(self):
        # Return empty or fake preferences
        return {
            "pin_new_peers": False,
            "auto_repair": True,
            "subnets_allowed": ["192.168.0.0/24"],
            "subnets_forbidden": [],
            "iface_prefix_allowed": ["eth", "wl"],
            "accept_nonlocal": False,
            "local_domains": ["local."],
            "ephemeral_mode": False,
            "accept_default_route": True,
            "record_events": False,
            "expire_time": 3600,
            "overwrite_unpinned": True,
        }

    def get_status(self):
        # Pretend all services are "inactive" or "active" to see the UI
        return {
            "publish": "inactive",
            "discover": "inactive",
            "organize": "inactive",
        }

    def our_latest_descriptors(self):
        # Return an empty JSON object or minimal structure
        return "{}"

    # The “write” type methods can just be no-ops:
    def delete_peer(self, peer_vk):
        pass

    def rename_peer(self, peer_vk, name):
        pass

    def pin_and_verify(self, peer_vk, peer_name):
        pass

    def add_peer(self, peer_vk, ip):
        pass

    def set_pref(self, pref, value):
        return ""

    def add_pref(self, pref, value):
        return ""

    def remove_pref(self, pref, value):
        return ""


# -----------------------------------------------------
# 2. Patch or override the actual DataProvider in code
# -----------------------------------------------------
def patch_dataprovider():
    """Replace the real DataProvider with our mock so that GUI calls the mock methods."""
    import vula.frontend.dataprovider as dp
    import vula.frontend

    dp.DataProvider = MockDataProvider
    vula.frontend.DataProvider = MockDataProvider


# -----------------------------------------------------
# 3. Patch or override the actual constants in code
# -----------------------------------------------------
def patch_constants():
    import vula.frontend.constants as c
    c.IMAGE_BASE_PATH = "/IdeaProjects/vula/misc/images/"


# -----------------------------------------------------
# 4. Run the GUI code directly
# -----------------------------------------------------
def main():
    # 3.1 patch the data provider
    patch_dataprovider()

    # 3.2 patch the constants
    patch_constants()

    # 3.3 import the GUI “main” after patching
    from vula.frontend.ui import App

    # 3.4 create and launch the Tkinter application
    root = App()
    root.title("Vula-GUI (Mocked)")
    root.mainloop()


if __name__ == "__main__":
    main()
